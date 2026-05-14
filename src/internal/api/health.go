// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/qr"
)

// handleHealth handles GET /health requests, returning detailed component status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbStatus := "connected"
	
	// Check Database Health
	if cs, ok := s.chain.Store().(*chain.SQLStore); ok {
		if err := cs.CheckHealth(); err != nil {
			status = "degraded"
			dbStatus = fmt.Sprintf("error: %v", err)
		}
	} else if _, ok := s.chain.Store().(*chain.MemoryStore); ok {
		dbStatus = "memory (non-persistent)"
	}

	// Check Chain Integrity
	chainStatus := "intact"
	if err := s.chain.Verify(); err != nil {
		status = "degraded"
		chainStatus = fmt.Sprintf("broken: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"components": map[string]string{
			"database": dbStatus,
			"chain":    chainStatus,
			"signer":   s.signer.Algorithm(),
		},
		"metrics": map[string]interface{}{
			"chain_length": s.chain.Len(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleDashboard returns a simple HTML page to visualize engine health.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>FacturaE Sidecar Dashboard</title>
		<style>
			body { font-family: sans-serif; margin: 40px; background: #f4f7f6; }
			.card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
			.status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-weight: bold; }
			.ok { background: #e6fffa; color: #2c7a7b; }
			.error { background: #fff5f5; color: #c53030; }
			table { width: 100%; border-collapse: collapse; margin-top: 20px; }
			th, td { text-align: left; padding: 12px; border-bottom: 1px solid #edf2f7; }
		</style>
	</head>
	<body>
		<h1>FacturaE Sidecar</h1>
		<div class="card">
			<h2>System Health</h2>
			<table>
				<tr><th>Component</th><th>Status</th></tr>
				<tr><td>Engine</td><td><span class="status ok">Running</span></td></tr>
				<tr><td>Signer</td><td>%s</td></tr>
				<tr><td>Chain Length</td><td>%d records</td></tr>
			</table>
			<p><small>Last Updated: %s</small></p>
		</div>
	</body>
	</html>`
	
	fmt.Fprintf(w, html, s.signer.Algorithm(), s.chain.Len(), time.Now().Format(time.RFC1123))
}

// handleQR handles GET /qr requests, generating a PNG QR code from the ?text=
// query parameter.
func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		writeError(w, http.StatusBadRequest, "missing ?text= parameter")
		return
	}
	pngData, err := qr.GeneratePNG(text)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "QR generation failed: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(pngData)
}

// writeError writes a JSON error response with the given HTTP status and
// message.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
