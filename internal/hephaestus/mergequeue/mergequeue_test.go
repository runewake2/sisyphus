package mergequeue

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/internal/hephaestus/mergeorder"
	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
)

// failingLine wraps the real line of a git repo, and fails every Pick, so that a test can see what
// Run does on an error: it must roll the line back before it returns.
type failingLine struct {
	line
	rollbacks []string
}

func (l *failingLine) Pick(commit string) ([]string, bool, error) {
	return nil, false, errors.New("the pick fails for the test")
}

func (l *failingLine) Rollback(to string) error {
	l.rollbacks = append(l.rollbacks, to)
	return l.line.Rollback(to)
}

// gitRepo makes a git repo with a commit on main and one more on the branch work.
func gitRepo(t *testing.T) repo.Repo {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "one.txt"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "one")
	git("checkout", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(dir, "two.txt"), []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "two")
	git("checkout", "-q", "main")
	r, err := repo.Find(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestJJIdentityFallsBackInPairOrder(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	t.Setenv("JJ_CONFIG", os.DevNull) // hide the identity of the machine
	args := jjIdentity(t.TempDir())
	want := []string{"--config", "user.email=hephaestus@localhost", "--config", "user.name=hephaestus merge queue"}
	if !slices.Equal(args, want) {
		// A sorted flat list would put the two --config flags first and break jj.
		t.Fatalf("the fallback is %v, want %v", args, want)
	}
}

func TestAnErrorRollsTheLineBack(t *testing.T) {
	r := gitRepo(t)
	baseID, err := r.CommitOf("main")
	if err != nil {
		t.Fatal(err)
	}
	real, err := newLine(r, baseID)
	if err != nil {
		t.Fatal(err)
	}
	defer real.Close()
	failing := &failingLine{line: real}
	o := Options{
		Repo: r, Base: "main", Name: "mq/test", Now: time.Now, Out: io.Discard,
		Tasks: []mergeorder.Task{{Name: "work-task", State: "closed", Bookmark: "work"}},
	}
	start := Report{BaseID: baseID, Head: baseID, Base: "main", Name: "mq/test", Merged: []Merged{}, LeftOut: []LeftOut{}, Warnings: []string{}}
	report, err := run(context.Background(), o, failing, start)
	if err == nil {
		t.Fatal("a failing pick does not stop the run")
	}
	if len(failing.rollbacks) == 0 || failing.rollbacks[len(failing.rollbacks)-1] != baseID {
		t.Fatalf("the run did not roll the line back to the base: rollbacks %v", failing.rollbacks)
	}
	if report.Created {
		t.Fatal("a failed run created the bookmark")
	}
}
