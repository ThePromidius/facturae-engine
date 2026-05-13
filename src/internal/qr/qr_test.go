// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package qr_test

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/qr"
)

func TestVerificationURL_ContainsAllParams(t *testing.T) {
	p := qr.VerifactuParams{
		EmisorCIF:   "B12345678",
		Numero:      "2024-001",
		Serie:       "F",
		Fecha:       "2024-05-12",
		Total:       1270.50,
		Fingerprint: "abc123def456",
	}
	u := qr.VerificationURL(p)

	checks := []string{"B12345678", "2024-001", "1270.50", "abc123def456", "agenciatributaria"}
	for _, s := range checks {
		if !strings.Contains(u, s) {
			t.Errorf("URL should contain %q\nGot: %s", s, u)
		}
	}
}

func TestVerificationURL_IsValidURL(t *testing.T) {
	p := qr.VerifactuParams{EmisorCIF: "B12345678", Numero: "001", Total: 100}
	u := qr.VerificationURL(p)
	if !strings.HasPrefix(u, "https://") {
		t.Errorf("expected HTTPS URL, got: %s", u)
	}
}

func TestGeneratePNG_ProducesValidPNG(t *testing.T) {
	data, err := qr.GeneratePNG("https://example.com/test")
	if err != nil {
		t.Fatalf("GeneratePNG error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GeneratePNG returned empty data")
	}
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("output is not a valid PNG: %v", err)
	}
}

func TestGeneratePNG_HasPNGHeader(t *testing.T) {
	data, err := qr.GeneratePNG("https://hacienda.es")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("output does not start with PNG magic bytes")
	}
}

func TestGenerateDataURI_HasCorrectPrefix(t *testing.T) {
	uri, err := qr.GenerateDataURI("https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Errorf("expected data URI prefix, got: %s...", uri[:30])
	}
}

func TestGeneratePNG_ShortText(t *testing.T) {
	_, err := qr.GeneratePNG("Hi")
	if err != nil {
		t.Errorf("should handle very short text: %v", err)
	}
}

func TestGeneratePNG_LongURLAtLimit(t *testing.T) {
	long := "https://www2.agenciatributaria.gob.es/wlpl/VERIFACTU/ConsultaPublica?nif=B12345678&num=2024-LONG-001&ser=FFFF&fec=2024-05-12&tot=99999.99&hp=" + strings.Repeat("a", 20)
	_, err := qr.GeneratePNG(long)
	if err != nil {
		t.Errorf("unexpected error for URL within limit: %v", err)
	}
}

func TestGeneratePNG_TooLong(t *testing.T) {
	tooLong := strings.Repeat("x", 400)
	_, err := qr.GeneratePNG(tooLong)
	if err == nil {
		t.Error("expected error for text > 300 chars")
	}
}

func TestGeneratePNG_DifferentTextsProduceDifferentImages(t *testing.T) {
	a, _ := qr.GeneratePNG("https://example.com/invoice/001")
	b, _ := qr.GeneratePNG("https://example.com/invoice/002")
	if bytes.Equal(a, b) {
		t.Error("different texts should produce different QR images")
	}
}
