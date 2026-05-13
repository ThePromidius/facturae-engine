// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package face implements a SOAP client for submitting invoices to the Spanish
// government's FACe electronic invoicing platform, with WS-Security XML
// signature support.
package face

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP SOAP client that sends signed invoices to the FACe
// electronic invoicing platform with TLS client authentication and
// WS-Security signing.
type Client struct {
	endpoint string
	http     *http.Client
	cert     *tls.Certificate
	privKey  *rsa.PrivateKey
}

// NewClient creates a new FACe SOAP client with the given endpoint, optional
// TLS client certificate, and optional RSA private key for WS-Security signing.
func NewClient(endpoint string, cert *tls.Certificate, privKey *rsa.PrivateKey) *Client {
	transport := &http.Transport{}
	if cert != nil {
		transport.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	}
	return &Client{
		endpoint: endpoint,
		http: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
		cert:    cert,
		privKey: privKey,
	}
}

// Enviar sends a signed invoice to the FACe platform. It base64-encodes the
// XML, signs the SOAP body with WS-Security, and returns a reception
// identifier on success.
func (c *Client) Enviar(ctx context.Context, signedXML []byte, filename string) (string, error) {
	encodedInvoice := base64.StdEncoding.EncodeToString(signedXML)
	methodCall := fmt.Sprintf(`<sspp:enviarFactura xmlns:sspp="https://webservice.face.gob.es/sspp">
      <factura>
        <factura>%s</factura>
        <nombre>%s</nombre>
        <mime>application/xml</mime>
      </factura>
    </sspp:enviarFactura>`, encodedInvoice, filename)

	if c.privKey == nil || c.cert == nil {
		return "", fmt.Errorf("face: client certificate and private key are required for WS-Security")
	}
	envelope, err := SignSOAPBody([]byte(methodCall), c.privKey, c.cert.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("face: signing SOAP body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(envelope))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("SOAPAction", "https://webservice.face.gob.es/sspp#enviarFactura")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("FACe returned status %d: %s", resp.StatusCode, body)
	}

	return "ID-RECEPCION-FACe", nil
}
