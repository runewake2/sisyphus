package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func setUpGraphIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "epic-root-test")
	r.mustRun("new", "sub-a-test", "--parent", "epic-root-test")
	r.mustRun("new", "sub-b-test", "--parent", "epic-root-test", "--state", "closed", "--resolution", "completed")
	r.mustRun("new", "sub-a-child-test", "--parent", "sub-a-test")
	r.mustRun("depends-on", "sub-b-test", "sub-a-test")
	return r
}

func TestGraphShowsTheWholeTreeFromTheRoot(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-a-child-test")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "epic-root-test [open]")
	contains(t, res.output, "sub-a-test [open]")
	contains(t, res.output, "sub-a-child-test [open]")
	contains(t, res.output, "sub-b-test [closed]")
}

func TestGraphMarksTheRequestedIssueAsFocus(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-a-child-test")

	equal(t, 0, res.exit)
	lines := res.lines()
	var focusLines, otherLines int
	for _, line := range lines {
		if strings.Contains(line, "sub-a-child-test") {
			contains(t, line, "you asked about this one")
			focusLines++
		} else if strings.Contains(line, "[") {
			isTrue(t, !strings.Contains(line, "you asked about this one"), "only the focus line is marked: "+line)
			otherLines++
		}
	}
	equal(t, 1, focusLines)
	isTrue(t, otherLines > 0, "other issues are printed too")
}

func TestGraphShowsDependsOnAnnotations(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "epic-root-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "sub-b-test [closed]  (depends on: sub-a-test)")
}

func TestGraphOfAnIssueWithNoRelationsIsOneLine(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "lonely-issue-test")

	res := r.run("graph", "lonely-issue-test")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"lonely-issue-test [open]  <-- you asked about this one"}, res.lines())
}

func TestGraphAcceptsEveryFormOfTheName(t *testing.T) {
	r := setUpGraphIssues(t)

	for _, reference := range []string{"sub-a-test", "[[sub-a-test]]", "#sub-a-test", "issues/open/sub-a-test.md"} {
		t.Run(reference, func(t *testing.T) {
			res := r.run("graph", reference)

			equal(t, 0, res.exit)
			contains(t, res.output, "epic-root-test [open]")
		})
	}
}

func TestGraphReportsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("graph", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'no-such-issue'")
}

func TestGraphPrintsJSON(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "--json", "epic-root-test")

	equal(t, 0, res.exit)
	var node graphNode
	if err := json.Unmarshal([]byte(res.output), &node); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	equal(t, "epic-root-test", node.Name)
	isTrue(t, node.Focus, "the root is the focus")
	equal(t, 2, len(node.Children))
	var subA, subB *graphNode
	for i := range node.Children {
		switch node.Children[i].Name {
		case "sub-a-test":
			subA = &node.Children[i]
		case "sub-b-test":
			subB = &node.Children[i]
		}
	}
	if subA == nil || subB == nil {
		t.Fatalf("expected both sub-a-test and sub-b-test, got %+v", node.Children)
	}
	equal(t, 1, len(subA.Children))
	equal(t, "sub-a-child-test", subA.Children[0].Name)
	equalSlices(t, []string{"sub-a-test"}, subB.DependsOn)
}
