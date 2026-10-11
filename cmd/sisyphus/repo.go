package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// skippedDirectories are never walked when sisyphus walks the repo itself (see repoFiles).
var skippedDirectories = map[string]bool{
	".git": true, ".jj": true, "bin": true, "obj": true, "node_modules": true,
}

// rootNotFound is the error of every command that runs outside a repo. It tells a new user how to set
// one up.
const rootNotFound = "Cannot find the repo root. Run sisyphus in a directory below the one that contains issues/TEMPLATE.md. To set up issue tracking here, run: sisyphus init"

// findRoot searches up from the current directory first, so that each jj workspace uses its own files.
func findRoot() (string, error) {
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	return findRootFrom(starts...)
}

// findRootFrom searches up from each start directory in turn for the directory that contains
// issues/TEMPLATE.md.
func findRootFrom(starts ...string) (string, error) {
	for _, start := range starts {
		dir := start
		for {
			if fileExists(filepath.Join(dir, "issues", "TEMPLATE.md")) {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", errors.New(rootNotFound)
}

// relative uses "/" on every platform, because wikilinks use "/".
func relative(root, file string) string {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return filepath.ToSlash(file)
	}
	return filepath.ToSlash(rel)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func warn(warnings io.Writer, message string) {
	fmt.Fprintf(warnings, "Warning: %s\n", message)
}

func hasSuffixFold(s, suffix string) bool {
	return len(s) >= len(suffix) && strings.EqualFold(s[len(s)-len(suffix):], suffix)
}

func trimMarkdownExtension(name string) string {
	if hasSuffixFold(name, ".md") {
		return name[:len(name)-3]
	}
	return name
}
