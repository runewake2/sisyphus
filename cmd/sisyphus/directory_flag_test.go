package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// runElsewhere runs a command with a findRoot that fails, so that only -C can find the repo.
func runElsewhere(args ...string) result {
	var output, stderr bytes.Buffer
	exit := run(args, func() (string, error) { return "", errors.New(rootNotFound) }, &output, &stderr)
	return result{exit: exit, output: output.String(), error: stderr.String()}
}

func TestDirectoryFlagFindsTheRepoFromAnAbsoluteOrRelativePath(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativePath, err := filepath.Rel(wd, filepath.Join(r.root, "issues"))
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range []string{r.root, relativePath} {
		res := runElsewhere("-C", dir, "list")
		equal(t, 0, res.exit)
		contains(t, res.output, "fix-login-bug")
	}
	res := runElsewhere("-C", filepath.Join(r.root, "README.md"), "list")
	equal(t, 1, res.exit)
	contains(t, res.error, "is not a directory")
}

func TestInitUsesTheDirectoryFlag(t *testing.T) {
	dir := t.TempDir()

	equal(t, 0, runElsewhere("-C", dir, "init").exit)
	isTrue(t, fileExists(filepath.Join(dir, "issues", "TEMPLATE.md")), "init writes into -C")
	equal(t, 0, runElsewhere("-C", dir, "init", "--dir", "sub", "--nested").exit)
	isTrue(t, fileExists(filepath.Join(dir, "sub", "issues", "TEMPLATE.md")), "a relative --dir is taken from -C")
}
