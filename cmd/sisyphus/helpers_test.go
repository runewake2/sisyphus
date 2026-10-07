package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type result struct {
	exit   int
	output string
	error  string
}

func (r result) lines() []string {
	var lines []string
	for _, line := range strings.Split(r.output, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// testRepo is a temporary repo with the real issues/TEMPLATE.md. Each test gets its own, so tests do not change the real repo.
type testRepo struct {
	t    *testing.T
	root string
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()
	r := &testRepo{t: t, root: t.TempDir()}
	r.write("issues/TEMPLATE.md", readSourceFile(t, "issues/TEMPLATE.md"))
	r.write("README.md", "# Readme\n")
	r.write("changelog/0.0.0.md", "# 0.0.0\n")
	r.write("design/widget-scheduler.md", "# Widget Scheduler\n\n## Plan Components\n\n## Status API\n\nText.\n")
	r.write("design/example/README.md", "# Example\n")
	r.write("design/example/plan/basic-plan.yaml", "plan:\n")
	return r
}

// readSourceFile reads a file of the real repo that contains the tests.
func readSourceFile(t *testing.T, relative string) string {
	t.Helper()
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func (r *testRepo) pathOf(relative string) string {
	return filepath.Join(r.root, filepath.FromSlash(relative))
}

func (r *testRepo) write(relative, content string) {
	r.t.Helper()
	path := r.pathOf(relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *testRepo) read(relative string) string {
	r.t.Helper()
	content, err := os.ReadFile(r.pathOf(relative))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(content)
}

// git runs a git command in the test repo, for tests that need a real .git directory.
func (r *testRepo) git(args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	if output, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %v failed: %s\n%s", args, err, output)
	}
}

func (r *testRepo) exists(relative string) bool {
	return fileExists(r.pathOf(relative))
}

func (r *testRepo) frontmatter(relative string) *document {
	r.t.Helper()
	doc, err := parseDocument(r.read(relative))
	if err != nil {
		r.t.Fatal(err)
	}
	return doc
}

func (r *testRepo) run(args ...string) result {
	var output, errors bytes.Buffer
	exit := run(args, func() (string, error) { return r.root, nil }, &output, &errors)
	return result{exit: exit, output: output.String(), error: errors.String()}
}

func (r *testRepo) mustRun(args ...string) result {
	r.t.Helper()
	res := r.run(args...)
	if res.exit != 0 {
		r.t.Fatalf("sisyphus %v failed: %s", args, res.error)
	}
	return res
}

func equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func equalSlices[T comparable](t *testing.T, want, got []T) {
	t.Helper()
	if !slices.Equal(want, got) {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func contains(t *testing.T, s, substring string) {
	t.Helper()
	if !strings.Contains(s, substring) {
		t.Errorf("want %q to contain %q", s, substring)
	}
}

func isTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Error(message)
	}
}
