// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package api provides the HTTP server that exposes FacturaE invoice processing
// endpoints.
package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/qr"
	"github.com/ThePromidius/facturae-engine/src/internal/schema"
)

// handleInvoice processes a POST /invoice request: it decodes the JSON invoice
// payload, builds the FacturaE XML, validates it structurally and against the
// XSD schema, signs it with the configured signer, appends a Verifactu chain
// record, optionally submits to AEAT, and returns the signed XML.
func (s *Server) handleInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req invoice.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	if err := invoice.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Meta = invoice.DefaultMeta(req.Meta)

	f, err := facturae.Build(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error construyendo FacturaE: "+err.Error())
		return
	}

	if err := facturae.ValidateStruct(f); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Validacion estructural: "+err.Error())
		return
	}

	xmlBytes, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error generando XML: "+err.Error())
		return
	}
	xmlFull := append([]byte(xml.Header), xmlBytes...)

	if err := facturae.ValidateXMLBytes(xmlFull); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Validacion XML: "+err.Error())
		return
	}

	xsdPath, err := s.schemas.SchemaPath(req.Meta.Version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error obteniendo esquema XSD: "+err.Error())
		return
	}
	if err := schema.ValidateXML(xmlFull, xsdPath); err != nil {
		if errors.Is(err, schema.ErrXmllintMissing) {
			fmt.Printf("[server] 'xmllint' no encontrado. Saltando validacion XSD estricta para %s\n", req.Meta.Version)
		} else {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	} else {
		fmt.Printf("[server] Validacion XSD estricta superada (%s)\n", req.Meta.Version)
	}

	signed, err := s.signer.Sign(xmlFull)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error firmando XML: "+err.Error())
		return
	}

	inv := f.Invoices.Invoice[0]
	issueDate, _ := time.Parse("2006-01-02", inv.InvoiceIssueData.IssueDate)
	emisorCIF := f.Parties.SellerParty.TaxIdentification.TaxIdentificationNumber
	rec, err := s.chain.Append(
		inv.InvoiceHeader.InvoiceNumber,
		inv.InvoiceHeader.InvoiceSeriesCode,
		emisorCIF,
		issueDate,
		inv.InvoiceTotals.InvoiceTotal,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error en cadena Verifactu: "+err.Error())
		return
	}

	fmt.Printf("[server] Factura %s firmada | Huella: %s\n",
		inv.InvoiceHeader.InvoiceNumber, rec.Fingerprint[:16]+"...")

	qrURL := qr.VerificationURL(qr.VerifactuParams{
		EmisorCIF:   emisorCIF,
		Numero:      inv.InvoiceHeader.InvoiceNumber,
		Serie:       inv.InvoiceHeader.InvoiceSeriesCode,
		Fecha:       inv.InvoiceIssueData.IssueDate,
		Total:       inv.InvoiceTotals.InvoiceTotal,
		Fingerprint: rec.Fingerprint,
	})
	if dataURI, err := qr.GenerateDataURI(qrURL); err == nil {
		w.Header().Set("X-Verifactu-QR-URL", qrURL)
		w.Header().Set("X-Verifactu-QR-DataURI", dataURI[:min(len(dataURI), 200)]+"...")
	}

	if s.aeat != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			result, err := s.aeat.Submit(ctx, signed, emisorCIF)
			if err != nil {
				fmt.Printf("[aeat] Error enviando factura %s: %v\n",
					inv.InvoiceHeader.InvoiceNumber, err)
				return
			}
			if result.IsAccepted() {
				fmt.Printf("[aeat] Factura %s aceptada | CSV: %s\n",
					inv.InvoiceHeader.InvoiceNumber, result.CSV)
			} else {
				fmt.Printf("[aeat] Factura %s rechazada: %s\n",
					inv.InvoiceHeader.InvoiceNumber, result.Descripcion)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("X-Verifactu-Fingerprint", rec.Fingerprint)
	w.Header().Set("X-Chain-Length", fmt.Sprintf("%d", s.chain.Len()))
	w.WriteHeader(http.StatusOK)
	w.Write(signed)
}

// handleChain handles GET /chain requests by returning all Verifactu chain
// records as a JSON array.
func (s *Server) handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Solo se permite GET")
		return
	}

	records := s.chain.All()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(records),
		"records": records,
	})
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
