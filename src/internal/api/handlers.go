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

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/qr"
	"github.com/ThePromidius/facturae-engine/src/internal/ubl"
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
	var chainRec *chain.Record

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
		
		// Chaining (Need it before building for UBL extension)
		// We'll calculate emisorCIF, number, series, issueDate, total from JSON
		emisorCIF = req.Emisor.CIF
		number = req.Factura.Numero
		series = req.Factura.Serie
		issueDate = req.Factura.Fecha
		// Calculate total for chaining
		var lineExtTotal float64
		var taxTotal float64
		for _, l := range req.Lineas {
			lineTotal := l.PrecioUnitario * l.QuantityFallback()
			lineExtTotal += lineTotal
			taxTotal += lineTotal * (l.IVATipo / 100.0)
		}
		total = lineExtTotal + taxTotal

		rec, err := s.chain.Append(number, series, emisorCIF, issueDate, total)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error en cadena Verifactu: "+err.Error())
			return
		}
		chainRec = &rec

		if req.Meta.Format == "ubl" {
			uBuilder := &ubl.Builder{}
			xmlFull, err = uBuilder.BuildWithCompliance(req, rec.Fingerprint, rec.PreviousFingerprint)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Error construyendo UBL: "+err.Error())
				return
			}
		} else {
			f, err := facturae.Build(req)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Error construyendo FacturaE: "+err.Error())
				return
			}
			xmlBytes, _ := xml.MarshalIndent(f, "", "  ")
			xmlFull = append([]byte(xml.Header), xmlBytes...)
		}
		
		// XSD Pre-flight (FacturaE only for now in strict mode)
		if req.Meta.Format != "ubl" {
			vXML := s.validator.ValidateFacturaEXML(xmlFull, version)
			if !vXML.Valid {
				writeError(w, http.StatusUnprocessableEntity, strings.Join(vXML.Errors, "; "))
				return
			}
		}

	} else if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		// --- XML MODE ---
		rawXML, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Error leyendo XML: "+err.Error())
			return
		}
		
		// XSD Pre-flight (Detecting FacturaE version or UBL)
		// For now we try 3.2.2 as default for native XML
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

		// Chaining
		rec, err := s.chain.Append(number, series, emisorCIF, issueDate, total)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error en cadena Verifactu: "+err.Error())
			return
		}
		chainRec = &rec
		
		// Patching metadata
		xmlFull = facturae.PatchVerifactu(xmlFull, rec.Fingerprint, rec.PreviousFingerprint)

	} else {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type debe ser application/json o application/xml")
		return
	}

	// --- COMMON: Signing ---
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
	w.Header().Set("X-Verifactu-Fingerprint", chainRec.Fingerprint)
	w.Header().Set("X-Chain-Length", fmt.Sprintf("%d", s.chain.Len()))
	
	// QR URL Generation (FacturaE specific)
	if emisorCIF != "" && number != "" {
		qrURL := qr.VerificationURL(qr.VerifactuParams{
			EmisorCIF:   emisorCIF,
			Numero:      number,
			Serie:       series,
			Fecha:       issueDate.Format("2006-01-02"),
			Total:       total,
			Fingerprint: chainRec.Fingerprint,
		})
		if dataURI, err := qr.GenerateDataURI(qrURL); err == nil {
			w.Header().Set("X-Verifactu-QR-URL", qrURL)
			w.Header().Set("X-Verifactu-QR-DataURI", dataURI[:min(len(dataURI), 200)]+"...")
		}
	}

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
