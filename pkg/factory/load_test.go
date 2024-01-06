package templates

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/acorn-io/function-builder/pkg/templates"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	for _, template := range templates.List() {
		t.Run(template.Name, func(t *testing.T) {
			content, err := Load(context.Background(), filepath.Join("./testdata/"+template.Name, "build.acorn"))
			require.NoError(t, err)
			autogold.ExpectFile(t, autogold.Raw(content))
		})
	}
}
