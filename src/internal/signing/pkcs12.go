// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package signing

import (
	"crypto/tls"
	"fmt"
	"os/exec"
	"strings"
)

// LoadFromP12File extracts the private key and certificate from a .p12/PFX file using
// OpenSSL and returns a P12Signer ready to sign invoices. OpenSSL must be installed and
// available in the system PATH.
func LoadFromP12File(p12Path, password string) (*P12Signer, error) {
	if _, err := exec.LookPath("openssl"); err != nil {
		return nil, fmt.Errorf("signing: openssl not found in PATH - "+
			"please convert the .p12 manually:\n"+
			"  openssl pkcs12 -in cert.p12 -nocerts -nodes -out key.pem\n"+
			"  openssl pkcs12 -in cert.p12 -nokeys -out cert.pem\n"+
			"Then use NewP12SignerFromPEM(keyPEM, certPEM)")
	}

	keyPEM, err := runOpenSSL("pkcs12",
		"-in", p12Path,
		"-nocerts", "-nodes",
		"-passin", "pass:"+password,
	)
	if err != nil {
		return nil, fmt.Errorf("signing: extracting private key from .p12: %w", err)
	}

	certPEM, err := runOpenSSL("pkcs12",
		"-in", p12Path,
		"-nokeys", "-clcerts",
		"-passin", "pass:"+password,
	)
	if err != nil {
		return nil, fmt.Errorf("signing: extracting certificate from .p12: %w", err)
	}

	return NewP12SignerFromPEM([]byte(keyPEM), []byte(certPEM))
}

// LoadTLSCertFromP12 extracts a tls.Certificate from a .p12/PFX file using OpenSSL.
// The returned certificate is suitable for use in TLS server configurations.
func LoadTLSCertFromP12(p12Path, password string) (*tls.Certificate, error) {
	keyPEM, err := runOpenSSL("pkcs12",
		"-in", p12Path,
		"-nocerts", "-nodes",
		"-passin", "pass:"+password,
	)
	if err != nil {
		return nil, fmt.Errorf("signing: TLS key from .p12: %w", err)
	}

	certPEM, err := runOpenSSL("pkcs12",
		"-in", p12Path,
		"-nokeys", "-clcerts",
		"-passin", "pass:"+password,
	)
	if err != nil {
		return nil, fmt.Errorf("signing: TLS cert from .p12: %w", err)
	}

	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("signing: creating TLS key pair: %w", err)
	}
	return &cert, nil
}

// OpenSSLAvailable reports whether the openssl executable is found in the system PATH.
func OpenSSLAvailable() bool {
	_, err := exec.LookPath("openssl")
	return err == nil
}

// runOpenSSL executes an openssl command with the given arguments and returns its stdout.
// On failure it wraps the error with any stderr output for debugging.
func runOpenSSL(args ...string) (string, error) {
	out, err := exec.Command("openssl", args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("openssl %s: %w\nstderr: %s",
				args[0], err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("openssl %s: %w", args[0], err)
	}
	return string(out), nil
}
