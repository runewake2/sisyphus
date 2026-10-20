package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// run runs a VCS command in dir and fails the test when it fails.
func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	switch name {
	case "git":
		args = append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)
	case "jj":
		args = append([]string{"--config", "user.name=test", "--config", "user.email=test@example.com"}, args...)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// gitRepo makes a git repo with a commit on main and one more on the branch work.
func gitRepo(t *testing.T) Repo {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "-b", "main")
	write(t, dir, "one.txt", "one")
	run(t, dir, "git", "add", "-A")
	run(t, dir, "git", "commit", "-q", "-m", "one")
	run(t, dir, "git", "checkout", "-q", "-b", "work")
	write(t, dir, "two.txt", "two")
	run(t, dir, "git", "add", "-A")
	run(t, dir, "git", "commit", "-q", "-m", "two")
	r, err := Find(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// jjRepo makes a jj repo with a commit on the bookmark main and one more on the bookmark work. It
// skips the test when jj is not installed.
func jjRepo(t *testing.T) Repo {
	t.Helper()
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	dir := t.TempDir()
	run(t, dir, "jj", "git", "init")
	write(t, dir, "one.txt", "one")
	run(t, dir, "jj", "commit", "-m", "one")
	run(t, dir, "jj", "bookmark", "create", "main", "-r", "@-")
	write(t, dir, "two.txt", "two")
	run(t, dir, "jj", "commit", "-m", "two")
	run(t, dir, "jj", "bookmark", "create", "work", "-r", "@-")
	r, err := Find(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// repos are the two repos that every fact must come from the same way.
var repos = map[string]func(*testing.T) Repo{"git": gitRepo, "jj": jjRepo}

func TestFindPrefersJJ(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "-b", "main")
	run(t, dir, "jj", "git", "init", "--colocate")
	r, err := Find(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.VCS != JJ {
		t.Fatalf("Find in a colocated repo gives the VCS %q, not jj", r.VCS)
	}
}

func TestIsAncestor(t *testing.T) {
	for name, make := range repos {
		t.Run(name, func(t *testing.T) {
			r := make(t)
			cases := []struct {
				a, b string
				want bool
			}{
				{"main", "work", true},
				{"main", "main", true},
				{"work", "main", false},
			}
			for _, c := range cases {
				got, err := r.IsAncestor(c.a, c.b)
				if err != nil {
					t.Fatalf("IsAncestor(%q, %q): %v", c.a, c.b, err)
				}
				if got != c.want {
					t.Errorf("IsAncestor(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
				}
			}
			if got, err := r.IsAncestor("no-such-revision", "main"); err == nil && got {
				t.Error("IsAncestor with a missing revision reports true without an error")
			}
			// The status command asks with a commit id from a message, not with a bookmark.
			commit, err := r.CommitOf("main")
			if err != nil {
				t.Fatalf("CommitOf(main): %v", err)
			}
			if got, err := r.IsAncestor(commit, "work"); err != nil || !got {
				t.Errorf("IsAncestor(<commit id of main>, work) = %v, %v, want true", got, err)
			}
		})
	}
}

func TestCommitOf(t *testing.T) {
	for name, make := range repos {
		t.Run(name, func(t *testing.T) {
			r := make(t)
			commit, err := r.CommitOf("main")
			if err != nil || len(commit) != 40 {
				t.Errorf("CommitOf(main) = %q, %v, want a full commit id", commit, err)
			}
			if commit, err := r.CommitOf("no-such-revision"); err == nil && commit != "" {
				t.Errorf("CommitOf(no-such-revision) = %q with no error", commit)
			}
		})
	}
}

func TestHasBookmark(t *testing.T) {
	for name, make := range repos {
		t.Run(name, func(t *testing.T) {
			r := make(t)
			if !r.HasBookmark("work") {
				t.Error("HasBookmark(work) is false")
			}
			if r.HasBookmark("no-such-bookmark") {
				t.Error("HasBookmark(no-such-bookmark) is true")
			}
			// A commit that is no bookmark does not count, so a free name stays free.
			if commit, err := r.CommitOf("main"); err == nil && r.HasBookmark(commit) {
				t.Error("HasBookmark(<commit id>) is true")
			}
		})
	}
}

func TestRunTrimsAndFormatsTheError(t *testing.T) {
	r := gitRepo(t)
	out, err := Run(r.Root, "git", nil, "rev-parse", "--verify", "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 40 || strings.TrimSpace(out) != out {
		t.Fatalf("the stdout is not trimmed: %q", out)
	}
	_, err = Run(r.Root, "git", nil, "rev-parse", "--verify", "no-such-thing")
	if err == nil || !strings.HasPrefix(err.Error(), "git rev-parse --verify no-such-thing: ") {
		t.Fatalf("the error does not name the command and args: %v", err)
	}
	if !strings.Contains(err.Error(), "fatal") {
		t.Fatalf("the error does not carry the stderr: %v", err)
	}
}

func TestRunKeepsHiddenFlagsOutOfTheError(t *testing.T) {
	r := gitRepo(t)
	_, err := Run(r.Root, "git", []string{"-c", "user.name=hephaestus"}, "rev-parse", "--verify", "no-such-thing")
	if err == nil || strings.Contains(err.Error(), "user.name") {
		t.Fatalf("a hidden flag reaches the error: %v", err)
	}
}

func TestLines(t *testing.T) {
	got := Lines("a\\b\n\n  c  \n")
	// ToSlash changes the separator only on Windows; everywhere, Lines trims and drops empties.
	if !slices.Equal(got, []string{filepath.ToSlash("a\\b"), "c"}) {
		t.Fatalf("Lines gives %v", got)
	}
}

func TestExportAndImportBookmarks(t *testing.T) {
	r := jjRepo(t)
	if err := r.ExportBookmarks(); err != nil {
		t.Fatal(err)
	}
	gitDir, err := r.GitDir()
	if err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) *exec.Cmd {
		return exec.Command("git", append([]string{"--git-dir=" + gitDir}, args...)...)
	}
	// The export makes the bookmark a git branch of the store.
	if err := git("rev-parse", "--verify", "--quiet", "refs/heads/work").Run(); err != nil {
		t.Fatalf("the exported bookmark work is not a git branch: %v", err)
	}
	// The import makes a git branch of the store a bookmark.
	if out, err := git("branch", "imported", "work").CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, out)
	}
	r.ImportBookmarks()
	if !r.HasBookmark("imported") {
		t.Fatal("the imported branch is not a bookmark")
	}
}

func TestExportBookmarksOutsideJJDoesNothing(t *testing.T) {
	r := gitRepo(t)
	if err := r.ExportBookmarks(); err != nil {
		t.Fatalf("the export of a git repo is not a no-op: %v", err)
	}
	r.ImportBookmarks() // and the import panics or errors nowhere
}

func TestExists(t *testing.T) {
	for name, make := range repos {
		t.Run(name, func(t *testing.T) {
			r := make(t)
			if !r.Exists("main") {
				t.Error("Exists(main) is false")
			}
			if r.Exists("no-such-revision") {
				t.Error("Exists(no-such-revision) is true")
			}
		})
	}
}
