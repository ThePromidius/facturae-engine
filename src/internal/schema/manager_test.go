// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package schema_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ThePromidius/facturae-engine/src/internal/schema"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "schema-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNewManager_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "schema-test-newdir")
	defer os.RemoveAll(dir)

	_, err := schema.NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("NewManager should create the cache directory")
	}
}

func TestManager_IsCached_FalseWhenEmpty(t *testing.T) {
	m, _ := schema.NewManager(tempDir(t))
	if m.IsCached("3.2.2") {
		t.Error("should not be cached on a fresh manager")
	}
}

func TestManager_SchemaPath_DownloadsAndCaches(t *testing.T) {
	fakeXSD := []byte(`<?xml version="1.0"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(fakeXSD)
	}))
	defer srv.Close()

	schema.KnownVersions["test-1.0"] = srv.URL + "/schema.xsd"
	defer delete(schema.KnownVersions, "test-1.0")

	dir := tempDir(t)
	m, _ := schema.NewManager(dir)

	path, err := m.SchemaPath("test-1.0")
	if err != nil {
		t.Fatalf("SchemaPath: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected schema file to exist after download")
	}

	data, _ := os.ReadFile(path)
	if string(data) != string(fakeXSD) {
		t.Error("downloaded content mismatch")
	}
}

func TestManager_SchemaPath_ReturnsCachedOnSecondCall(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte(`<xs:schema/>`))
	}))
	defer srv.Close()

	schema.KnownVersions["test-cache"] = srv.URL
	defer delete(schema.KnownVersions, "test-cache")

	dir := tempDir(t)
	m, _ := schema.NewManager(dir)

	m.SchemaPath("test-cache")
	m.SchemaPath("test-cache")

	if callCount > 1 {
		t.Errorf("expected 1 HTTP call, got %d (caching not working)", callCount)
	}
}

func TestManager_SchemaPath_UnknownVersion(t *testing.T) {
	m, _ := schema.NewManager(tempDir(t))
	_, err := m.SchemaPath("99.9.9")
	if err == nil {
		t.Error("expected error for unknown version")
	}
}

func TestManager_SchemaPath_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	schema.KnownVersions["test-fail"] = srv.URL
	defer delete(schema.KnownVersions, "test-fail")

	m, _ := schema.NewManager(tempDir(t))
	_, err := m.SchemaPath("test-fail")
	if err == nil {
		t.Error("expected error for server 503 response")
	}
}

func TestManager_ClearCache(t *testing.T) {
	fakeXSD := []byte(`<xs:schema/>`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(fakeXSD)
	}))
	defer srv.Close()

	schema.KnownVersions["test-clear"] = srv.URL
	defer delete(schema.KnownVersions, "test-clear")

	dir := tempDir(t)
	m, _ := schema.NewManager(dir)
	m.SchemaPath("test-clear")

	if !m.IsCached("test-clear") {
		t.Fatal("should be cached before clear")
	}
	m.ClearCache()
	if m.IsCached("test-clear") {
		t.Error("should not be cached after ClearCache()")
	}
}
