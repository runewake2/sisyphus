package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeEditor writes a script that appends a line to the file it gets, and returns its path.
func fakeEditor(t *testing.T, line string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake editor is a shell script")
	}
	script := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\\n' '"+line+"' >> \"$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestEditWithOnlyANameRunsTheEditorOnTheIssueFile(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/fix-login-bug")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", fakeEditor(t, "Edited in the editor."))

	res := r.mustRun("edit", "fix-login-bug")

	equal(t, "", res.error)
	contains(t, r.read("issues/open/web/fix-login-bug.md"), "Edited in the editor.")
}

func TestEditInTheEditorWarnsAboutAProblemAfterTheEdit(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("issues/open/fix-login-bug.md", r.read("issues/open/fix-login-bug.md")[4:]) // drop the first "---"
	t.Setenv("VISUAL", fakeEditor(t, "x"))

	contains(t, r.mustRun("edit", "fix-login-bug").error, "cannot be read")
}

func TestEditInTheEditorWithoutAnEditorFails(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	res := r.run("edit", "fix-login-bug")

	equal(t, 1, res.exit)
	contains(t, res.error, "Set VISUAL or EDITOR")
}

func TestEditWithOnlyANameRejectsTheSectionFlags(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	t.Setenv("VISUAL", fakeEditor(t, "x"))

	res := r.run("edit", "fix-login-bug", "--append")

	equal(t, 1, res.exit)
	contains(t, res.error, "--append needs <section> and <text>")
	isTrue(t, !strings.Contains(r.read("issues/open/fix-login-bug.md"), "\nx\n"), "the editor does not run")
}

func TestEditWithTwoArgumentsIsAnError(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	res := r.run("edit", "fix-login-bug", "summary")

	equal(t, 1, res.exit)
	contains(t, res.error, "Give <name> to open the issue in your editor, or <name> <section> <text>")
}
