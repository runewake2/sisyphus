package main

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestNewCreatesAnIssueInASubdirectory(t *testing.T) {
	r := newTestRepo(t)

	res := r.mustRun("new", "web/auth/fix-login-bug", "--tags", "bug")

	equal(t, "issues/open/web/auth/fix-login-bug.md\n", res.output)
	equal(t, "Fix login bug", r.frontmatter("issues/open/web/auth/fix-login-bug.md").get("title"))
}

func TestNewRejectsAnInvalidDirectory(t *testing.T) {
	for _, name := range []string{"Web/fix-login-bug", "web_ui/fix-login-bug", "web//fix-login-bug", "/fix-login-bug", "web/fix"} {
		t.Run(name, func(t *testing.T) {
			r := newTestRepo(t)

			res := r.run("new", name)

			equal(t, 1, res.exit)
			contains(t, res.error, "Invalid issue name")
		})
	}
}

func TestUpdateKeepsTheSubdirectory(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/fix-login-bug")

	r.mustRun("update", "fix-login-bug", "in-progress", "--bookmark", "ai/work")
	res := r.mustRun("update", "web/auth/fix-login-bug", "closed")

	equal(t, "issues/closed/web/auth/fix-login-bug.md\n", res.output)
	isTrue(t, !r.exists("issues/in-progress/web/auth/fix-login-bug.md"), "the in-progress file is gone")
}

func TestEveryFormOfANameFindsAnIssueInASubdirectory(t *testing.T) {
	forms := []string{
		"fix-login-bug",
		"auth/fix-login-bug",
		"web/auth/fix-login-bug",
		"[[web/auth/fix-login-bug]]",
		"[[fix-login-bug|the bug]]",
		"#web/auth/fix-login-bug",
		"issues/open/web/auth/fix-login-bug.md",
	}
	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "web/auth/fix-login-bug")

			res := r.mustRun("show", form, "--json")

			var view issueView
			if err := json.Unmarshal([]byte(res.output), &view); err != nil {
				t.Fatal(err)
			}
			equal(t, "web/auth/fix-login-bug", view.Name)
		})
	}
}

func TestABareNameThatMatchesTwoIssuesIsRefused(t *testing.T) {
	for _, args := range [][]string{
		{"show", "fix-login-bug"},
		{"update", "fix-login-bug", "closed"},
		{"parent", "fix-login-bug", "--clear"},
		{"graph", "fix-login-bug"},
		{"new", "add-dark-mode", "--parent", "fix-login-bug"},
		{"new", "add-dark-mode", "--deferred-from", "fix-login-bug"},
	} {
		t.Run(args[0], func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "web/fix-login-bug")
			r.mustRun("new", "mobile/fix-login-bug")

			res := r.run(args...)

			equal(t, 1, res.exit)
			contains(t, res.error, "'fix-login-bug' matches more than one issue: mobile/fix-login-bug, web/fix-login-bug.")
			isTrue(t, r.exists("issues/open/web/fix-login-bug.md"), "nothing moved")
			isTrue(t, !r.exists("issues/open/add-dark-mode.md"), "nothing was created")
		})
	}
}

func TestAFullNameWinsOverALongerNameThatEndsWithIt(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "web/fix-login-bug")

	res := r.mustRun("update", "fix-login-bug", "closed")

	equal(t, "issues/closed/fix-login-bug.md\n", res.output)
	isTrue(t, r.exists("issues/open/web/fix-login-bug.md"), "the issue in the subdirectory did not move")
}

func TestLinksBetweenIssuesStoreTheFullName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/login-epic")
	r.mustRun("new", "web/auth/fix-login-bug")

	r.mustRun("new", "web/split-login-form", "--parent", "login-epic", "--depends-on", "fix-login-bug", "--deferred-from", "fix-login-bug")

	doc := r.frontmatter("issues/open/web/split-login-form.md")
	equal(t, "[[web/auth/login-epic]]", doc.get("parent"))
	equal(t, `["[[web/auth/fix-login-bug]]"]`, doc.get("depends-on"))
	equal(t, "[[web/auth/fix-login-bug]]", doc.get("deferred-from"))
}

func TestDeferredFromKeepsABookmarkThatIsNotAnIssue(t *testing.T) {
	r := newTestRepo(t)

	r.mustRun("new", "add-dark-mode", "--deferred-from", "ai/work")

	equal(t, "ai/work", r.frontmatter("issues/open/add-dark-mode.md").get("deferred-from"))
}

func TestABareLinkInFrontmatterStillFindsAnIssueInASubdirectory(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/login-epic")
	r.mustRun("new", "web/fix-login-bug")
	r.write("issues/open/web/split-login-form.md",
		"---\ntitle: \"Split login form\"\nstate: open\npriority: medium\nparent: \"[[login-epic]]\"\ndepends-on: [\"[[fix-login-bug]]\"]\n---\n")

	children := r.mustRun("list", "--parent", "web/login-epic")
	blocked := r.mustRun("list", "--blocked")
	closing := r.mustRun("update", "login-epic", "closed")

	contains(t, children.output, "web/split-login-form")
	contains(t, blocked.output, "web/split-login-form")
	contains(t, closing.error, "has sub-issues that are not closed: web/split-login-form.")
}

func TestListAndSearchPrintTheFullName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/fix-login-bug", "--context", "The session cookie expires early.")

	var listed []listRow
	if err := json.Unmarshal([]byte(r.mustRun("list", "--json").output), &listed); err != nil {
		t.Fatal(err)
	}
	var found []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "cookie", "--json").output), &found); err != nil {
		t.Fatal(err)
	}

	equal(t, 1, len(listed))
	equal(t, "web/auth/fix-login-bug", listed[0].Name)
	equal(t, 1, len(found))
	equal(t, "web/auth/fix-login-bug", found[0].Name)
}

func TestGraphUsesFullNames(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/login-epic")
	r.mustRun("new", "web/fix-login-bug", "--parent", "login-epic")

	var graph issueGraph
	if err := json.Unmarshal([]byte(r.mustRun("graph", "login-epic", "--json").output), &graph); err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, node := range graph.Nodes {
		names = append(names, node.Name)
	}
	slices.Sort(names)
	equalSlices(t, []string{"web/fix-login-bug", "web/login-epic"}, names)
}

func TestResolveFindsAnIssueInASubdirectoryByAnyEndOfItsPath(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/fix-login-bug")

	for _, link := range []string{"[[fix-login-bug]]", "[[auth/fix-login-bug]]", "[[web/auth/fix-login-bug]]"} {
		equal(t, "issues/open/web/auth/fix-login-bug.md\n", r.mustRun("resolve", link).output)
	}
}

func TestResolveRefusesABareLinkThatMatchesTwoIssues(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/fix-login-bug")
	r.mustRun("new", "mobile/fix-login-bug")

	res := r.run("resolve", "[[fix-login-bug]]")

	equal(t, 1, res.exit)
	contains(t, res.error, "is ambiguous")
}
