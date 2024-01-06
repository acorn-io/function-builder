package types

import (
	_ "embed"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/pkg/value"
)

var (
	Schema value.Schema

	//go:embed schema.acorn
	schemaFile []byte
)

func init() {
	err := aml.Unmarshal(schemaFile, &Schema, aml.DecoderOption{
		SourceName: "schema.acorn",
	})
	if err != nil {
		panic(err)
	}
}
