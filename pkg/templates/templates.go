package templates

import (
	"context"
	"embed"
	"io"
	"log"
	"strings"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/function-builder/pkg/types"
)

const Suffix = ".acorn"

var (
	//go:embed *.acorn
	templatesFS embed.FS
	templates   []types.Template
)

func Decode(ctx context.Context, in io.Reader, name string, args map[string]any) (*types.Template, error) {
	var template types.Template

	err := aml.NewDecoder(in, aml.DecoderOption{
		Context:     ctx,
		Args:        args,
		SourceName:  name,
		SchemaValue: types.Schema,
	}).Decode(&template)
	if err != nil {
		return nil, err
	}

	if template.Name == "" {
		template.Name = strings.TrimSuffix(name, Suffix)
	}

	return &template, nil
}

func init() {
	entries, err := templatesFS.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		f, err := templatesFS.Open(entry.Name())
		if err != nil {
			panic(err)
		}

		template, err := Decode(context.Background(), f, entry.Name(), nil)
		if err != nil {
			panic(err)
		}

		_ = f.Close()
		templates = append(templates, *template)
	}
}

func List() []types.Template {
	return templates
}
