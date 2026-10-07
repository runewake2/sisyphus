package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// The kit files end in .tmpl, so that their copies of AGENTS.md and other files do not take part in the wikilinks of this repo.
//
//go:embed all:kit
var kit embed.FS

type kitValues struct {
	Project        string
	BookmarkPrefix string
	Date           string
}

// initRepo writes the workflow kit into dir. It skips files that exist, unless force is set.
func initRepo(dir string, values kitValues, force bool, out, warnings io.Writer) error {
	return fs.WalkDir(kit, "kit", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative := strings.TrimSuffix(strings.TrimPrefix(name, "kit/"), ".tmpl")
		target := filepath.Join(dir, filepath.FromSlash(relative))
		if fileExists(target) && !force {
			warn(warnings, fmt.Sprintf("%s exists, so init did not change it. Use --force to replace it.", relative))
			return nil
		}
		source, err := kit.ReadFile(name)
		if err != nil {
			return err
		}
		tmpl, err := template.New(relative).Delims("{%", "%}").Parse(string(source))
		if err != nil {
			return err
		}
		var content bytes.Buffer
		if err := tmpl.Execute(&content, values); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content.Bytes(), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(out, relative)
		return nil
	})
}
