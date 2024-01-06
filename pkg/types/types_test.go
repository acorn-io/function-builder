package types

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/acorn-io/aml"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	entries, err := fs.Glob(os.DirFS("testdata/TestUnmarshal"), "*.acorn")
	require.NoError(t, err)

	templateEntries, err := fs.Glob(os.DirFS("testdata/TestUnmarshal"), "templates/*.acorn")
	require.NoError(t, err)

	for _, entry := range append(entries, templateEntries...) {
		t.Run(entry, func(t *testing.T) {
			p := filepath.Join("testdata/TestUnmarshal", entry)
			f, err := aml.Open(p)
			require.NoError(t, err)

			var build Template
			err = aml.NewDecoder(f, aml.DecoderOption{
				SourceName:  p,
				SchemaValue: Schema,
			}).Decode(&build)
			require.NoError(t, err)

			out, err := json.MarshalIndent(build, "", "  ")
			require.NoError(t, err)

			// ensure out matches schema
			err = aml.NewDecoder(bytes.NewReader(out), aml.DecoderOption{
				SourceName:  p,
				SchemaValue: Schema,
			}).Decode(&build)
			require.NoError(t, err)

			autogold.ExpectFile(t, autogold.Raw(out))
		})
	}
}
