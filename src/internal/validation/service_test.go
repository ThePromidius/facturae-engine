// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package validation

import (
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/schema"
)

func TestService_ValidateFacturaEXML(t *testing.T) {
	sm, _ := schema.NewManager(t.TempDir())
	svc := NewService(sm)

	tests := []struct {
		name    string
		xml     string
		version string
		valid   bool
	}{
		{
			name: "Valid Minimal FacturaE",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<fe:Facturae xmlns:fe="http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml">
	<FileHeader><SchemaVersion>3.2.2</SchemaVersion><Modality>I</Modality></FileHeader>
</fe:Facturae>`,
			version: "3.2.2",
			valid:   true,
		},
		{
			name: "Malformed XML",
			xml: `<?xml version="1.0"?><WrongRoot>`,
			version: "3.2.2",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := svc.ValidateFacturaEXML([]byte(tt.xml), tt.version)
			if res.Valid != tt.valid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.valid, res.Valid, res.Errors)
			}
		})
	}
}
