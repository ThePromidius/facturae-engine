// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package signing

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// P12Signer implements the Signer interface using an RSA private key and X.509 certificate
// to produce XAdES-BES enveloped signatures for Facturae XML documents.
type P12Signer struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// NewP12SignerFromPEM creates a P12Signer from PEM-encoded RSA private key and X.509 certificate data.
// The key may be in PKCS#1 or PKCS#8 format; only RSA keys are supported.
func NewP12SignerFromPEM(keyPEM, certPEM []byte) (*P12Signer, error) {
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, errors.New("signing: failed to decode PEM key block")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("signing: parsing private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("signing: only RSA keys are supported")
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, errors.New("signing: failed to decode PEM certificate block")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("signing: parsing certificate: %w", err)
	}

	return &P12Signer{privateKey: rsaKey, certificate: cert}, nil
}

// Algorithm returns the signing algorithm identifier string for this signer.
func (s *P12Signer) Algorithm() string { return "RSA-SHA256 / XAdES-BES (simplified C14N)" }

// Sign produces an XAdES-BES enveloped signature for the given Facturae XML data.
// It computes a SHA-256 digest, builds the ds:SignedInfo block, signs it with RSA-PKCS1v15,
// and inserts the ds:Signature element before the closing root tag.
func (s *P12Signer) Sign(xmlData []byte) ([]byte, error) {
	digest := sha256.Sum256(xmlData)
	digestB64 := base64.StdEncoding.EncodeToString(digest[:])

	signedInfo := buildSignedInfo(digestB64)

	sigHash := sha256.Sum256([]byte(signedInfo))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, sigHash[:])
	if err != nil {
		return nil, fmt.Errorf("signing: RSA sign: %w", err)
	}
	sigB64 := base64.StdEncoding.EncodeToString(sigBytes)

	certDER := s.certificate.Raw
	certB64 := base64.StdEncoding.EncodeToString(certDER)

	sigBlock := buildSignatureBlock(signedInfo, sigB64, certB64)

	closing := "</fe:Facturae>"
	idx := strings.LastIndex(string(xmlData), closing)
	if idx < 0 {
		return append(xmlData, []byte("\n"+sigBlock)...), nil
	}

	var out []byte
	out = append(out, xmlData[:idx]...)
	out = append(out, []byte("\n"+sigBlock+"\n")...)
	out = append(out, xmlData[idx:]...)
	return out, nil
}

// buildSignedInfo constructs the ds:SignedInfo XML block with C14N canonicalization,
// RSA-SHA256 signature method, enveloped signature transform, and SHA-256 digest.
func buildSignedInfo(digestB64 string) string {
	return fmt.Sprintf(`<ds:SignedInfo>
  <ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
  <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
  <ds:Reference URI="">
    <ds:Transforms>
      <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
    </ds:Transforms>
    <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
    <ds:DigestValue>%s</ds:DigestValue>
  </ds:Reference>
</ds:SignedInfo>`, digestB64)
}

// buildSignatureBlock constructs the full ds:Signature XML block including the SignedInfo,
// SignatureValue, and KeyInfo containing the X.509 certificate.
func buildSignatureBlock(signedInfo, sigB64, certB64 string) string {
	return fmt.Sprintf(`<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
%s
<ds:SignatureValue>%s</ds:SignatureValue>
<ds:KeyInfo>
  <ds:X509Data>
    <ds:X509Certificate>%s</ds:X509Certificate>
  </ds:X509Data>
</ds:KeyInfo>
</ds:Signature>`, signedInfo, sigB64, certB64)
}
