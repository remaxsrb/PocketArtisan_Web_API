// Command seedgen emits the reference-data seed (crafts, product categories,
// and their links) as a SQL migration body on stdout.
//
// The data and generation logic live in the config package (config.BuildSeedSQL)
// so there is a single, authoritative copy of the reference data and its
// `search_keywords` normalization. When the reference data changes, regenerate a
// NEW versioned migration:
//
//	go run ./cmd/seedgen > migrations/<version>_seed_reference_data.sql
//	atlas migrate hash --dir file://migrations
package main

import (
	"fmt"
	"os"

	"PocketArtisan/config"
)

func main() {
	if _, err := os.Stdout.WriteString(config.BuildSeedSQL()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write seed sql: %v\n", err)
		os.Exit(1)
	}
}
