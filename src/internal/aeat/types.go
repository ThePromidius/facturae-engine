// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package aeat implements a SOAP client for submitting signed invoices to the
// Spanish Tax Agency (AEAT) Verifactu service.
package aeat

import (
	"encoding/xml"
)

// Environment represents the AEAT target environment (test or production).
type Environment string

const (
	// EnvTest is the AEAT pre-production testing environment.
	EnvTest Environment = "test"
	// EnvProd is the AEAT production environment.
	EnvProd Environment = "prod"
)

// Endpoints maps each Environment to its corresponding AEAT Verifactu SOAP
// endpoint URL.
var Endpoints = map[Environment]string{
	EnvTest: "https://prewww1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
	EnvProd: "https://www1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
}

// SOAPAction is the fixed SOAPAction header value required by the AEAT
// Verifactu service.
const SOAPAction = "http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/RegistroFacturacion"

// SubmitResult holds the response returned by the AEAT Verifactu service after
// submitting an invoice, including the CSV assignment and the processing status.
type SubmitResult struct {
	CSV         string
	Estado      string
	Descripcion string
	HTTPStatus  int
}

// IsAccepted reports whether the submission was accepted, either with a
// "Correcto" or an "AceptadoConErrores" status.
func (r SubmitResult) IsAccepted() bool {
	return r.Estado == "Correcto" || r.Estado == "AceptadoConErrores"
}

// SuministroLR is the top-level XML element for the Verifactu invoice
// submission payload.
type SuministroLR struct {
	XMLName  xml.Name `xml:"sum:RegFactuSistemaFacturacion"`
	XmlnsSum string   `xml:"xmlns:sum,attr"`
	Cabecera Cabecera `xml:"sum:Cabecera"`
	Registro Registro `xml:"sum:RegistroFacturacion"`
}

// Cabecera contains the header information for a Verifactu submission,
// including the obligated issuer.
type Cabecera struct {
	ObligadoEmision Obligado `xml:"sum:ObligadoEmision"`
}

// Obligado identifies the obligated issuer of the invoice by their tax
// identification number (NIF).
type Obligado struct {
	NIF string `xml:"sum:NIF"`
}

// Registro represents a single invoice record within a Verifactu submission.
type Registro struct {
	IDFactura      IDFactura `xml:"sum:IDFactura"`
	FechaOperacion string    `xml:"sum:FechaOperacion"`
	TipoFactura    string    `xml:"sum:TipoFactura"`
	CuotaTotal     float64   `xml:"sum:CuotaTotal"`
	ImporteTotal   float64   `xml:"sum:ImporteTotal"`
	Huella         string    `xml:"sum:Huella"`
	FechaHoraHito  string    `xml:"sum:FechaHoraHito"`
	SistemaInformatico Sistema `xml:"sum:SistemaInformatico"`
}

// IDFactura groups the identifiers that uniquely reference an invoice: the
// issuer, the serial number, and the issue date.
type IDFactura struct {
	IDEmisorFactura IDEmisor `xml:"sum:IDEmisorFactura"`
	NumSerieFactura string   `xml:"sum:NumSerieFactura"`
	FechaExpedicion string   `xml:"sum:FechaExpedicionFactura"`
}

// IDEmisor holds the tax identifier (NIF) of the invoice issuer.
type IDEmisor struct {
	NIF string `xml:"sum:NIF"`
}

// Sistema describes the software system used to generate the invoice record.
type Sistema struct {
	Nombre           string `xml:"sum:Nombre"`
	Version          string `xml:"sum:Version"`
	NumeroSerie      string `xml:"sum:NumeroSerie"`
	NIFEntidadDesarr string `xml:"sum:NIFEntidadDesarr"`
}

// soapResponse is the internal type used to unmarshal the AEAT SOAP response.
type soapResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Respuesta struct {
			CSV                    string `xml:"CSV"`
			EstadoEnvio            string `xml:"EstadoEnvio"`
			DescripcionEstadoEnvio string `xml:"DescripcionEstadoEnvio"`
		} `xml:"RespuestaRegFactuSistemaFacturacion"`
	} `xml:"Body"`
}
