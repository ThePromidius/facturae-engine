// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package invoice

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var cifRe = regexp.MustCompile(`^(?:[A-Za-z]\d{7}[A-Za-z0-9]|\d{8}[A-Za-z])$`)

// ValidationError collects multiple validation error messages and implements the error interface.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return "validation failed:\n  * " + strings.Join(e.Errors, "\n  * ")
}

func (e *ValidationError) add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) any() bool { return len(e.Errors) > 0 }

func validateCIF(cif, label string) error {
	if len(cif) != 9 {
		return fmt.Errorf("%s.cif: debe tener 9 caracteres (recibido %d)", label, len(cif))
	}
	if !cifRe.MatchString(cif) {
		return fmt.Errorf("%s.cif: formato invalido (debe ser 1 letra + 7 digitos + 1 digito/letra)", label)
	}
	return nil
}

// Validate checks an invoice.Request for required fields, valid ranges, and structural completeness.
// It returns nil if the request is valid, or a ValidationError listing all issues found.
func Validate(req Request) error {
	ve := &ValidationError{}

	if strings.TrimSpace(req.Factura.Numero) == "" {
		ve.add("factura.numero: campo obligatorio")
	}
	if req.Factura.Fecha.IsZero() {
		ve.add("factura.fecha: fecha invalida o ausente")
	}

	if err := validatePartyFull(req.Emisor, "emisor"); err != nil {
		var partErr *ValidationError
		if errors.As(err, &partErr) {
			ve.Errors = append(ve.Errors, partErr.Errors...)
		}
	}

	if strings.TrimSpace(req.Receptor.CIF) == "" {
		ve.add("receptor.cif: campo obligatorio")
	} else if err := validateCIF(req.Receptor.CIF, "receptor"); err != nil {
		ve.add(err.Error())
	}
	if strings.TrimSpace(req.Receptor.Nombre) == "" {
		ve.add("receptor.nombre: campo obligatorio")
	}

	if len(req.Lineas) == 0 {
		ve.add("lineas: debe incluir al menos una linea de factura")
	}
	for i, l := range req.Lineas {
		prefix := fmt.Sprintf("lineas[%d]", i)
		if strings.TrimSpace(l.Descripcion) == "" {
			ve.add(prefix + ".desc: campo obligatorio")
		}
		if l.Cantidad <= 0 {
			ve.add(prefix + ".cantidad: debe ser mayor que cero")
		}
		if l.PrecioUnitario < 0 {
			ve.add(prefix + ".precio_unitario: no puede ser negativo")
		}
		if l.IVATipo < 0 || l.IVATipo > 100 {
			ve.add(fmt.Sprintf("%s.iva_tipo: valor invalido (%.2f)", prefix, l.IVATipo))
		}
	}

	if ve.any() {
		return ve
	}
	return nil
}

// validatePartyFull validates all required fields for a party (CIF, name, address, post code, city, country).
// It is used for the issuer (emisor) where full address details are mandatory.
func validatePartyFull(p Party, label string) error {
	ve := &ValidationError{}
	if strings.TrimSpace(p.CIF) == "" {
		ve.add(label + ".cif: campo obligatorio")
	} else if err := validateCIF(p.CIF, label); err != nil {
		ve.add(err.Error())
	}
	if strings.TrimSpace(p.Nombre) == "" {
		ve.add(label + ".nombre: campo obligatorio")
	}
	if strings.TrimSpace(p.Direccion) == "" {
		ve.add(label + ".direccion: campo obligatorio para el emisor")
	}
	if strings.TrimSpace(p.CP) == "" {
		ve.add(label + ".cp: campo obligatorio para el emisor")
	}
	if strings.TrimSpace(p.Ciudad) == "" {
		ve.add(label + ".ciudad: campo obligatorio para el emisor")
	}
	if strings.TrimSpace(p.Pais) == "" {
		ve.add(label + ".pais: campo obligatorio para el emisor")
	}
	if ve.any() {
		return ve
	}
	return nil
}
