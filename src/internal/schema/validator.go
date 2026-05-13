// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package schema manages FacturaE XSD schema files: it caches them locally and
// provides validation of XML documents against the cached schemas via xmllint.
package schema

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// ErrXmllintMissing is returned when the xmllint binary is not found in the
// system PATH.
var ErrXmllintMissing = errors.New("xmllint no esta instalado")

// ValidateXML validates xmlBytes against the XSD schema at schemaPath using
// the xmllint command-line tool. It returns ErrXmllintMissing if xmllint is not
// installed.
func ValidateXML(xmlBytes []byte, schemaPath string) error {
	if _, err := exec.LookPath("xmllint"); err != nil {
		return ErrXmllintMissing
	}

	if _, err := os.Stat(schemaPath); err != nil {
		return fmt.Errorf("el archivo esquema XSD no se encuentra en %s: %w", schemaPath, err)
	}

	cmd := exec.Command("xmllint", "--noout", "--schema", schemaPath, "-")
	cmd.Stdin = bytes.NewReader(xmlBytes)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := bytes.TrimSpace(stderr.Bytes())
		if len(errMsg) == 0 {
			errMsg = []byte(err.Error())
		}
		return fmt.Errorf("validacion XSD fallida:\n%s", string(errMsg))
	}

	return nil
}
