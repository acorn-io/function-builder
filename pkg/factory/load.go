package factory

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/function-builder/pkg/templates"
	"github.com/acorn-io/function-builder/pkg/types"
)

func isValidExt(templatePath string) bool {
	ext := strings.ToLower(filepath.Ext(templatePath))
	for _, i := range []string{".acorn", ".aml", ".yaml", ".yml", ".json"} {
		if i == ext {
			return true
		}
	}
	return false
}

func Load(ctx context.Context, templatePath string) ([]byte, error) {
	templateDir := filepath.Dir(templatePath)

	f, err := aml.Open(templatePath)
	if errors.Is(err, fs.ErrNotExist) {
		dockerfile, err := os.ReadFile(filepath.Join(templateDir, "Dockerfile"))
		if err == nil {
			return dockerfile, nil
		} else if errors.Is(err, fs.ErrNotExist) {
		} else {
			return nil, err
		}

		template, err := Detect(os.DirFS(templateDir))
		if err != nil {
			return nil, err
		}

		return template.Process(templateDir)
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	if isValidExt(templatePath) {
		template, err := templates.Decode(ctx, f, filepath.Base(templatePath), nil)
		if err != nil {
			return nil, err
		}
		return template.Process(templateDir)
	}

	return io.ReadAll(f)
}

type templateMatch struct {
	name     string
	priority int
}

func match(cwd fs.FS, detect types.DetectFile) (bool, error) {
	if len(detect.Files) == 0 {
		return false, nil
	}
	for _, pattern := range detect.Files {
		m, err := fs.Glob(cwd, pattern)
		if err != nil {
			return false, err
		}
		if len(m) == 0 {
			return false, nil
		}
	}
	return true, nil
}

func toPriority(i *int) int {
	if i == nil {
		return 100
	}
	return *i
}

type ErrNoTemplate error

func Detect(cwd fs.FS) (*types.Template, error) {
	var (
		matches []templateMatch
		byName  = map[string]types.Template{}
	)

	for _, template := range templates.List() {
		byName[template.Name] = template
		for _, detect := range template.Detect {
			if ok, err := match(cwd, detect); err != nil {
				return nil, err
			} else if ok {
				matches = append(matches, templateMatch{
					name:     template.Name,
					priority: toPriority(detect.Priority),
				})
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].priority == matches[j].priority {
			return matches[i].name < matches[j].name
		}
		return matches[i].priority < matches[j].priority
	})

	if len(matches) == 0 {
		return nil, (ErrNoTemplate)(fmt.Errorf("failed to analyze source and determine language runtime in directory: %s", cwd))
	}
	result := byName[matches[0].name]
	return &result, nil
}
