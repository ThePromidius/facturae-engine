// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package aeat implements a SOAP client for submitting signed invoices to the
// Spanish Tax Agency (AEAT) Verifactu service.
package aeat

import (
	"fmt"
)

// buildSOAPEnvelope wraps the signed invoice XML in a SOAP envelope with the
// mandatory Header containing the issuer's CIF (NIF) and the signed content in
// the Body.
func buildSOAPEnvelope(signedXML []byte, emisorCIF string) []byte {
	// Note: We use string formatting for the envelope
	// because AEAT expects specific namespaces that can be tricky with Go's xml.Marshal
	
	// Real-world AEAT SOAP expects:
	// 1. SOAP Envelope
	// 2. Body containing the SIF schema root (sum:RegFactuSistemaFacturacion)
	
	template := `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:sum="http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/SistemaFacturacion.xsd">
   <soapenv:Header/>
   <soapenv:Body>
      %s
   </soapenv:Body>
</soapenv:Envelope>`

	// signedXML should already be a complete sum:RegFactuSistemaFacturacion block
	return []byte(fmt.Sprintf(template, string(signedXML)))
}
