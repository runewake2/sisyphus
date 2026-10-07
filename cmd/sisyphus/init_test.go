package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var kitFiles = []string{
	"AGENTS.md",
	"CHANGELOG.md",
	"CONTRIBUTING.md",
	"VERSION",
	"changelog/0.0.0.md",
	"design/decisions/.gitkeep",
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

func TestInitWritesTheKit(t *testing.T) {
	dir := t.TempDir()

	res := runIn(t, dir, "init", "--dir", dir, "--project", "calculator", "--bookmark-prefix", "samw/ai/")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equalSlices(t, kitFiles, res.lines())
	for _, file := range kitFiles {
		content := readIn(t, dir, file)
		isTrue(t, !strings.Contains(content, "{%"), file+" has no placeholders left")
	}
	equal(t, "0.0.0\n", readIn(t, dir, "VERSION"))
	agents := readIn(t, dir, "AGENTS.md")
	contains(t, agents, "../workspaces/calculator-<workspace-name>")
	contains(t, agents, "samw/ai/<workspace-name>")
	contains(t, readIn(t, dir, "CONTRIBUTING.md"), "# Contributing to calculator")
	contains(t, readIn(t, dir, "CHANGELOG.md"), "## [[0.0.0]] - "+today())
}

func TestInitUsesTheNameOfTheDirectoryAndADefaultPrefix(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-service")

	res := runIn(t, dir, "init", "--dir", dir)

	equal(t, 0, res.exit)
	agents := readIn(t, dir, "AGENTS.md")
	contains(t, agents, "Rules for AI agents that work in my-service.")
	contains(t, agents, "ai/<workspace-name>")
}

func TestInitKeepsFilesThatExist(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# My own rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runIn(t, dir, "init", "--dir", dir)

	equal(t, 0, res.exit)
	contains(t, res.error, "AGENTS.md exists, so init did not change it")
	equal(t, "# My own rules\n", readIn(t, dir, "AGENTS.md"))
	equal(t, len(kitFiles)-1, len(res.lines()))
}

func TestInitReplacesFilesWithForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# My own rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runIn(t, dir, "init", "--dir", dir, "--force")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, readIn(t, dir, "AGENTS.md"), "# AGENTS.md")
}

func TestInitMakesARepoThatTheOtherCommandsCanUse(t *testing.T) {
	dir := t.TempDir()
	runIn(t, dir, "init", "--dir", dir)

	created := runIn(t, dir, "new", "first-issue-test")
	equal(t, 0, created.exit)
	equal(t, "issues/open/first-issue-test.md", strings.TrimSpace(created.output))

	for _, document := range []string{"CONTRIBUTING.md", "AGENTS.md", "CHANGELOG.md", "changelog/0.0.0.md"} {
		res := runIn(t, dir, "links", filepath.Join(dir, document), "--json")
		equal(t, 0, res.exit)
		var rows []linkRow
		if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
			t.Fatalf("%s: %v", document, err)
		}
		for _, row := range rows {
			if row.Status != "ok" {
				t.Errorf("%s line %d: %s is %s", document, row.Line, row.Link, row.Status)
			}
		}
	}
}
