// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package face

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// SignSOAPBody wraps bodyContent in a SOAP envelope with a WS-Security header
// containing an X.509 BinarySecurityToken and an RSA-SHA256 signature over the
// body. It returns the full signed SOAP envelope as a byte slice.
func SignSOAPBody(bodyContent []byte, privKey *rsa.PrivateKey, certDER []byte) ([]byte, error) {
	bodyID := "id-body-" + fmt.Sprintf("%d", time.Now().UnixNano())
	tokenID := "id-token-" + fmt.Sprintf("%d", time.Now().UnixNano())

	taggedBody := fmt.Sprintf(`<soapenv:Body xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd" wsu:Id="%s">%s</soapenv:Body>`,
		bodyID, string(bodyContent))

	h := sha256.Sum256([]byte(taggedBody))
	digestB64 := base64.StdEncoding.EncodeToString(h[:])

	signedInfo := fmt.Sprintf(`<ds:SignedInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
      <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
      <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
      <ds:Reference URI="#%s">
        <ds:Transforms>
          <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
        </ds:Transforms>
        <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
        <ds:DigestValue>%s</ds:DigestValue>
      </ds:Reference>
    </ds:SignedInfo>`, bodyID, digestB64)

	sigHash := sha256.Sum256([]byte(signedInfo))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, sigHash[:])
	if err != nil {
		return nil, err
	}
	sigB64 := base64.StdEncoding.EncodeToString(sigBytes)

	certB64 := base64.StdEncoding.EncodeToString(certDER)
	securityHeader := fmt.Sprintf(`
  <soapenv:Header>
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
      <wsse:BinarySecurityToken EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary" ValueType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3" wsu:Id="%s">%s</wsse:BinarySecurityToken>
      <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        %s
        <ds:SignatureValue>%s</ds:SignatureValue>
        <ds:KeyInfo>
          <wsse:SecurityTokenReference>
            <wsse:Reference URI="#%s" ValueType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3"/>
          </wsse:SecurityTokenReference>
        </ds:KeyInfo>
      </ds:Signature>
    </wsse:Security>
  </soapenv:Header>`, tokenID, certB64, signedInfo, sigB64, tokenID)

	fullEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  %s
  %s
</soapenv:Envelope>`, securityHeader, taggedBody)

	return []byte(fullEnvelope), nil
}
