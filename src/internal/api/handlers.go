// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)


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

// handleInvoice processes a POST /invoice request.
// It supports both JSON (transformation mode) and XML (native mode).
func (s *Server) handleInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var xmlFull []byte
	var emisorCIF, number, series string
	var issueDate time.Time
	var total float64
	var version = "3.2.2" // Default

	if strings.Contains(contentType, "application/json") {
		// --- JSON MODE ---
		var req invoice.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
			return
		}
		
		// Centralized Validation
		vRes := s.validator.ValidateRequest(req)
		if !vRes.Valid {
			writeError(w, http.StatusBadRequest, strings.Join(vRes.Errors, "; "))
			return
		}

		req.Meta = invoice.DefaultMeta(req.Meta)
		version = req.Meta.Version
		
		f, err := facturae.Build(req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error construyendo FacturaE: "+err.Error())
			return
		}
		
		xmlBytes, _ := xml.MarshalIndent(f, "", "  ")
		xmlFull = append([]byte(xml.Header), xmlBytes...)
		
		// XSD Pre-flight
		vXML := s.validator.ValidateFacturaEXML(xmlFull, version)
		if !vXML.Valid {
			writeError(w, http.StatusUnprocessableEntity, strings.Join(vXML.Errors, "; "))
			return
		}

		inv := f.Invoices.Invoice[0]
		emisorCIF = f.Parties.SellerParty.TaxIdentification.TaxIdentificationNumber
		number = inv.InvoiceHeader.InvoiceNumber
		series = inv.InvoiceHeader.InvoiceSeriesCode
		issueDate, _ = time.Parse("2006-01-02", inv.InvoiceIssueData.IssueDate)
		total = inv.InvoiceTotals.InvoiceTotal

	} else if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		// --- XML MODE ---
		rawXML, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Error leyendo XML: "+err.Error())
			return
		}
		
		// XSD Pre-flight
		vXML := s.validator.ValidateFacturaEXML(rawXML, version)
		if !vXML.Valid {
			writeError(w, http.StatusUnprocessableEntity, strings.Join(vXML.Errors, "; "))
			return
		}

		meta, err := facturae.ExtractMetadata(rawXML)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "Error extrayendo metadata: "+err.Error())
			return
		}
		xmlFull = rawXML
		emisorCIF = meta.IssuerNIF
		number = meta.Number
		series = meta.Series
		issueDate = meta.Date
		total = meta.Total
	} else {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type debe ser application/json o application/xml")
		return
	}

	// --- COMMON: Chaining ---
	rec, err := s.chain.Append(number, series, emisorCIF, issueDate, total)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error en cadena Verifactu: "+err.Error())
		return
	}

	// --- COMMON: Patching & Signing ---
	if strings.Contains(contentType, "xml") {
		xmlFull = facturae.PatchVerifactu(xmlFull, rec.Fingerprint, rec.PreviousFingerprint)
	}
	
	signed, err := s.signer.Sign(xmlFull)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error firmando XML: "+err.Error())
		return
	}

	// --- COMMON: AEAT Submission ---
	if s.aeat != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			result, err := s.aeat.Submit(ctx, signed, emisorCIF)
			if err != nil {
				fmt.Printf("[aeat] Error enviando factura %s: %v\n", number, err)
				return
			}
			fmt.Printf("[aeat] Factura %s: %s\n", number, result.MapResult())
		}()
	}

	// --- RESPONSE ---
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("X-Verifactu-Fingerprint", rec.Fingerprint)
	w.Header().Set("X-Chain-Length", fmt.Sprintf("%d", s.chain.Len()))
	w.WriteHeader(http.StatusOK)
	w.Write(signed)
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
