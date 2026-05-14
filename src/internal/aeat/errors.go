// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package aeat

import "fmt"

// ErrorRegistry provides human-readable translations for AEAT Verifactu error codes.
// Derived from the official errores.properties (v1.0 2026).
var ErrorRegistry = map[string]string{
	// 4xxx: Global Rejection Errors
	"4102": "El XML no cumple el esquema (Falta campo obligatorio).",
	"4107": "El NIF del emisor no está identificado en el censo de la AEAT.",
	"4109": "El formato del NIF es incorrecto.",
	"4112": "El titular del certificado no tiene permisos para este NIF (Apoderamiento necesario).",
	"4136": "Error en el encadenamiento: El registro anterior no es correcto.",
	"4141": "Acceso suspendido temporalmente por la AEAT. Contacte con soporte.",

	// 1xxx: Invoice Rejection Errors
	"1104": "El número o serie de la factura es incorrecto.",
	"1108": "El NIF de la factura debe coincidir con el del certificado.",
	"1112": "La fecha de expedición no puede ser futura.",
	"1152": "La fecha de expedición no puede ser anterior al 28 de octubre de 2024.",
	"1210": "El ImporteTotal no coincide con el sumatorio de bases y cuotas.",
	"1262": "La longitud de la huella (hash) es incorrecta.",
	"1278": "La huella actual no puede ser igual a la anterior.",
	"3000": "Factura duplicada: Este número y serie ya han sido registrados.",

	// 2xxx: Accepted with Errors (Action required)
	"2000": "Aceptado con errores: El cálculo de la huella es incorrecto.",
	"2001": "Aceptado con errores: El NIF del destinatario no está en el censo.",
	"2004": "Aceptado con errores: La diferencia horaria con la AEAT supera el minuto.",
}

// GetAdvice returns a human-readable explanation and advice for a given AEAT error code.
func GetAdvice(code string) string {
	msg, ok := ErrorRegistry[code]
	if !ok {
		return fmt.Sprintf("Error desconocido (%s). Consulte el manual técnico de la AEAT.", code)
	}
	return msg
}

// MapResult provides a high-level summary of the submission outcome.
func (r *SubmitResult) MapResult() string {
	if r.IsAccepted() {
		if r.Estado == "AceptadoConErrores" {
			return "ALERTA: Aceptado con errores técnicos. Requiere revisión."
		}
		return "ÉXITO: Registro procesado correctamente por la AEAT."
	}
	
	advice := GetAdvice(r.Estado) // In some AEAT responses, ErrorCode is in the 'Estado' field or 'Descripcion'
	return fmt.Sprintf("RECHAZADO: %s", advice)
}
