// Package scratch makes scratch git worktrees: a short-lived copy of one commit of a repo, outside
// every working copy of a person or an agent, for a reviewer or a tool to work in. The worktrees of
// one git directory are added and removed one at a time, because one git worktree prune can remove
// the half-made worktree of another.
package scratch

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	locksMu sync.Mutex
	locks   = map[string]*sync.Mutex{}
)

// lock returns the lock of the worktrees of a git directory.
func lock(gitDir string) *sync.Mutex {
	locksMu.Lock()
	defer locksMu.Unlock()
	if locks[gitDir] == nil {
		locks[gitDir] = &sync.Mutex{}
	}
	return locks[gitDir]
}

// Worktree is one scratch worktree.
type Worktree struct {
	GitDir string
	Dir    string
}

// New adds a scratch worktree of commit, detached, in a new temporary directory whose name starts
// with prefix.
func New(gitDir, commit, prefix string) (*Worktree, error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return nil, err
	}
	os.Remove(dir) // git worktree add makes it
	l := lock(gitDir)
	l.Lock()
	defer l.Unlock()
	w := &Worktree{GitDir: gitDir, Dir: dir}
	if _, err := Git(gitDir, "", "worktree", "add", "--detach", "--quiet", dir, commit); err != nil {
		return nil, err
	}
	return w, nil
}

// Remove removes the worktree and its directory.
func (w *Worktree) Remove() {
	l := lock(w.GitDir)
	l.Lock()
	defer l.Unlock()
	if _, err := Git(w.GitDir, "", "worktree", "remove", "--force", w.Dir); err != nil {
		os.RemoveAll(w.Dir)
		Git(w.GitDir, "", "worktree", "prune")
	}
}

// Git runs git in the worktree.
func (w *Worktree) Git(args ...string) (string, error) {
	return Git(w.GitDir, w.Dir, args...)
}

// Commit commits the changes of the worktree, if any, and returns the head and the diff from base.
// With no commit and no change, it returns no head. The commits from base must be a linear line.
func (w *Worktree) Commit(base, message string) (string, string, error) {
	status, err := w.Git("status", "--porcelain")
	if err != nil {
		return "", "", err
	}
	if status != "" {
		if _, err := w.Git("add", "-A"); err != nil {
			return "", "", err
		}
		if _, err := w.Git("commit", "--quiet", "-m", message); err != nil {
			return "", "", err
		}
	}
	head, err := w.Git("rev-parse", "HEAD")
	if err != nil || head == base {
		return "", "", err
	}
	if merges, _ := w.Git("rev-list", "--merges", base+"..HEAD"); merges != "" {
		return "", "", errors.New("the changes have merge commits; they must be a linear line of commits on the base")
	}
	if ancestors, _ := w.Git("rev-list", "HEAD.."+base); ancestors != "" {
		return "", "", errors.New("the changes do not start at the base commit")
	}
	diff, err := w.Git("diff", base, "HEAD")
	return head, diff, err
}

// Git runs git on the git directory, in the worktree dir if it is not empty, with a committer when
// the user has none, and returns its trimmed stdout.
func Git(gitDir, dir string, args ...string) (string, error) {
	full := []string{"--git-dir=" + gitDir}
	if dir != "" {
		full = []string{"-C", dir}
	}
	full = append(full, Identity()...)
	cmd := exec.Command("git", append(full, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// Identity gives git a committer when the user has none, so that a commit never fails for it. A
// configured name and email win.
func Identity() []string {
	var args []string
	for key, fallback := range map[string]string{"user.name": "hephaestus", "user.email": "hephaestus@localhost"} {
		if out, err := exec.Command("git", "config", key).Output(); err != nil || strings.TrimSpace(string(out)) == "" {
			args = append(args, "-c", key+"="+fallback)
		}
	}
	return args
}
