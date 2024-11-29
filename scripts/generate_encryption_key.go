package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	keyFile = "../.encryption_key"
	keySize = 32 // AES-256
)

func main() {
	// Get absolute path for key file
	absKeyFile, err := filepath.Abs(keyFile)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if key file already exists
	if _, err := os.Stat(absKeyFile); err == nil {
		fmt.Printf("Error: Key file already exists at %s\n", absKeyFile)
		fmt.Println("Remove the existing key file first if you want to generate a new one.")
		os.Exit(1)
	}

	// Generate random key
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		os.Exit(1)
	}

	// Save the key
	if err := os.WriteFile(absKeyFile, key, 0600); err != nil {
		fmt.Printf("Error writing key file: %v\n", err)
		os.Exit(1)
	}

	// Print the key in hex format for reference
	fmt.Printf("Generated encryption key (hex): %s\n", hex.EncodeToString(key))
	fmt.Printf("Key file created: %s\n", absKeyFile)
	fmt.Printf("File permissions: 0600 (read/write for owner only)\n")
}
