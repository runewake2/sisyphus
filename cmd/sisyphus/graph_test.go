package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// setUpGraphIssues makes an epic with two sub-issues, where sub-b depends on sub-a, and sub-a has a
// sub-issue of its own:
//
//	epic-root-test ──► sub-a-test ──► sub-a-child-test
//	               └─► sub-b-test   (sub-a-test ┄┄► sub-b-test)
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

func graphJSON(t *testing.T, r *testRepo, args ...string) issueGraph {
	t.Helper()
	res := r.run(append([]string{"graph", "--json"}, args...)...)
	equal(t, 0, res.exit)
	var graph issueGraph
	if err := json.Unmarshal([]byte(res.output), &graph); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	return graph
}

func nodeNames(graph issueGraph) []string {
	var names []string
	for _, node := range graph.Nodes {
		names = append(names, node.Name)
	}
	return names
}

func TestGraphOfAnEpicDrawsEverythingBelowIt(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "epic-root-test")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "📍 epic-root-test [open]")
	contains(t, res.output, "sub-a-test [open]")
	contains(t, res.output, "sub-a-child-test [open]")
	contains(t, res.output, "sub-b-test [closed]")
	contains(t, res.output, "──► sub-issue   ┄┄► needed by")
	isTrue(t, !strings.Contains(res.output, "not drawn"), "nothing is hidden")
}

func TestGraphDrawsEverythingBelowAndThePathAbove(t *testing.T) {
	r := setUpGraphIssues(t)

	graph := graphJSON(t, r, "sub-a-test")

	equalSlices(t, []string{"epic-root-test", "sub-a-test", "sub-a-child-test", "sub-b-test"}, nodeNames(graph))
	equal(t, 0, graph.Hidden)
}

func TestGraphLeavesOutIssuesOffThePathAbove(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-b-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "📍 sub-b-test [closed]")
	contains(t, res.output, "sub-a-test [open]")
	contains(t, res.output, "epic-root-test [open]")
	isTrue(t, !strings.Contains(res.output, "sub-a-child-test"), "a sibling's sub-issue is not drawn")
	contains(t, res.output, "1 more linked issue is not drawn. Use --full to draw it.")
}

func TestGraphCountsEveryHiddenIssue(t *testing.T) {
	r := setUpGraphIssues(t)
	r.mustRun("new", "sub-c-test", "--parent", "epic-root-test")

	res := r.run("graph", "sub-a-child-test")

	equal(t, 0, res.exit)
	isTrue(t, !strings.Contains(res.output, "sub-b-test"), "a sibling is not drawn")
	isTrue(t, !strings.Contains(res.output, "sub-c-test"), "a sibling is not drawn")
	contains(t, res.output, "2 more linked issues are not drawn. Use --full to draw them.")
}

func TestGraphFullDrawsEveryConnectedIssue(t *testing.T) {
	r := setUpGraphIssues(t)
	r.mustRun("new", "unrelated-issue-test")

	res := r.run("graph", "--full", "sub-a-child-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "sub-b-test [closed]")
	isTrue(t, !strings.Contains(res.output, "unrelated-issue-test"), "an unconnected issue is not drawn")
	isTrue(t, !strings.Contains(res.output, "not drawn"), "nothing is hidden")
}

func TestGraphFollowsDependenciesOutsideTheFamily(t *testing.T) {
	r := setUpGraphIssues(t)
	r.mustRun("new", "outside-blocker-test")
	r.mustRun("depends-on", "sub-a-child-test", "outside-blocker-test")

	graph := graphJSON(t, r, "outside-blocker-test")

	equalSlices(t, []string{"outside-blocker-test", "sub-a-child-test"}, nodeNames(graph))
	equal(t, 3, graph.Hidden)
}

func TestGraphMarksTheRequestedIssueAsFocus(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-a-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "📍 sub-a-test [open]")
	equal(t, 2, strings.Count(res.output, "📍 sub-a-test")) // Its box and the legend.
	isTrue(t, !strings.Contains(res.output, "📍 epic-root-test"), "only the focus is marked")
}

func TestGraphPrintsMermaid(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "--mermaid", "sub-b-test")

	equal(t, 0, res.exit)
	equalSlices(t, []string{
		"graph LR",
		`    n0["epic-root-test [open]"]`,
		`    n1["sub-a-test [open]"]`,
		`    n2["📍 sub-b-test [closed]"]`,
		"    n0 --> n1",
		"    n0 --> n2",
		"    n1 -.-> n2",
		"    %% 1 more linked issue is not drawn. Use --full to draw it.",
	}, res.lines())
}

func TestGraphOfAnIssueWithNoRelationsIsOneBox(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "lonely-issue-test")

	res := r.run("graph", "lonely-issue-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "📍🍃 lonely-issue-test [open]")
	equal(t, 1, strings.Count(res.output, "┌"))
	isTrue(t, !strings.Contains(res.output, "needed by"), "no legend for edges that are not drawn")
}

func TestGraphShowsAMissingDependency(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "waiting-issue-test")
	r.write("issues/open/waiting-issue-test.md", strings.Replace(r.read("issues/open/waiting-issue-test.md"),
		"depends-on: []", `depends-on: ["[[ghost-issue-test]]"]`, 1))

	graph := graphJSON(t, r, "waiting-issue-test")

	equal(t, 2, len(graph.Nodes))
	equal(t, graphNode{Name: "ghost-issue-test", Title: "(missing)", State: "?"}, graph.Nodes[0])
	equalSlices(t, []graphEdge{{From: "ghost-issue-test", To: "waiting-issue-test", Kind: "depends-on"}}, graph.Edges)
}

func TestGraphAcceptsEveryFormOfTheName(t *testing.T) {
	r := setUpGraphIssues(t)

	for _, reference := range []string{"sub-a-test", "[[sub-a-test]]", "#sub-a-test", "issues/open/sub-a-test.md"} {
		t.Run(reference, func(t *testing.T) {
			res := r.run("graph", reference)

			equal(t, 0, res.exit)
			contains(t, res.output, "📍 sub-a-test [open]")
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

	graph := graphJSON(t, r, "--full", "sub-a-child-test")

	equal(t, "sub-a-child-test", graph.Focus)
	equal(t, 0, graph.Hidden)
	equalSlices(t, []string{"epic-root-test", "sub-a-test", "sub-a-child-test", "sub-b-test"}, nodeNames(graph))
	equalSlices(t, []graphEdge{
		{From: "epic-root-test", To: "sub-a-test", Kind: "parent"},
		{From: "epic-root-test", To: "sub-b-test", Kind: "parent"},
		{From: "sub-a-test", To: "sub-a-child-test", Kind: "parent"},
		{From: "sub-a-test", To: "sub-b-test", Kind: "depends-on"},
	}, graph.Edges)
}

func TestGraphMarksLeaves(t *testing.T) {
	r := setUpGraphIssues(t)
	r.mustRun("new", "waits-on-open-test", "--parent", "epic-root-test", "--depends-on", "sub-a-child-test")
	r.mustRun("new", "waits-on-closed-test", "--parent", "epic-root-test", "--depends-on", "sub-b-test")

	graph := graphJSON(t, r, "epic-root-test")

	leaves := map[string]bool{}
	for _, node := range graph.Nodes {
		leaves[node.Name] = node.Leaf
	}
	isTrue(t, !leaves["epic-root-test"], "an issue with open sub-issues is not a leaf")
	isTrue(t, !leaves["sub-a-test"], "an issue with an open sub-issue is not a leaf")
	isTrue(t, leaves["sub-a-child-test"], "an open issue with nothing to wait on is a leaf")
	isTrue(t, !leaves["sub-b-test"], "a closed issue is not a leaf")
	isTrue(t, !leaves["waits-on-open-test"], "an issue with an open dependency is not a leaf")
	isTrue(t, leaves["waits-on-closed-test"], "an issue whose dependencies are all closed is a leaf")
}

func TestGraphDrawsTheLeafMark(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-a-child-test")

	equal(t, 0, res.exit)
	contains(t, res.output, "📍🍃 sub-a-child-test [open]")
	contains(t, res.output, "🍃 leaf (ready to start)")
	isTrue(t, !strings.Contains(res.output, "🍃 sub-a-test"), "only leaves are marked")
}

func TestGraphLeavesOutTheLeafLegendWithoutLeaves(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-b-test")

	equal(t, 0, res.exit)
	isTrue(t, !strings.Contains(res.output, "leaf"), "no leaf is drawn")
}
