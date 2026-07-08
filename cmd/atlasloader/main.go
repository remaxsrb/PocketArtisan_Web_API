// Command atlasloader prints the DDL for all GORM entities so that Atlas can
// diff the desired schema (the Go models) against the current migration state.
//
// It is NOT part of the server binary: `go build ./cmd/` only builds the
// server package. Atlas invokes this via the `external_schema` data source in
// atlas.hcl (see the `program` list there).
package main

import (
	"fmt"
	"io"
	"os"

	"PocketArtisan/internal/entities"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&entities.Cart{},
		&entities.CartItem{},
		&entities.User{},
		&entities.Craft{},
		&entities.Craftsman{},
		&entities.CraftsmanApplication{},
		&entities.ProductCategory{},
		&entities.CraftProductCategory{},
		&entities.Product{},
		&entities.ProductImage{},
		&entities.ProductVideo{},
		&entities.Order{},
		&entities.OrderItem{},
		&entities.CraftsmanRatingRecord{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.WriteString(os.Stdout, stmts); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write schema: %v\n", err)
		os.Exit(1)
	}
}
