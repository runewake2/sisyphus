package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testRepo is a temporary git repo with a main branch.
type testRepo struct {
	t    *testing.T
	root string
}

func newGitRepo(t *testing.T) *testRepo {
	t.Helper()
	r := &testRepo{t: t, root: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	return r
}

func (r *testRepo) write(rel, content string) {
	r.t.Helper()
	path := filepath.Join(r.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *testRepo) read(rel string) string {
	r.t.Helper()
	data, err := os.ReadFile(filepath.Join(r.root, filepath.FromSlash(rel)))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(data)
}

func (r *testRepo) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(r.root, filepath.FromSlash(rel)))
	return err == nil
}

func (r *testRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// commit commits every file of the working copy.
func (r *testRepo) commit(message string) {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "-q", "-m", message)
}

type result struct {
	exit           int
	output, errors string
}

func (r *testRepo) run(args ...string) result {
	var out, errOut bytes.Buffer
	exit := run(append([]string{"-C", r.root}, args...), &out, &errOut)
	return result{exit: exit, output: out.String(), errors: errOut.String()}
}

func (r *testRepo) mustRun(args ...string) result {
	r.t.Helper()
	res := r.run(args...)
	if res.exit != 0 {
		r.t.Fatalf("hephaestus %v failed: %s", args, res.errors)
	}
	return res
}

func equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func contains(t *testing.T, s, substring string) {
	t.Helper()
	if !strings.Contains(s, substring) {
		t.Errorf("want %q to contain %q", s, substring)
	}
}

func notContains(t *testing.T, s, substring string) {
	t.Helper()
	if strings.Contains(s, substring) {
		t.Errorf("want %q not to contain %q", s, substring)
	}
}

func isTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Error(message)
	}
}
