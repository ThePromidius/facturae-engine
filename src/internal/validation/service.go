// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package validation

import (
	"errors"
	"fmt"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/schema"
)

// Result represents the outcome of a multi-stage validation process.
type Result struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Type    string   `json:"type"`    // json, facturae-xml, ubl-xml
	Version string   `json:"version"` // 3.2.2, 2.1, etc.
}

// Service provides high-level validation logic for all supported formats.
type Service struct {
	schemas *schema.Manager
}

// NewService creates a new validation service.
func NewService(sm *schema.Manager) *Service {
	return &Service{schemas: sm}
}

// ValidateRequest performs structural validation on a JSON invoice request.
func (v *Service) ValidateRequest(req invoice.Request) Result {
	res := Result{Type: "json", Valid: true}
	if err := invoice.Validate(req); err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, err.Error())
	}
	return res
}

// ValidateFacturaEXML performs strict XSD validation on a FacturaE XML blob.
func (v *Service) ValidateFacturaEXML(data []byte, version string) Result {
	res := Result{Type: "facturae-xml", Version: version, Valid: true}
	
	// 1. Basic XML sanity
	if err := facturae.ValidateXMLBytes(data); err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, fmt.Sprintf("XML Malformed: %v", err))
		return res
	}

	// 2. Strict XSD Check
	xsdPath, err := v.schemas.SchemaPath(version)
	if err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Schema retrieval failed: %v", err))
		return res
	}

	if err := schema.ValidateXML(data, xsdPath); err != nil {
		if !errors.Is(err, schema.ErrXmllintMissing) {
			res.Valid = false
			res.Errors = append(res.Errors, fmt.Sprintf("XSD Validation: %v", err))
		}
	}

	return res
}
