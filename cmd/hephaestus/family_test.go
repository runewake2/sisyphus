package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// familyRole is a tool role of a family: formatter/upper, whose script is in its directory.
func familyRole(name string) string {
	return toolRoleFile(name, `sh "$HEPHAESTUS_ROLE_DIR/format.sh"`, "", "")
}

func TestARoleOfAFamilyLivesInADirectory(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/formatter/upper/role.md", familyRole("formatter/upper"))
	r.write("hephaestus/roles/formatter/README.md", "# Formatters\n\nNot a role file.\n")
	r.write("hephaestus/roles/formatter/upper/notes.md", "Not a role file either.\n")

	res := r.mustRun("roles")

	contains(t, res.output, "formatter/upper  tool")
	notContains(t, res.output, "README")
	r.mustRun("delegate", "fix-login-bug", "formatter/upper")
	equal(t, "formatter/upper", r.metadataOf("fix-login-bug")["delegate"])
}

func TestTheNameOfAFamilyRoleMustBeItsPath(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/formatter/upper/role.md", familyRole("upper"))

	res := r.run("roles")

	contains(t, res.errors, `formatter/upper/role.md: role is "upper", but the path of the file gives "formatter/upper"`)
}

func TestAFileBelowTheRolesThatIsNotRoleMdIsNoRole(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/formatter/upper.md", familyRole("formatter/upper"))

	res := r.run("delegate", "fix-login-bug", "formatter/upper")

	equal(t, 1, res.exit)
	contains(t, res.errors, `No role "formatter/upper"`)
}

func TestAFamilyToolReviewerRunsTheScriptOfItsDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the script is sh")
	}
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/formatter/upper/role.md", familyRole("formatter/upper"))
	r.write("hephaestus/roles/formatter/upper/format.sh", `cd "$(hephaestus files --root)"
hephaestus files --changed '*.txt' | while read -r f; do tr a-z A-Z < "$f" > "$f.new" && mv "$f.new" "$f"; done
echo "base $HEPHAESTUS_BASE"
`)
	r.write("old.txt", "old text\n")
	r.commit("tracker, roles, and old text")
	r.git("checkout", "-q", "-b", "ai/text")
	r.write("new.txt", "new text\n")
	r.commit("new text")
	r.git("checkout", "-q", "main")
	r.mustRunSisyphus("update", "fix-login-bug", "in-progress", "--owner", "alice", "--bookmark", "ai/text")
	hephaestusOnPath(t)

	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "formatter/upper")

	if res.exit != 0 {
		t.Fatalf("the referendum did not pass: %s", res.errors)
	}
	contains(t, res.output, "round 1 (v1): formatter/upper modify.")
	// Only the file that the change touches is formatted.
	v2 := "referendum/" + strings.TrimSpace(r.metadataOf("fix-login-bug")["referendum-record"]) + "/v2"
	equal(t, "NEW TEXT\n", r.git("show", v2+":new.txt"))
	equal(t, "old text\n", r.git("show", v2+":old.txt"))
	contains(t, r.read("referendums/"+r.metadataOf("fix-login-bug")["referendum-record"]+".md"), "base "+strings.TrimSpace(r.git("rev-parse", "main")))
}

func TestFilesPrintsTheFilesThatTheWorkChanges(t *testing.T) {
	r := newGitRepo(t)
	r.write("a.go", "package a\n")
	r.write("docs/a.md", "# A\n")
	r.commit("base")
	r.git("checkout", "-q", "-b", "work")
	r.write("b.go", "package b\n")
	r.write("docs/b.md", "# B\n")
	r.write("skip/c.md", "# C\n")
	r.commit("work")
	r.write("d.go", "package d\n") // not committed yet

	equal(t, "a.go\nb.go\nd.go\n", r.mustRun("files", "*.go").output)
	equal(t, "b.go\nd.go\n", r.mustRun("files", "--changed", "*.go").output)
	equal(t, "docs/b.md\n", r.mustRun("files", "--changed", "--exclude", "skip", "*.md").output)
	equal(t, "docs/b.md\n", r.mustRun("files", "--changed", "docs/*").output)
	equal(t, "b.go\x00d.go\x00", r.mustRun("files", "--changed", "-0", "*.go").output)
	equal(t, r.root+"\n", r.mustRun("files", "--root").output)
	t.Setenv("HEPHAESTUS_BASE", "work")
	equal(t, "d.go\n", r.mustRun("files", "--changed", "*.go").output)
}

// hephaestusOnPath builds this hephaestus command onto PATH, for the scripts of roles that run it.
func hephaestusOnPath(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	cmd := execCommandIn("..", "go", "build", "-o", filepath.Join(dir, "hephaestus"), "./hephaestus")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building hephaestus: %v\n%s", err, out)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func execCommandIn(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd
}

func TestWorksWithMustNameRolesThatExist(t *testing.T) {
	r := newTrackerRepo(t)
	withPartners := strings.Replace(toolRoleFile("formatter/upper", "true", "", ""), "## Tools", "## Works with\n\n- `lint-runner`: runs the linters.\n- `formatter`: the family.\n- `nobody`: is not a role.\n\n## Tools", 1)
	r.write("hephaestus/roles/formatter/upper/role.md", withPartners)

	res := r.run("roles")

	equal(t, 1, res.exit)
	contains(t, res.errors, `formatter/upper/role.md: the section "## Works with" names "nobody", which is no role or family of roles`)
	notContains(t, res.errors, `"lint-runner"`)
	notContains(t, res.errors, `names "formatter"`)
}
