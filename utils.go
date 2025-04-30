package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// getEnvVar retrieves an environment variable or exits if not found.
func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Missing environment variable %s", key)
	}
	log.Printf("Using environment variable %s=%s", key, value)
	return value
}

func writeFile(outDir, fileName string, data []byte) error {
	file := filepath.Join(outDir, fileName)
	log.Printf("Writing %s", file)
	err := os.WriteFile(file, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", file, err)
	}
	return nil
}

func writeJSONFile[T any](outDir, fileName string, data T) error {
	json, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal for %s: %w", fileName, err)
	}
	return writeFile(outDir, fileName, json)
}
