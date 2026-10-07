package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var kitFiles = []string{
	"issues/TEMPLATE.md",
	"issues/closed/.gitkeep",
	"issues/in-progress/.gitkeep",
	"issues/open/.gitkeep",
}

func runIn(t *testing.T, dir string, args ...string) result {
	t.Helper()
	var output, errors bytes.Buffer
	exit := run(args, func() (string, error) { return dir, nil }, &output, &errors)
	return result{exit: exit, output: output.String(), error: errors.String()}
}

func readIn(t *testing.T, dir, relative string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestInitWritesIssuesOnly(t *testing.T) {
	dir := t.TempDir()

	res := runIn(t, dir, "init", "--dir", dir)

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equalSlices(t, kitFiles, res.lines())
	for _, file := range kitFiles {
		content := readIn(t, dir, file)
		isTrue(t, !strings.Contains(content, "{%"), file+" has no placeholders left")
	}
	isTrue(t, !fileExists(filepath.Join(dir, "AGENTS.md")), "init does not write AGENTS.md")
	isTrue(t, !fileExists(filepath.Join(dir, "CONTRIBUTING.md")), "init does not write CONTRIBUTING.md")
	isTrue(t, !fileExists(filepath.Join(dir, "VERSION")), "init does not write VERSION")
	isTrue(t, !fileExists(filepath.Join(dir, ".github")), "init does not write .github")
}

func TestInitKeepsFilesThatExist(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issues", "TEMPLATE.md"), []byte("# My own template\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runIn(t, dir, "init", "--dir", dir)

	equal(t, 0, res.exit)
	contains(t, res.error, "issues/TEMPLATE.md exists, so init did not change it")
	equal(t, "# My own template\n", readIn(t, dir, "issues/TEMPLATE.md"))
	equal(t, len(kitFiles)-1, len(res.lines()))
}

func TestInitReplacesFilesWithForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issues", "TEMPLATE.md"), []byte("# My own template\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runIn(t, dir, "init", "--dir", dir, "--force")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, readIn(t, dir, "issues/TEMPLATE.md"), "Copy this file to issues/open/<issue-name>.md")
}

func TestInitMakesARepoThatTheOtherCommandsCanUse(t *testing.T) {
	dir := t.TempDir()
	runIn(t, dir, "init", "--dir", dir)

	created := runIn(t, dir, "new", "first-issue-test")
	equal(t, 0, created.exit)
	equal(t, "issues/open/first-issue-test.md", strings.TrimSpace(created.output))
}
