package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var skippedDirectories = map[string]bool{
	".git": true, ".jj": true, "bin": true, "obj": true, "node_modules": true,
}

// findRoot searches up from the current directory first, so that each jj workspace uses its own files.
func findRoot() (string, error) {
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
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
	return "", errors.New("Cannot find the repo root. Run sisyphus in a directory below the one that contains issues/TEMPLATE.md.")
}

// version reads the VERSION file at the repo root, written by `sisyphus init`. It returns "" if the repo has no VERSION file yet.
func version() string {
	root, err := findRoot()
	if err != nil {
		return ""
	}
	content, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// relative uses "/" on every platform, because wikilinks use "/".
func relative(root, file string) string {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return filepath.ToSlash(file)
	}
	return filepath.ToSlash(rel)
}

func repoFiles(root string) []string {
	var files []string
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if path != root && skippedDirectories[strings.ToLower(entry.Name())] {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
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
