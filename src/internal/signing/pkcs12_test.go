// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package signing_test

import (
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/signing"
)

func TestOpenSSLAvailable_ReturnsBoolean(t *testing.T) {
	available := signing.OpenSSLAvailable()
	t.Logf("OpenSSL available: %v", available)
}

func TestLoadFromP12File_MissingFile(t *testing.T) {
	_, err := signing.LoadFromP12File("/nonexistent/path/cert.p12", "password")
	if err == nil {
		t.Error("expected error for missing .p12 file")
	}
}

func TestLoadTLSCertFromP12_MissingFile(t *testing.T) {
	_, err := signing.LoadTLSCertFromP12("/nonexistent/path/cert.p12", "password")
	if err == nil {
		t.Error("expected error for missing .p12 file")
	}
}
