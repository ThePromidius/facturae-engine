// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package qr provides QR code generation utilities for VERIFACTU invoice verification.
// It includes functions to build verification URLs and render QR codes as PNG images or data URIs.
package qr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/url"
)

// VerifactuParams holds the parameters needed to build a VERIFACTU verification URL
// for the Spanish tax agency (AEAT) invoice consultation system.
type VerifactuParams struct {
	EmisorCIF   string
	Numero      string
	Serie       string
	Fecha       string
	Total       float64
	Fingerprint string
}

// VerificationURL builds the Spanish tax agency VERIFACTU consultation URL from the given parameters
// according to Art. 21 Orden HAC/1177/2024.
func VerificationURL(p VerifactuParams) string {
	// Base URL for Verifactu verification in AEAT Sede Electrónica.
	base := "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/qr.html"
	v := url.Values{}
	v.Set("nif", p.EmisorCIF)
	v.Set("numserie", p.Serie+p.Numero)
	v.Set("fecha", p.Fecha)
	v.Set("importe", fmt.Sprintf("%.2f", p.Total))
	return base + "?" + v.Encode()
}

// GeneratePNG renders a QR code as a PNG image (RGBA, 4px per module, 4-module quiet zone)
// from the given text content using a simplified QR encoder.
func GeneratePNG(text string) ([]byte, error) {
	modules, size, err := encode(text)
	if err != nil {
		return nil, fmt.Errorf("qr: encoding %w", err)
	}

	const blockSize = 4
	const quiet = 4
	imgSize := (size + 2*quiet) * blockSize

	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			img.Set(x, y, color.White)
		}
	}

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if modules[row*size+col] {
				px := (col + quiet) * blockSize
				py := (row + quiet) * blockSize
				for dy := 0; dy < blockSize; dy++ {
					for dx := 0; dx < blockSize; dx++ {
						img.Set(px+dx, py+dy, color.Black)
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("qr: PNG encode: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateDataURI renders a QR code as a base64-encoded data URI in the format
// "data:image/png;base64,..." suitable for embedding in HTML or PDF documents.
func GenerateDataURI(text string) (string, error) {
	pngBytes, err := GeneratePNG(text)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	return "data:image/png;base64," + b64, nil
}
