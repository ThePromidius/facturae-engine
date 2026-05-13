// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package legal implements the mandatory documentation and certification requirements
// for the Verifactu regulation.
package legal

import (
	"fmt"
	"time"
)

// SoftwareInfo holds the metadata required for the Declaración Responsable
// as specified in Art. 15 of Orden HAC/1177/2024.
type SoftwareInfo struct {
	Name        string
	Version     string
	ProducerNIF string
}

// GenerateDeclaracionResponsable produces the text of the declaration certifying
// that the software complies with RD 1007/2023 and Orden HAC/1177/2024.
func GenerateDeclaracionResponsable(info SoftwareInfo) string {
	return fmt.Sprintf(`DECLARACIÓN RESPONSABLE DE CUMPLIMIENTO NORMATIVO

En cumplimiento de lo dispuesto en el artículo 15 de la Orden HAC/1177/2024, de 17 de octubre, 
y en el artículo 12 del Reglamento aprobado por el Real Decreto 1007/2023, de 5 de diciembre.

SOFTWARE: %s
VERSIÓN: %s
NIF DEL PRODUCTOR: %s

DECLARA:
Que el sistema informático arriba identificado ha sido diseñado y desarrollado de modo que 
garantiza la integridad, conservación, accesibilidad, legibilidad, trazabilidad e 
inalterabilidad de los registros de facturación, sin interpolaciones, omisiones o 
alteraciones de las que no quede la debida anotación en el sistema mismo.

Asimismo, se certifica que el sistema cumple con todas las especificaciones técnicas 
y funcionales exigidas por la normativa Veri*factu.

Fecha de generación: %s
`, info.Name, info.Version, info.ProducerNIF, time.Now().Format("02/01/2006"))
}
