package mergequeue

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
)

// A line is the new history that a run builds, in a scratch working copy of its own: a jj workspace
// in a jj repo, or a git worktree in a git repo. The files of the working copy follow the tip of
// the line, so the planners and the checks read them there.
type line interface {
	// Dir is the directory of the scratch working copy.
	Dir() string
	// Repo is the scratch working copy as a repo, for the planners.
	Repo() repo.Repo
	// Commits returns the commits of bookmark that the line does not have yet, oldest first: not in
	// the base, and not in one of the done bookmarks.
	Commits(bookmark string, done []string) ([]string, error)
	// HasMerges reports whether one of those commits is a merge commit.
	HasMerges(bookmark string, done []string) (bool, error)
	// Pick puts one commit onto the line. A conflict gives ok false, with the files of the conflict.
	Pick(commit string) (files []string, ok bool, err error)
	// CommitFiles commits the changed files of the working copy, with the message.
	CommitFiles(message string) error
	// AddedSince returns the files below dir that the line added since the commit since, in the
	// order of the commits that added them.
	AddedSince(since, dir string) ([]string, error)
	// Head returns the commit at the tip of the line.
	Head() (string, error)
	// Rollback puts the line and its working copy back at the commit to, and drops what came after.
	// A rollback that fails is an error: what came after would leak into the repo or the line.
	Rollback(to string) error
	// Bookmark makes the bookmark name at the commit.
	Bookmark(name, commit string) error
	// Close removes the scratch working copy.
	Close()
}

// newLine makes the line of a run, at the base commit.
func newLine(r repo.Repo, baseID string) (line, error) {
	switch r.VCS {
	case repo.JJ:
		return newJJLine(r, baseID)
	case repo.Git:
		return newGitLine(r, baseID)
	}
	return nil, errors.New("The merge queue needs a jj or git repo.")
}

// jjLine builds the line with jj, in a scratch workspace of the repo. Its working-copy commit rides
// the tip: jj duplicate --insert-before @ puts the copy of a commit under it, so the duplicate is
// always @-, and the files of the working copy follow.
type jjLine struct {
	root     string // the workspace of the user, where the scratch workspace was added
	dir      string
	name     string
	baseID   string
	identity []string
}

func newJJLine(r repo.Repo, baseID string) (line, error) {
	dir, err := os.MkdirTemp("", "hephaestus-merge-queue-")
	if err != nil {
		return nil, err
	}
	os.Remove(dir) // jj workspace add makes it
	l := &jjLine{root: r.Root, dir: dir, name: filepath.Base(dir), baseID: baseID, identity: jjIdentity(r.Root)}
	if _, err := l.jjIn(r.Root, "workspace", "add", "--name", l.name, "-r", baseID, dir); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *jjLine) Dir() string     { return l.dir }
func (l *jjLine) Repo() repo.Repo { return repo.Repo{Root: l.dir, VCS: repo.JJ} }

// jj runs jj in the scratch workspace, and returns its trimmed stdout.
func (l *jjLine) jj(args ...string) (string, error) {
	return l.jjIn(l.dir, args...)
}

func (l *jjLine) jjIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("jj", append(slices.Clone(l.identity), args...)...)
	cmd.Dir = dir
	out, errOut := &strings.Builder{}, &strings.Builder{}
	cmd.Stdout, cmd.Stderr = out, errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("jj %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// jjIdentity gives jj an author when the user has none, so that a run never fails for it. A
// configured name and email win.
func jjIdentity(dir string) []string {
	var args []string
	for key, fallback := range map[string]string{"user.name": "hephaestus merge queue", "user.email": "hephaestus@localhost"} {
		cmd := exec.Command("jj", "config", "get", key)
		cmd.Dir = dir
		if out, err := cmd.Output(); err != nil || strings.TrimSpace(string(out)) == "" {
			args = append(args, "--config", key+"="+fallback)
		}
	}
	slices.Sort(args)
	return args
}

// newCommits is the revset of the commits of bookmark that the line does not have yet.
func (l *jjLine) newCommits(bookmark string, done []string) string {
	set := fmt.Sprintf("::bookmarks(exact:%s) ~ ::%s", strconv.Quote(bookmark), l.baseID)
	for _, bm := range done {
		set += fmt.Sprintf(" ~ ::bookmarks(exact:%s)", strconv.Quote(bm))
	}
	return set
}

func (l *jjLine) Commits(bookmark string, done []string) ([]string, error) {
	out, err := l.jj("log", "--no-graph", "--reversed", "-r", l.newCommits(bookmark, done), "-T", `commit_id ++ "\n"`)
	return lines(out), err
}

func (l *jjLine) HasMerges(bookmark string, done []string) (bool, error) {
	out, err := l.jj("log", "--no-graph", "-r", "("+l.newCommits(bookmark, done)+") & merges()", "-T", `commit_id ++ "\n"`)
	return len(lines(out)) > 0, err
}

func (l *jjLine) Pick(commit string) ([]string, bool, error) {
	if _, err := l.jj("duplicate", "-r", commit, "--insert-before", "@"); err != nil {
		return nil, false, err
	}
	conflict, err := l.jj("log", "--no-graph", "-r", "@-", "-T", `if(conflict, "true", "false")`)
	if err != nil {
		return nil, false, err
	}
	if conflict != "true" {
		return nil, true, nil
	}
	list, _ := l.jj("resolve", "--list", "-r", "@-")
	var files []string
	for _, file := range lines(list) {
		// A line of jj resolve --list is the path, two or more spaces, and what the conflict is.
		if i := strings.Index(file, "  "); i > 0 {
			file = file[:i]
		}
		files = append(files, file)
	}
	return files, false, nil
}

func (l *jjLine) CommitFiles(message string) error {
	_, err := l.jj("commit", "-m", message)
	return err
}

func (l *jjLine) AddedSince(since, dir string) ([]string, error) {
	commits, err := l.jj("log", "--no-graph", "--reversed", "-r", fmt.Sprintf("(%s..@) ~ @", since), "-T", `commit_id ++ "\n"`)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, commit := range lines(commits) {
		summary, err := l.jj("diff", "-r", commit, "--summary")
		if err != nil {
			return nil, err
		}
		for _, change := range lines(summary) {
			if file, ok := strings.CutPrefix(change, "A "); ok && strings.HasPrefix(file, dir+"/") {
				files = append(files, file)
			}
		}
	}
	return files, nil
}

func (l *jjLine) Head() (string, error) {
	return l.jj("log", "--no-graph", "-r", "@-", "-T", "commit_id")
}

// Rollback abandons the commits above to, also the working-copy commit: jj makes a new empty one on
// to, which drops what the checks left in the working copy, and the files follow. A failed abandon
// is an error: the abandoned commits would stay as visible heads of the repo.
func (l *jjLine) Rollback(to string) error {
	_, err := l.jj("abandon", "-r", to+"..@")
	return err
}

func (l *jjLine) Bookmark(name, commit string) error {
	_, err := l.jj("bookmark", "create", name, "-r", commit)
	return err
}

// Close forgets the scratch workspace and abandons its working-copy commit, so that the repo keeps
// no trace of the run but the result.
func (l *jjLine) Close() {
	wc, err := l.jj("log", "--no-graph", "-r", "@", "-T", "commit_id")
	l.jjIn(l.root, "workspace", "forget", l.name)
	if err == nil {
		l.jjIn(l.root, "--ignore-working-copy", "abandon", "-r", wc)
	}
	os.RemoveAll(l.dir)
}

// gitLine builds the line with git, in a scratch worktree of the git directory of the repo,
// detached at the base: HEAD is the tip.
type gitLine struct {
	gitDir   string
	dir      string
	baseID   string
	identity []string
}

func newGitLine(r repo.Repo, baseID string) (line, error) {
	gitDir, err := r.GitDir()
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "hephaestus-merge-queue-")
	if err != nil {
		return nil, err
	}
	os.Remove(dir) // git worktree add makes it
	l := &gitLine{gitDir: gitDir, dir: dir, baseID: baseID}
	if _, err := l.gitDirCmd("worktree", "add", "--detach", "--quiet", dir, baseID); err != nil {
		return nil, err
	}
	l.identity = identity(dir)
	return l, nil
}

func (l *gitLine) Dir() string     { return l.dir }
func (l *gitLine) Repo() repo.Repo { return repo.Repo{Root: l.dir, VCS: repo.Git} }

// git runs git in the worktree, and returns its stdout.
func (l *gitLine) git(args ...string) (string, error) {
	cmd := exec.Command("git", append(slices.Clone(l.identity), args...)...)
	cmd.Dir = l.dir
	out, errOut := &strings.Builder{}, &strings.Builder{}
	cmd.Stdout, cmd.Stderr = out, errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}

// gitDirCmd runs git on the git directory, outside the worktree.
func (l *gitLine) gitDirCmd(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"--git-dir=" + l.gitDir}, args...)...)
	out, errOut := &strings.Builder{}, &strings.Builder{}
	cmd.Stdout, cmd.Stderr = out, errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// identity gives git a committer when the repo has none, so that a run never fails for it. A
// configured name and email win.
func identity(dir string) []string {
	var args []string
	for key, fallback := range map[string]string{"user.name": "hephaestus merge queue", "user.email": "hephaestus@localhost"} {
		cmd := exec.Command("git", "config", key)
		cmd.Dir = dir
		if out, err := cmd.Output(); err != nil || strings.TrimSpace(string(out)) == "" {
			args = append(args, "-c", key+"="+fallback)
		}
	}
	slices.Sort(args)
	return args
}

// newCommits is the rev-list arguments of the commits of bookmark that the line does not have yet.
func (l *gitLine) newCommits(bookmark string, done []string) []string {
	args := []string{bookmark, "^" + l.baseID}
	for _, bm := range done {
		args = append(args, "^"+bm)
	}
	return args
}

func (l *gitLine) Commits(bookmark string, done []string) ([]string, error) {
	out, err := l.git(append([]string{"rev-list", "--reverse", "--topo-order"}, l.newCommits(bookmark, done)...)...)
	return lines(out), err
}

func (l *gitLine) HasMerges(bookmark string, done []string) (bool, error) {
	out, err := l.git(append([]string{"rev-list", "--merges"}, l.newCommits(bookmark, done)...)...)
	return len(lines(out)) > 0, err
}

func (l *gitLine) Pick(commit string) ([]string, bool, error) {
	if _, err := l.git("cherry-pick", "--allow-empty", "--keep-redundant-commits", commit); err != nil {
		files, _ := l.git("diff", "--name-only", "--diff-filter=U")
		return lines(files), false, nil
	}
	return nil, true, nil
}

func (l *gitLine) CommitFiles(message string) error {
	if _, err := l.git("add", "-A"); err != nil {
		return err
	}
	_, err := l.git("commit", "--quiet", "-m", message)
	return err
}

func (l *gitLine) AddedSince(since, dir string) ([]string, error) {
	out, err := l.git("log", "--reverse", "--diff-filter=A", "--name-only", "--format=", since+"..HEAD", "--", dir)
	return lines(out), err
}

func (l *gitLine) Head() (string, error) {
	head, err := l.git("rev-parse", "HEAD")
	return strings.TrimSpace(head), err
}

func (l *gitLine) Rollback(to string) error {
	l.git("cherry-pick", "--abort") // there may be no cherry-pick to abort
	if _, err := l.git("reset", "--quiet", "--hard", to); err != nil {
		return err
	}
	_, err := l.git("clean", "-fdq")
	return err
}

func (l *gitLine) Bookmark(name, commit string) error {
	_, err := l.gitDirCmd("branch", name, commit)
	return err
}

func (l *gitLine) Close() {
	l.gitDirCmd("worktree", "remove", "--force", l.dir)
	os.RemoveAll(l.dir)
	l.gitDirCmd("worktree", "prune")
}
