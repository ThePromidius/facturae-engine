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
	"time"
)

// P12Signer implements the Signer interface using an RSA private key and X.509 certificate
// to produce XAdES-BES enveloped signatures for Facturae XML documents.
type P12Signer struct {
	privateKey  *rsa.PrivateKey
	certificate *x509.Certificate
}

// NewP12SignerFromPEM creates a P12Signer from PEM-encoded RSA private key and X.509 certificate data.
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

// Algorithm returns the signing algorithm name for identification.
func (s *P12Signer) Algorithm() string { return "RSA-SHA256 / XAdES-BES (ETSI EN 319 132)" }

// Sign produces an XAdES-BES enveloped signature.
func (s *P12Signer) Sign(xmlData []byte) ([]byte, error) {
	// 1. Compute Document Digest
	digest := sha256.Sum256(xmlData)
	digestB64 := base64.StdEncoding.EncodeToString(digest[:])

	// 2. Compute Certificate Digest (Mandatory for XAdES-BES)
	certDigest := sha256.Sum256(s.certificate.Raw)
	certDigestB64 := base64.StdEncoding.EncodeToString(certDigest[:])
	
	now := time.Now().UTC().Format(time.RFC3339)
	signatureID := fmt.Sprintf("Signature-%d", time.Now().Unix())

	// 3. Build SignedInfo
	signedInfo := buildSignedInfo(digestB64, signatureID)

	// 4. Build QualifyingProperties (The "X" in XAdES)
	qualifyingProps := buildQualifyingProperties(signatureID, now, certDigestB64)

	// 5. Sign the SignedInfo
	sigHash := sha256.Sum256([]byte(signedInfo))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, sigHash[:])
	if err != nil {
		return nil, fmt.Errorf("signing: RSA sign: %w", err)
	}
	sigB64 := base64.StdEncoding.EncodeToString(sigBytes)

	certB64 := base64.StdEncoding.EncodeToString(s.certificate.Raw)

	// 6. Assemble Full Signature Block
	sigBlock := assembleXAdESBlock(signatureID, signedInfo, sigB64, certB64, qualifyingProps)

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

func buildSignedInfo(digestB64, sigID string) string {
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
  <ds:Reference Type="http://uri.etsi.org/01903#SignedProperties" URI="#%s-SignedProperties">
    <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
    <ds:DigestValue><!-- Placeholder for SignedProperties Digest --></ds:DigestValue>
  </ds:Reference>
</ds:SignedInfo>`, digestB64, sigID)
}

func buildQualifyingProperties(sigID, timestamp, certDigestB64 string) string {
	return fmt.Sprintf(`<xades:QualifyingProperties Target="#%s" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#">
  <xades:SignedProperties Id="%s-SignedProperties">
    <xades:SignedGeneralProperties>
      <xades:SigningTime>%s</xades:SigningTime>
    </xades:SignedGeneralProperties>
    <xades:SignedDataObjectProperties>
      <xades:DataObjectFormat ObjectReference="">
        <xades:Description>Factura electrónica</xades:Description>
        <xades:MimeType>text/xml</xades:MimeType>
      </xades:DataObjectFormat>
    </xades:SignedDataObjectProperties>
    <xades:SigningCertificate>
      <xades:Cert>
        <xades:CertDigest>
          <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
          <ds:DigestValue>%s</ds:DigestValue>
        </xades:CertDigest>
        <xades:IssuerSerial>
           <!-- Issuer info omitted for brevity in POC -->
        </xades:IssuerSerial>
      </xades:Cert>
    </xades:SigningCertificate>
  </xades:SignedProperties>
</xades:QualifyingProperties>`, sigID, sigID, timestamp, certDigestB64)
}

func assembleXAdESBlock(sigID, signedInfo, sigB64, certB64, qualifyingProps string) string {
	return fmt.Sprintf(`<ds:Signature Id="%s" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
%s
<ds:SignatureValue>%s</ds:SignatureValue>
<ds:KeyInfo>
  <ds:X509Data>
    <ds:X509Certificate>%s</ds:X509Certificate>
  </ds:X509Data>
</ds:KeyInfo>
<ds:Object>
%s
</ds:Object>
</ds:Signature>`, sigID, signedInfo, sigB64, certB64, qualifyingProps)
}
