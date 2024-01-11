package types

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

const DefaultTarget = "default"

type wrappedWriter struct {
	out io.Writer
	err error
}

func (w *wrappedWriter) Write(s string) {
	if w.err != nil {
		return
	}
	_, w.err = w.out.Write([]byte(s))
}

func mergeDockerIgnore(target, content string) error {
	existingLines := map[string]bool{}

	existing, err := os.ReadFile(target)
	if err != nil {
		return err
	}

	for _, existingLine := range strings.Split(string(existing), "\n") {
		existingLines[strings.TrimSpace(strings.TrimPrefix(existingLine, "#"))] = true
	}

	var missing []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !existingLines[line] {
			missing = append(missing)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	f, err := os.Open(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte("\n" + strings.Join(missing, "\n")))
	return err
}

func (b *Template) WriteLocalFiles(cwd string) error {
	for file, content := range b.LocalFiles {
		target := filepath.Join(cwd, file)
		if _, err := os.Stat(target); err == nil {
			if filepath.Base(file) == ".dockerignore" {
				if err := mergeDockerIgnore(target, content); err != nil {
					return err
				}
			}
			// exists
			continue
		} else if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(target, []byte(content), 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (b *Template) Process(cwd string) ([]byte, error) {
	if err := b.WriteLocalFiles(cwd); err != nil {
		return nil, err
	}
	return b.ToDockerfile()
}

func (b *Template) ToDockerfile() ([]byte, error) {
	out := &bytes.Buffer{}
	w := &wrappedWriter{
		out: out,
	}
	for i, key := range b.order() {
		if i > 0 {
			w.Write("\n")
		}
		t := b.Targets[key]
		if err := t.toDockerfile(key, w); err != nil {
			return nil, err
		}
	}

	result := out.Bytes()
	logrus.Debugf("Generated dockerfile: %s", result)
	return result, w.err
}

func (b *Template) order() []string {
	var (
		hasDefault bool
		keys       []string
	)

	for key := range b.Targets {
		if key == DefaultTarget {
			hasDefault = true
		} else {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)
	if hasDefault {
		keys = append(keys, DefaultTarget)
	}

	return keys
}

func writeJSON(w *wrappedWriter, fmtString string, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	w.Write(fmt.Sprintf(fmtString, string(data)))
	return nil
}

func (b *Build) toDockerfile(key string, w *wrappedWriter) error {
	if b.From == "" {
		return nil
	}

	w.Write(fmt.Sprintf("FROM %s AS %s\n", b.From, key))

	if b.Acornfile != "" {
		b64 := base64.StdEncoding.EncodeToString([]byte(b.Acornfile))
		w.Write(fmt.Sprintf("LABEL io.acorn.acornfile.fragment=\"%s\"\n", b64))
	}

	if len(b.Shell) > 0 {
		if err := writeJSON(w, "SHELL %s\n", b.Shell); err != nil {
			return err
		}
	}

	if b.StopSignal != "" {
		w.Write(fmt.Sprintf("STOPSIGNAL %s\n", b.StopSignal))
	}

	if len(b.CacheVolumes) > 0 {
		if err := writeJSON(w, "VOLUME %s\n", b.CacheVolumes); err != nil {
			return err
		}
	}

	if b.Workdir != "" {
		w.Write(fmt.Sprintf("WORKDIR %s\n", b.Workdir))
	}

	for _, env := range b.Env {
		w.Write(fmt.Sprintf("ENV %s\n", env))
	}

	for _, v := range b.Run {
		if err := v.toDockerfile(w, b.CacheVolumes); err != nil {
			return err
		}
	}

	if b.User != nil {
		w.Write(fmt.Sprintf("USER %d\n", *b.User))
	}

	if len(b.Entrypoint) > 0 {
		if err := writeJSON(w, "ENTRYPOINT %s\n", b.Entrypoint); err != nil {
			return err
		}
	}

	return nil
}

func (i *Instruction) toDockerfile(w *wrappedWriter, cacheVolumes []string) error {
	if i.Run.Command != "" {
		w.Write("RUN ")
		for _, m := range i.Run.Mount {
			if m.Cache != nil {
				w.Write("--mount=type=cache,target=" + m.Cache.Target + " ")
			}
		}
		for _, cacheVolume := range cacheVolumes {
			w.Write("--mount=type=cache,target=" + cacheVolume + " ")
		}
		w.Write("<<EOT\n")
		w.Write(i.Run.Command)
		w.Write("\nEOT\n")
	}
	if i.Copy != nil && i.Copy.Dest != "" {
		w.Write("COPY ")
		if i.Copy.Link {
			w.Write("--link ")
		}
		if i.Copy.From != "" {
			w.Write("--from=" + i.Copy.From + " ")
		}
		if i.Copy.Source == "" {
			w.Write(". ")
		} else {
			w.Write(i.Copy.Source + " ")
		}
		w.Write(i.Copy.Dest)
		w.Write("\n")
	}
	if i.Volume.Volume != "" {
		if err := writeJSON(w, "VOLUME %s\n", []string{i.Volume.Volume}); err != nil {
			return err
		}
	}
	if i.Workdir.Workdir != "" {
		w.Write(fmt.Sprintf("WORKDIR %s\n", i.Workdir.Workdir))
	}
	return nil
}
