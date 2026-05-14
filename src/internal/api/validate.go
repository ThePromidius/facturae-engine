// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
	"github.com/ThePromidius/facturae-engine/src/internal/validation"
)

// handleValidate handles POST /validate for testing invoices without emission.
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var result validation.Result

	if strings.Contains(contentType, "application/json") {
		var req invoice.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		result = s.validator.ValidateRequest(req)
	} else if strings.Contains(contentType, "xml") {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Read error")
			return
		}
		// Defaulting to 3.2.2 for validation testing
		result = s.validator.ValidateFacturaEXML(data, "3.2.2")
	} else {
		writeError(w, http.StatusUnsupportedMediaType, "Unsupported format")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
