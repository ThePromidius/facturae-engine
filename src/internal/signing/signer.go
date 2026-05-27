// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package signing defines the interface and implementations for XAdES-BES XML digital signatures
// used in Facturae electronic invoices.
package signing

// Signer defines the interface for XML digital signature providers.
// Implementations must be able to sign raw XML bytes and report the signing algorithm used.
type Signer interface {
	// Sign applies an XML signature to the given document bytes and returns
	// the signed result.
	Sign(xmlData []byte) ([]byte, error)
	// Algorithm returns the name of the signing algorithm used.
	Algorithm() string
}
