package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// The kit files end in .tmpl, so that sisyphus (and Obsidian) never mistake the kit for a real issues/TEMPLATE.md
// in whatever repo holds this source.
// Everything else a project needs (CONTRIBUTING.md, AGENTS.md, versioning, CI, GitHub Actions, and so on) is a
// separate concern of a project-scaffolding template, not of sisyphus: init only sets up issue tracking.
//
//go:embed all:kit
var kit embed.FS

// initRepo writes issues/ (TEMPLATE.md and the open/in-progress/closed directories) into dir. It skips files that
// exist, unless force is set.
func initRepo(dir string, force, nested bool, out, warnings io.Writer) error {
	if above := trackerAbove(dir); above != "" && !nested {
		return fmt.Errorf("%s already has an issue tracker (issues/TEMPLATE.md). A tracker in %s would hide it from every command that runs there. Run sisyphus below %s, or use --nested to make a second tracker on purpose.", above, dir, above)
	}
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
		content, err := kit.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
		fmt.Fprintln(out, relative)
		return nil
	})
}

// trackerAbove returns the nearest directory above dir that has issues/TEMPLATE.md, or "" when there
// is none. findRoot stops at the nearest such directory, so a tracker in dir would hide that one.
func trackerAbove(dir string) string {
	for current := filepath.Dir(dir); ; current = filepath.Dir(current) {
		if fileExists(filepath.Join(current, "issues", "TEMPLATE.md")) {
			return current
		}
		if filepath.Dir(current) == current {
			return ""
		}
	}
}
