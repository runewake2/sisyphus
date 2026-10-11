package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConcurrentStartsOfOneIssueGiveItOneOwner(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	const agents = 8
	results := make([]result, agents)
	var wg sync.WaitGroup
	for i := range agents {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = r.run("start", "fix-login-bug", "--owner", fmt.Sprintf("agent-%d", i))
		}()
	}
	wg.Wait()

	winners := 0
	for _, res := range results {
		if res.exit == 0 {
			winners++
		} else {
			contains(t, res.error, "is in progress, owned by")
		}
	}
	equal(t, 1, winners)
}

func TestConcurrentAppendsLoseNoText(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	const writers = 10
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.mustRun("edit", "fix-login-bug", "notes", fmt.Sprintf("note-%d", i), "--append")
		}()
	}
	wg.Wait()

	body := r.read("issues/open/fix-login-bug.md")
	for i := range writers {
		contains(t, body, fmt.Sprintf("note-%d", i))
	}
}

func TestReadersNeverSeeAPartOfAnIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	long := strings.Repeat("A long line of notes. ", 2000)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 20 {
			r.mustRun("edit", "fix-login-bug", "notes", fmt.Sprintf("%d %s", i, long))
		}
		close(done)
	}()
	for reads := 0; ; reads++ {
		select {
		case <-done:
			wg.Wait()
			return
		default:
		}
		res := r.run("show", "fix-login-bug", "--json")
		if res.exit != 0 || res.error != "" {
			t.Fatalf("read %d failed: %s", reads, res.error)
		}
	}
}

func TestSaveLeavesNoTemporaryFile(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("edit", "fix-login-bug", "notes", "Text.")

	entries, err := os.ReadDir(filepath.Join(r.root, "issues", "open"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		isTrue(t, !strings.HasSuffix(entry.Name(), ".tmp"), "a temporary file stays: "+entry.Name())
	}
	isTrue(t, !r.exists("issues/.lock"), "the lock is released")
}

func TestAHeldLockMakesACommandWaitAndThenFail(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("issues/.lock", "pid 1 since now\n")
	lockWait = 100 * time.Millisecond
	t.Cleanup(func() { lockWait = 10 * time.Second })

	res := r.run("start", "fix-login-bug", "--owner", "alice")

	equal(t, 1, res.exit)
	contains(t, res.error, "Another sisyphus command holds the lock")
	contains(t, res.error, "pid 1")
}

func TestAStaleLockIsRemoved(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("issues/.lock", "pid 1\n")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(r.root, "issues", ".lock"), old, old); err != nil {
		t.Fatal(err)
	}

	r.mustRun("start", "fix-login-bug", "--owner", "alice")

	isTrue(t, !r.exists("issues/.lock"), "the stale lock is gone")
}

func TestSharedStateDirIsTheSameForEveryJJWorkspace(t *testing.T) {
	base := t.TempDir()
	mustWrite(t, filepath.Join(base, "main", ".jj", "repo", "store", "type"), "git")
	mustWrite(t, filepath.Join(base, "workspaces", "second", ".jj", "repo"), "../../../main/.jj/repo")

	first, ok1 := sharedStateDir(filepath.Join(base, "main"))
	second, ok2 := sharedStateDir(filepath.Join(base, "workspaces", "second", "sub"))

	isTrue(t, ok1 && ok2, "both workspaces have a shared directory")
	equal(t, filepath.Join(base, "main", ".jj", "repo", "sisyphus"), first)
	equal(t, first, second)
}

func TestSharedStateDirIsTheSameForEveryGitWorktree(t *testing.T) {
	base := t.TempDir()
	mustWrite(t, filepath.Join(base, "main", ".git", "worktrees", "second", "commondir"), "../..\n")
	mustWrite(t, filepath.Join(base, "second", ".git"), "gitdir: "+filepath.Join(base, "main", ".git", "worktrees", "second")+"\n")

	first, ok1 := sharedStateDir(filepath.Join(base, "main"))
	second, ok2 := sharedStateDir(filepath.Join(base, "second"))

	isTrue(t, ok1 && ok2, "both worktrees have a shared directory")
	equal(t, filepath.Join(base, "main", ".git", "sisyphus"), first)
	equal(t, first, second)
}

func TestNoSharedStateDirWithoutAVCS(t *testing.T) {
	_, ok := sharedStateDir(t.TempDir())

	isTrue(t, !ok, "a directory in no repo shares nothing")
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
