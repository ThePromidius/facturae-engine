// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

// Command facturae-engine is the sidecar server entry point for FacturaE invoice
// processing, XML signing, Verifactu chain hashing, and optional automatic
// submission to AEAT or FACe.
//
// Usage:
//
//	facturae-engine [flags]
//
// Flags:
//
//	-socket   Address to listen on (default: 127.0.0.1:8080). Use a path for Unix sockets.
//	-schemas  Directory for cached XSD files (default: ./schemas)
//	-p12      Path to PKCS#12 certificate (.p12/.pfx) for real signing and AEAT TLS
//	-p12pass  Password for the PKCS#12 file
//	-dev-p12  Path to a development PKCS#12 file. If the file exists it is
//	          loaded; otherwise a self-signed 2048-bit RSA certificate is
//	          generated and saved at this path for reuse.
//	-key      Path to PEM private key (alternative to -p12)
//	-cert     Path to PEM certificate (alternative to -p12)
//	-aeat     AEAT Environment (test|prod) - enables automatic submission
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThePromidius/facturae-engine/src/internal/aeat"
	"github.com/ThePromidius/facturae-engine/src/internal/api"
	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/signing"
)

// main parses CLI flags, initialises the signer (from a PKCS#12 file, PEM files,
// or a mock), optionally configures the AEAT client, selects the chain store
// backend, and starts the HTTP server.
func main() {
	socketPath := flag.String("socket", getEnv("ENGINE_SOCKET", defaultSocketPath()), "Listen address (e.g. 127.0.0.1:8080)")
	schemaDir := flag.String("schemas", getEnv("ENGINE_SCHEMAS", "./schemas"), "Directory for cached XSD files")
	p12Path := flag.String("p12", getEnv("CERT_P12_PATH", ""), "Path to PKCS#12 file (.p12/.pfx)")
	p12Pass := flag.String("p12pass", getEnv("CERT_P12_PASS", ""), "Password for PKCS#12 file")
	devP12 := flag.String("dev-p12", getEnv("DEV_P12_PATH", ""), "Development PKCS#12 path (auto-generated if missing)")
	devP12Pass := flag.String("dev-p12pass", getEnv("DEV_P12_PASS", "changeit"), "Password for dev PKCS#12 file")
	keyPath := flag.String("key", getEnv("CERT_KEY_PATH", ""), "Path to PEM private key")
	certPath := flag.String("cert", getEnv("CERT_PEM_PATH", ""), "Path to PEM certificate")
	aeatEnv := flag.String("aeat", getEnv("AEAT_ENV", ""), "AEAT Environment (test|prod) to enable submission")
	dbDriver := flag.String("db", getEnv("DB_DRIVER", "memory"), "Database driver (memory|postgres|sqlite)")
	dbDSN := flag.String("dsn", getEnv("DB_DSN", ""), "Database DSN (connection string)")
	flag.Parse()

	var signer signing.Signer
	var aeatClient *aeat.Client

	if *p12Path != "" {
		if !signing.OpenSSLAvailable() {
			log.Fatalf("OpenSSL is required to parse .p12 files directly. Either install OpenSSL or pre-convert your cert to PEM.")
		}
		var err error
		signer, err = signing.LoadFromP12File(*p12Path, *p12Pass)
		if err != nil {
			log.Fatalf("Loading .p12: %v", err)
		}
		fmt.Printf("Firma real activada (PKCS#12): %s\n", signer.Algorithm())

		if *aeatEnv != "" {
			tlsCert, err := signing.LoadTLSCertFromP12(*p12Path, *p12Pass)
			if err != nil {
				log.Fatalf("Loading TLS cert from .p12 for AEAT: %v", err)
			}
			aeatClient = aeat.NewClient(aeat.Environment(*aeatEnv), tlsCert)
			fmt.Printf("Envio automatico AEAT activado (Entorno: %s, con certificado TLS)\n", *aeatEnv)
		}

	} else if *devP12 != "" {
		var err error
		signer, err = signing.GenerateSelfSignedP12(*devP12, *devP12Pass)
		if err != nil {
			log.Fatalf("Dev PKCS#12 setup: %v", err)
		}
		fmt.Printf("Firma desarrollo activada (dev-p12: %s): %s\n", *devP12, signer.Algorithm())

		if *aeatEnv != "" {
			fmt.Printf("AVISO: modo dev-p12 no apto para AEAT real (certificado autofirmado). Solo pruebas.\n")
			aeatClient = aeat.NewClient(aeat.Environment(*aeatEnv), nil)
		}

	} else if *keyPath != "" && *certPath != "" {
		keyPEM, err := os.ReadFile(*keyPath)
		if err != nil {
			log.Fatalf("reading key file %q: %v", *keyPath, err)
		}
		certPEM, err := os.ReadFile(*certPath)
		if err != nil {
			log.Fatalf("reading cert file %q: %v", *certPath, err)
		}
		signer, err = signing.NewP12SignerFromPEM(keyPEM, certPEM)
		if err != nil {
			log.Fatalf("initialising P12 signer: %v", err)
		}
		fmt.Printf("Firma real activada (PEM): %s\n", signer.Algorithm())

		if *aeatEnv != "" {
			aeatClient = aeat.NewClient(aeat.Environment(*aeatEnv), nil)
			fmt.Printf("Envio automatico AEAT activado (Entorno: %s, SIN certificado TLS configurado - fallara en Prod)\n", *aeatEnv)
		}

	} else {
		signer = signing.MockSigner{}
		fmt.Println("Modo desarrollo: firma simulada (MockSigner)")
		fmt.Println("  Para firma real: -p12 <cert.p12> -p12pass <pass> (o usar PEM)")

		if *aeatEnv != "" {
			aeatClient = aeat.NewClient(aeat.Environment(*aeatEnv), nil)
			fmt.Printf("Envio automatico AEAT activado (Entorno: %s, Modo simulacion - fallara sin firma valida)\n", *aeatEnv)
		}
	}

	var cs chain.Store
	if *dbDriver != "memory" {
		var err error
		cs, err = chain.NewSQLStore(*dbDriver, *dbDSN)
		if err != nil {
			log.Fatalf("Error inicializando base de datos (%s): %v", *dbDriver, err)
		}
		fmt.Printf("Persistencia SQL activada (%s)\n", *dbDriver)
	}

	srv, err := api.New(api.Config{
		SocketPath: *socketPath,
		SchemaDir:  *schemaDir,
		Signer:     signer,
		AEATClient: aeatClient,
		ChainStore: cs,
	})

	if err != nil {
		log.Fatalf("creating server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nSenal recibida - apagando servidor...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("error durante el apagado: %v", err)
		}
		os.Remove(*socketPath)
	}()

	printBanner(*socketPath, signer.Algorithm())

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Servidor detenido: %v\n", err)
	}
}

// defaultSocketPath returns the default listen address ("127.0.0.1:8080").
func defaultSocketPath() string {
	return "127.0.0.1:8080"
}

// getEnv looks up an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// printBanner prints the startup banner showing the socket address, signing
// algorithm, and available endpoints.
func printBanner(socket, algo string) {
	fmt.Println()
	fmt.Println("============================================")
	fmt.Println("  FACTURAE ENGINE v2  -  Sidecar")
	fmt.Println("============================================")
	fmt.Printf("  Socket  : %s\n", socket)
	fmt.Printf("  Firma   : %s\n", truncate(algo, 37))
	fmt.Println("--------------------------------------------")
	fmt.Println("  POST /invoice   ->  JSON -> XML firmado")
	fmt.Println("  GET  /health    ->  Estado del engine")
	fmt.Println("  GET  /chain     ->  Ultimo registro Verifactu")
	fmt.Println("============================================")
	fmt.Println()
}

// truncate shortens s to at most limit characters, appending "..." if truncated.
func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}
