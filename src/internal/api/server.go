// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Package api provides the HTTP server that exposes FacturaE invoice processing
// endpoints: /invoice (POST), /chain (GET), /health (GET), and /qr (GET). It
// wires together signing, schema validation, Verifactu chain hashing, and
// optional AEAT submission.
package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/aeat"
	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/schema"
	"github.com/ThePromidius/facturae-engine/src/internal/signing"
)

// Config holds the dependencies and settings required to create a new Server.
type Config struct {
	SocketPath string
	SchemaDir  string
	Signer     signing.Signer
	AEATClient *aeat.Client
	ChainStore chain.Store
}

// Server is an HTTP server that handles FacturaE invoice processing, chain
// queries, health checks, and QR generation.
type Server struct {
	cfg     Config
	chain   *chain.Chain
	schemas *schema.Manager
	signer  signing.Signer
	aeat    *aeat.Client
	http    *http.Server
}

// New creates a new Server with the given configuration. It initialises the
// schema manager, the Verifactu chain, and registers HTTP routes. If no
// ChainStore is supplied, it defaults to an in-memory store.
func New(cfg Config) (*Server, error) {
	sm, err := schema.NewManager(cfg.SchemaDir)
	if err != nil {
		return nil, fmt.Errorf("server: initialising schema manager: %w", err)
	}

	signer := cfg.Signer
	if signer == nil {
		signer = signing.MockSigner{}
		fmt.Println("[server] No signer configured - using MockSigner (development mode)")
	}

	cs := cfg.ChainStore
	if cs == nil {
		cs = chain.NewMemoryStore()
	}

	s := &Server{
		cfg:     cfg,
		chain:   chain.NewChain(cs),
		schemas: sm,
		signer:  signer,
		aeat:    cfg.AEATClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/invoice", s.handleInvoice)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/chain", s.handleChain)
	mux.HandleFunc("/chain/verify", s.handleChainVerify)
	mux.HandleFunc("/qr", s.handleQR)

	s.http = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s, nil
}

// ListenAndServe starts the HTTP server on the configured socket path.
// It detects whether to use TCP (if the address contains a colon) or Unix
// domain sockets (otherwise).
func (s *Server) ListenAndServe() error {
	network := "unix"
	addr := s.cfg.SocketPath
	if strings.Contains(addr, ":") {
		network = "tcp"
	} else {
		if _, err := os.Stat(addr); err == nil {
			os.Remove(addr)
		}
	}

	ln, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("server: listening on %s %q: %w", network, addr, err)
	}

	fmt.Printf("[server] Escuchando en %s (modo: %s)\n", addr, s.signer.Algorithm())
	return s.http.Serve(ln)
}

// Shutdown gracefully shuts down the HTTP server, waiting for active
// connections to complete.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// ServeHTTP delegates to the internal HTTP server's handler mux.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.http.Handler.ServeHTTP(w, r)
}
