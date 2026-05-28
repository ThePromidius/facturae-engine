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

type Client struct {
	env      Environment
	http     *http.Client
	maxTries int
}

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

func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

func (c *Client) SetMaxTries(n int) {
	c.maxTries = n
}

func (c *Client) Submit(ctx context.Context, suministroLRXML []byte) (*SubmitResult, error) {
	envelope := buildSOAPEnvelope(suministroLRXML)
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
		Descripcion: soapResp.Body.Respuesta.EstadoEnvio,
	}
	return result, nil
}
