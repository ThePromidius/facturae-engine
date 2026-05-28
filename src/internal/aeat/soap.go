// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package aeat

import "fmt"

func buildSOAPEnvelope(suministroLRXML []byte) []byte {
	template := `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:sfLR="%s">
   <soapenv:Header/>
   <soapenv:Body>
      %s
   </soapenv:Body>
</soapenv:Envelope>`

	return []byte(fmt.Sprintf(template, NSLR, string(suministroLRXML)))
}
