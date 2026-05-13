// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package schema manages FacturaE XSD schema files: it caches them locally and
// provides validation of XML documents against the cached schemas via xmllint.
package schema

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KnownVersions maps FacturaE version strings to their remote XSD schema URLs.
var KnownVersions = map[string]string{
	"3.2.2": "https://www.facturae.gob.es/content/dam/facturae/formato/versiones/Facturaev3_2_2.xml",
	"3.2.1": "https://www.facturae.gob.es/content/dam/facturae/formato/versiones/Facturaev3_2_1.xml",
}

// Manager caches FacturaE XSD schemas on disk and provides access to their
// local file paths.
type Manager struct {
	cacheDir string
	client   *http.Client
	mu       sync.Mutex
}

// NewManager creates a Manager with the given cache directory, creating it if
// it does not exist.
func NewManager(cacheDir string) (*Manager, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("schema: creating cache dir %q: %w", cacheDir, err)
	}
	return &Manager{
		cacheDir: cacheDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SchemaPath returns the local file path for the given FacturaE version's XSD
// schema. If the file is not already cached, it downloads it from the remote
// URL defined in KnownVersions.
func (m *Manager) SchemaPath(version string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	localPath := filepath.Join(m.cacheDir, "facturae_"+version+".xsd")

	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	url, ok := KnownVersions[version]
	if !ok {
		return "", fmt.Errorf("schema: unknown FacturaE version %q", version)
	}

	fmt.Printf("[schema] Descargando esquema FacturaE %s desde %s...\n", version, url)
	if err := m.download(url, localPath); err != nil {
		return "", fmt.Errorf("schema: downloading %s: %w", version, err)
	}

	fmt.Printf("[schema] Esquema %s guardado en %s\n", version, localPath)
	return localPath, nil
}

// IsCached reports whether the XSD schema for the given version is already
// present in the local cache directory.
func (m *Manager) IsCached(version string) bool {
	path := filepath.Join(m.cacheDir, "facturae_"+version+".xsd")
	_, err := os.Stat(path)
	return err == nil
}

// ClearCache removes all .xsd files from the cache directory.
func (m *Manager) ClearCache() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries, err := os.ReadDir(m.cacheDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".xsd" {
			os.Remove(filepath.Join(m.cacheDir, e.Name()))
		}
	}
	return nil
}

// download fetches the XSD file from url and writes it atomically to dest via
// a temporary file and rename.
func (m *Manager) download(url, dest string) error {
	resp, err := m.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(m.cacheDir, "*.xsd.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	_, err = io.Copy(tmp, resp.Body)
	tmp.Close()
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, dest)
}
