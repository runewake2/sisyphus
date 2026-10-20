package repo

import (
	"os"
	"os/exec"
	"path/filepath"
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
