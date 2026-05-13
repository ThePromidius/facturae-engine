// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package signing_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/signing"
)

func TestMockSigner_AppendsMockComment(t *testing.T) {
	s := signing.MockSigner{}
	xml := []byte(`<?xml version="1.0"?><fe:Facturae>content</fe:Facturae>`)
	out, err := s.Sign(xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "MOCK SIGNATURE") {
		t.Error("expected mock signature comment in output")
	}
}

func TestMockSigner_ContainsSHA256(t *testing.T) {
	s := signing.MockSigner{}
	xml := []byte(`<?xml version="1.0"?><fe:Facturae>hello</fe:Facturae>`)
	out, err := s.Sign(xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := sha256.Sum256(xml)
	hashHex := make([]byte, 64)
	hexEncode(hashHex, h[:])
	if !strings.Contains(string(out), string(hashHex)) {
		t.Error("expected SHA-256 of input in mock signature comment")
	}
}

func TestMockSigner_OutputContainsOriginalXML(t *testing.T) {
	s := signing.MockSigner{}
	input := []byte(`<?xml version="1.0"?><fe:Facturae>payload</fe:Facturae>`)
	out, err := s.Sign(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "payload") {
		t.Error("signed output should contain original XML payload")
	}
}

func TestMockSigner_Algorithm(t *testing.T) {
	s := signing.MockSigner{}
	if !strings.Contains(s.Algorithm(), "mock") {
		t.Error("MockSigner.Algorithm() should mention 'mock'")
	}
}

func generateTestKey(t *testing.T) (keyPEM, certPEM []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test Signer"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating certificate: %v", err)
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshalling key: %v", err)
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return
}

func TestP12Signer_SignsAndContainsSignatureElement(t *testing.T) {
	keyPEM, certPEM := generateTestKey(t)
	s, err := signing.NewP12SignerFromPEM(keyPEM, certPEM)
	if err != nil {
		t.Fatalf("NewP12SignerFromPEM: %v", err)
	}

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<fe:Facturae xmlns:fe="http://facturae.gob.es" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <FileHeader><SchemaVersion>3.2.2</SchemaVersion></FileHeader>
</fe:Facturae>`)

	out, err := s.Sign(xml)
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	if !strings.Contains(string(out), "<ds:Signature") {
		t.Error("signed output should contain <ds:Signature> element")
	}
	if !strings.Contains(string(out), "<ds:SignatureValue>") {
		t.Error("signed output should contain <ds:SignatureValue>")
	}
	if !strings.Contains(string(out), "<ds:X509Certificate>") {
		t.Error("signed output should contain <ds:X509Certificate>")
	}
}

func TestP12Signer_ClosingTagPreserved(t *testing.T) {
	keyPEM, certPEM := generateTestKey(t)
	s, err := signing.NewP12SignerFromPEM(keyPEM, certPEM)
	if err != nil {
		t.Fatalf("NewP12SignerFromPEM: %v", err)
	}

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<fe:Facturae xmlns:fe="http://facturae.gob.es" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
</fe:Facturae>`)

	out, err := s.Sign(xml)
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	if !strings.HasSuffix(strings.TrimSpace(string(out)), "</fe:Facturae>") {
		t.Error("signed XML should still end with </fe:Facturae>")
	}
}

func TestP12Signer_Algorithm(t *testing.T) {
	keyPEM, certPEM := generateTestKey(t)
	s, err := signing.NewP12SignerFromPEM(keyPEM, certPEM)
	if err != nil {
		t.Fatalf("NewP12SignerFromPEM: %v", err)
	}
	if !strings.Contains(s.Algorithm(), "RSA-SHA256") {
		t.Error("P12Signer.Algorithm() should mention RSA-SHA256")
	}
}

func TestNewP12SignerFromPEM_InvalidKey(t *testing.T) {
	badKey := []byte("not a pem key")
	_, certPEM := generateTestKey(t)
	_, err := signing.NewP12SignerFromPEM(badKey, certPEM)
	if err == nil {
		t.Error("expected error for invalid key PEM")
	}
}

func TestNewP12SignerFromPEM_InvalidCert(t *testing.T) {
	keyPEM, _ := generateTestKey(t)
	_, err := signing.NewP12SignerFromPEM(keyPEM, []byte("not a cert"))
	if err == nil {
		t.Error("expected error for invalid certificate PEM")
	}
}

const hexTable = "0123456789abcdef"

func hexEncode(dst []byte, src []byte) {
	for i, b := range src {
		dst[i*2] = hexTable[b>>4]
		dst[i*2+1] = hexTable[b&0xf]
	}
}
