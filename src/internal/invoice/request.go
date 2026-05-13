// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package invoice defines the input data model for generating electronic invoices.
// It provides types for requests, parties, line items, payment info, and administrative codes.
package invoice

import "time"

// Request is the top-level input for invoice generation. It contains metadata,
// invoice identifiers, the issuing and receiving parties, line items, and optional payment info.
type Request struct {
	Meta    Meta    `json:"meta"`
	Factura Factura `json:"factura"`
	Emisor  Party   `json:"emisor"`
	Receptor Party  `json:"receptor"`
	Lineas  []Linea `json:"lineas"`
	Pago    *Pago   `json:"pago,omitempty"`
}

// Meta holds version and currency metadata for the invoice generation process.
type Meta struct {
	Version string `json:"version_formato"`
	Moneda  string `json:"moneda"`
}

// Factura contains the invoice identification fields: number, series code, and issue date.
type Factura struct {
	Numero string    `json:"numero"`
	Serie  string    `json:"serie"`
	Fecha  time.Time `json:"fecha"`
}

// Party represents an invoicing party (issuer or receiver) with tax ID, legal name, address, and optional DIR3 codes.
type Party struct {
	CIF        string `json:"cif"`
	Nombre     string `json:"nombre"`
	Direccion  string `json:"direccion,omitempty"`
	CP         string `json:"cp,omitempty"`
	Ciudad     string `json:"ciudad,omitempty"`
	Provincia  string `json:"provincia,omitempty"`
	Pais       string `json:"pais,omitempty"`
	DIR3       *DIR3  `json:"dir3,omitempty"`
}

// Linea is a single invoice line item with description, quantity, unit price, and VAT rate.
type Linea struct {
	Descripcion    string  `json:"desc"`
	Cantidad       float64 `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	IVATipo        float64 `json:"iva_tipo"`
}

// Pago contains payment-related information for the invoice, including the status, date, and method.
type Pago struct {
	Estado   string     `json:"estado"`
	FechaPago *time.Time `json:"fecha_pago,omitempty"`
	Metodo   string     `json:"metodo,omitempty"`
}

// DIR3 holds Spanish public administration organizational codes for the invoicing party.
// These codes are used when invoicing Spanish government entities.
type DIR3 struct {
	OficinaContable   string `json:"oficina_contable"`
	OrganoGestor      string `json:"organo_gestor"`
	UnidadTramitadora string `json:"unidad_tramitadora"`
}

// DefaultMeta fills in default values for empty Meta fields:
// version defaults to "3.2.2" and currency defaults to "EUR".
func DefaultMeta(m Meta) Meta {
	if m.Version == "" {
		m.Version = "3.2.2"
	}
	if m.Moneda == "" {
		m.Moneda = "EUR"
	}
	return m
}
