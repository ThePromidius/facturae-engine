// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package aeat implements a SOAP client for submitting signed invoices to the
// Spanish Tax Agency (AEAT) Verifactu service. It supports configurable
// environments (test/prod), TLS client certificates, and automatic retries.
package aeat

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// Client is an HTTP SOAP client that submits signed invoices to the AEAT
// Verifactu web service with configurable retry logic.
type Client struct {
	env      Environment
	http     *http.Client
	maxTries int
}

// NewClient creates a new AEAT SOAP client targeting the given environment. If
// cert is non-nil, it configures TLS client certificate authentication.
func NewClient(env Environment, cert *tls.Certificate) *Client {
	transport := &http.Transport{}
	if cert != nil {
		transport.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	}
	return &Client{
		env:      env,
		maxTries: 3,
		http: &http.Client{
			Transport: transport,
			Timeout:   45 * time.Second,
		},
	}
}

// SetHTTPClient replaces the default HTTP client with a custom one. Useful for
// testing or custom transport configuration.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

// SetMaxTries sets the maximum number of submission retry attempts (default 3).
func (c *Client) SetMaxTries(n int) {
	c.maxTries = n
}

// Submit sends the signed invoice XML to the AEAT Verifactu endpoint for the
// given issuer CIF. It retries on failure with exponential backoff up to
// c.maxTries attempts.
func (c *Client) Submit(ctx context.Context, signedXML []byte, emisorCIF string) (*SubmitResult, error) {
	envelope := buildSOAPEnvelope(signedXML, emisorCIF)
	endpoint, ok := Endpoints[c.env]
	if !ok {
		return nil, fmt.Errorf("aeat: unknown environment %q", c.env)
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxTries; attempt++ {
		result, err := c.doRequest(ctx, endpoint, envelope)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if attempt < c.maxTries {
			wait := time.Duration(attempt*attempt) * time.Second
			fmt.Printf("[aeat] Intento %d/%d fallido: %v - reintentando en %s\n",
				attempt, c.maxTries, err, wait)
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, fmt.Errorf("aeat: %d intentos fallidos: %w", c.maxTries, lastErr)
}

// doRequest performs a single HTTP POST of the SOAP envelope to the given
// endpoint and unmarshals the SOAP response into a SubmitResult.
func (c *Client) doRequest(ctx context.Context, endpoint string, envelope []byte) (*SubmitResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		bytes.NewReader(envelope))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", SOAPAction)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var soapResp soapResponse
	if err := xml.NewDecoder(resp.Body).Decode(&soapResp); err != nil {
		return nil, fmt.Errorf("parsing SOAP response: %w", err)
	}

	result := &SubmitResult{
		HTTPStatus:  resp.StatusCode,
		CSV:         soapResp.Body.Respuesta.CSV,
		Estado:      soapResp.Body.Respuesta.EstadoEnvio,
		Descripcion: soapResp.Body.Respuesta.DescripcionEstadoEnvio,
	}
	return result, nil
}
