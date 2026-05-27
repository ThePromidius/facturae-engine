package signing

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// GenerateSelfSignedP12 creates a 2048-bit RSA self-signed certificate,
// packages it into a PKCS#12 file at the given path, and returns a P12Signer
// ready to use.  If the file already exists it is loaded directly.
//
// OpenSSL is preferred when available; otherwise it falls back to Docker
// running the alpine/openssl image.
func GenerateSelfSignedP12(path, password string) (*P12Signer, error) {
	if _, err := os.Stat(path); err == nil {
		return LoadFromP12File(path, password)
	}

	key, cert, err := generateKeyCert()
	if err != nil {
		return nil, err
	}

	workDir := filepath.Dir(path)
	if workDir == "" {
		workDir = "."
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	keyPath := filepath.Join(workDir, ".dev-key.pem")
	certPath := filepath.Join(workDir, ".dev-cert.pem")
	defer os.Remove(keyPath)
	defer os.Remove(certPath)

	certDER, _ := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err := writePEM(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)); err != nil {
		return nil, err
	}
	if err := writePEM(certPath, "CERTIFICATE", certDER); err != nil {
		return nil, err
	}

	switch {
	case OpenSSLAvailable():
		if err := runOpenSSLExport(keyPath, certPath, path, password); err != nil {
			return nil, err
		}
	case dockerAvailable():
		if err := runDockerOpenSSLExport(keyPath, certPath, path, password); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("neither OpenSSL nor Docker found. Install one, or use -p12 with a pre-generated cert")
	}

	return &P12Signer{privateKey: key, certificate: cert}, nil
}

func generateKeyCert() (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generating RSA key: %w", err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "VIVE-X Sidecar Development",
			Organization: []string{"VIVE-X Sistemas y Servicios"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("creating certificate: %w", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing generated certificate: %w", err)
	}
	return key, cert, nil
}

func writePEM(path, blockType string, derBytes []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: derBytes})
}

func runOpenSSLExport(keyPath, certPath, outPath, password string) error {
	args := []string{
		"pkcs12", "-export",
		"-in", certPath,
		"-inkey", keyPath,
		"-out", outPath,
		"-passout", "pass:" + password,
	}
	cmd := exec.Command("openssl", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("openssl pkcs12 failed: %w\n%s", err, string(out))
	}
	return nil
}

func runDockerOpenSSLExport(keyPath, certPath, outPath, password string) error {
	absKey, _ := filepath.Abs(keyPath)
	absCert, _ := filepath.Abs(certPath)
	absOut, _ := filepath.Abs(outPath)
	mount := filepath.Dir(absOut)

	args := []string{
		"run", "--rm",
		"-v", mount + ":/out",
		"alpine/openssl",
		"pkcs12", "-export",
		"-in", "/out/" + filepath.Base(absCert),
		"-inkey", "/out/" + filepath.Base(absKey),
		"-out", "/out/" + filepath.Base(absOut),
		"-passout", "pass:" + password,
	}
	cmd := exec.Command("docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker alpine/openssl failed: %w\n%s", err, string(out))
	}
	return nil
}

func dockerAvailable() bool {
	return exec.Command("docker", "--version").Run() == nil
}
