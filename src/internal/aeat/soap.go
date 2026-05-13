// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package aeat implements a SOAP client for submitting signed invoices to the
// Spanish Tax Agency (AEAT) Verifactu service.
package aeat

import (
	"encoding/xml"
)

// buildSOAPEnvelope wraps the signed invoice XML in a SOAP envelope with the
// mandatory Header containing the issuer's CIF (NIF) and the signed content in
// the Body.
func buildSOAPEnvelope(signedXML []byte, emisorCIF string) []byte {
	envelope := struct {
		XMLName   xml.Name `xml:"soapenv:Envelope"`
		XmlnsSoap string   `xml:"xmlns:soapenv,attr"`
		XmlnsSum  string   `xml:"xmlns:sum,attr"`
		Header    struct {
			Cabecera struct {
				Obligado struct {
					NIF string `xml:"sum:NIF"`
				} `xml:"sum:ObligadoEmision"`
			} `xml:"sum:Cabecera"`
		} `xml:"soapenv:Header"`
		Body struct {
			RegFactu struct {
				Content []byte `xml:",innerxml"`
			} `xml:"sum:RegFactuSistemaFacturacion"`
		} `xml:"soapenv:Body"`
	}{}

	envelope.XmlnsSoap = "http://schemas.xmlsoap.org/soap/envelope/"
	envelope.XmlnsSum = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd"
	envelope.Header.Cabecera.Obligado.NIF = emisorCIF
	envelope.Body.RegFactu.Content = signedXML

	out, _ := xml.MarshalIndent(envelope, "", "  ")
	return append([]byte(xml.Header), out...)
}
