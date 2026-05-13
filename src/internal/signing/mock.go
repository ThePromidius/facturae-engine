// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package signing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MockSigner is a no-op implementation of Signer that appends a mock signature
// comment instead of a real XML digital signature. Useful for testing and development.
type MockSigner struct{}

// Algorithm returns the mock algorithm identifier "mock/no-op".
func (MockSigner) Algorithm() string { return "mock/no-op" }

// Sign appends an XML comment containing a SHA-256 fingerprint of the data as a mock signature.
// It does not produce a valid XML digital signature.
func (MockSigner) Sign(xmlData []byte) ([]byte, error) {
	h := sha256.Sum256(xmlData)
	comment := fmt.Sprintf(
		"\n<!-- XAdES-BES MOCK SIGNATURE | SHA-256: %s -->",
		hex.EncodeToString(h[:]),
	)
	return append(xmlData, []byte(comment)...), nil
}
