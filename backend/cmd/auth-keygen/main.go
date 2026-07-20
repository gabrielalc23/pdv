package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := flag.String("dir", ".", "output directory for keys")
	kid := flag.String("kid", "", "key ID (used as filename prefix)")
	force := flag.Bool("force", false, "overwrite existing files")
	flag.Parse()

	if *kid == "" {
		fmt.Fprintln(os.Stderr, "error: --kid is required")
		os.Exit(1)
	}

	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create directory %q: %v\n", *dir, err)
		os.Exit(1)
	}

	pubPath := filepath.Join(*dir, *kid+".pem")
	privPath := filepath.Join(*dir, *kid+".priv.pem")

	if !*force {
		if _, err := os.Stat(pubPath); err == nil {
			fmt.Fprintf(os.Stderr, "error: %q already exists (use --force to overwrite)\n", pubPath)
			os.Exit(1)
		}
		if _, err := os.Stat(privPath); err == nil {
			fmt.Fprintf(os.Stderr, "error: %q already exists (use --force to overwrite)\n", privPath)
			os.Exit(1)
		}
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to generate key: %v\n", err)
		os.Exit(1)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to marshal private key: %v\n", err)
		os.Exit(1)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to marshal public key: %v\n", err)
		os.Exit(1)
	}

	privFile, err := os.OpenFile(privPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create private key file: %v\n", err)
		os.Exit(1)
	}
	defer privFile.Close()

	if err := pem.Encode(privFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write private key PEM: %v\n", err)
		os.Exit(1)
	}
	privFile.Close()

	pubFile, err := os.Create(pubPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create public key file: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()

	if err := pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write public key PEM: %v\n", err)
		os.Exit(1)
	}
	pubFile.Close()

	fmt.Printf("Generated Ed25519 key pair with kid=%q\n", *kid)
	fmt.Printf("  Private: %s\n", privPath)
	fmt.Printf("  Public:  %s\n", pubPath)
}
