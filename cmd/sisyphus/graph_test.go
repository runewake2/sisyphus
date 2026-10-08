package main

import (
	"encoding/json"
	"slices"
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
	equal(t, '║', borderOf(t, res.output, "epic-root-test [open]"))
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
	equal(t, '║', borderOf(t, res.output, "sub-b-test [closed]"))
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
	equal(t, '║', borderOf(t, res.output, "sub-a-test [open]"))
	equal(t, '│', borderOf(t, res.output, "epic-root-test [open]"))
	equal(t, 1, strings.Count(res.output, "╔═╗ sub-a-test"))
	equal(t, 2, strings.Count(res.output, "╔")) // The focus box and the legend.
}

func TestGraphPrintsMermaid(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "--mermaid", "sub-b-test")

	equal(t, 0, res.exit)
	equalSlices(t, []string{
		"graph LR",
		`    n0["epic-root-test [open]"]`,
		`    n1["sub-a-test [open]"]`,
		`    n2["sub-b-test [closed]"]`,
		"    n0 --> n1",
		"    n0 --> n2",
		"    n1 -.-> n2",
		"    classDef focus stroke-width:5px",
		"    class n2 focus",
		"    %% 1 more linked issue is not drawn. Use --full to draw it.",
	}, res.lines())
}

func TestGraphOfAnIssueWithNoRelationsIsOneBox(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "lonely-issue-test")

	res := r.run("graph", "lonely-issue-test")

	equal(t, 0, res.exit)
	equal(t, '║', borderOf(t, res.output, "lonely-issue-test [open]"))
	contains(t, res.output, "╔═╗ lonely-issue-test (available)")
	equal(t, 2, strings.Count(res.output, "╔")) // The box and the legend.
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
			equal(t, '║', borderOf(t, res.output, "sub-a-test [open]"))
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

func TestGraphMarksAvailableIssues(t *testing.T) {
	r := setUpGraphIssues(t)
	r.mustRun("new", "waits-on-open-test", "--parent", "epic-root-test", "--depends-on", "sub-a-child-test")
	r.mustRun("new", "waits-on-closed-test", "--parent", "epic-root-test", "--depends-on", "sub-b-test")

	graph := graphJSON(t, r, "epic-root-test")

	available := map[string]bool{}
	for _, node := range graph.Nodes {
		available[node.Name] = node.Available
	}
	isTrue(t, !available["epic-root-test"], "an issue with open sub-issues is not available")
	isTrue(t, !available["sub-a-test"], "an issue with an open sub-issue is not available")
	isTrue(t, available["sub-a-child-test"], "an open issue with nothing to wait on is available")
	isTrue(t, !available["sub-b-test"], "a closed issue is not available")
	isTrue(t, !available["waits-on-open-test"], "an issue with an open dependency is not available")
	isTrue(t, available["waits-on-closed-test"], "an issue whose dependencies are all closed is available")
}

func TestGraphDrawsAnAvailableIssueInAHeavyBox(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "epic-root-test")

	equal(t, 0, res.exit)
	equal(t, '┃', borderOf(t, res.output, "sub-a-child-test [open]"))
	equal(t, '│', borderOf(t, res.output, "sub-a-test [open]"))
	equal(t, '│', borderOf(t, res.output, "sub-b-test [closed]"))
	contains(t, res.output, "┏━┓ available (ready to start)")
}

func TestGraphDrawsAnAvailableFocusInADoubleBox(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-a-child-test")

	equal(t, 0, res.exit)
	equal(t, '║', borderOf(t, res.output, "sub-a-child-test [open]"))
	contains(t, res.output, "╔═╗ sub-a-child-test (available)")
	isTrue(t, !strings.Contains(res.output, "┏━┓ available"), "no other available issue is drawn")
}

func TestGraphLeavesOutTheAvailableLegendWithoutAvailableIssues(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "sub-b-test")

	equal(t, 0, res.exit)
	isTrue(t, !strings.Contains(res.output, "available"), "no available issue is drawn")
}

// borderOf is the left border character of the box that holds label.
func borderOf(t *testing.T, drawing, label string) rune {
	t.Helper()
	for _, line := range strings.Split(drawing, "\n") {
		if before, _, found := strings.Cut(line, label); found {
			left := []rune(strings.TrimRight(before, " "))
			return left[len(left)-1]
		}
	}
	t.Fatalf("no box holds %q:\n%s", label, drawing)
	return 0
}

// frame is the area inside the border of one directory box, in line and column numbers.
type frameArea struct {
	top, left, bottom, right int
}

// framesOf finds each directory box by its name in the top border. The drawing has no wide
// characters, so a rune is one column.
func framesOf(t *testing.T, drawing string) map[string][]frameArea {
	t.Helper()
	var grid [][]rune
	for _, line := range strings.Split(drawing, "\n") {
		grid = append(grid, []rune(line))
	}
	frames := map[string][]frameArea{}
	for y, line := range grid {
		for x := 0; x+2 < len(line); x++ {
			if line[x] != '╭' || line[x+1] != '─' || line[x+2] != ' ' {
				continue
			}
			name := strings.Fields(string(line[x+3:]))[0]
			f := frameArea{top: y, left: x}
			f.right = x + slices.Index(line[x:], '╮')
			for f.bottom = y + 1; f.bottom < len(grid) && (len(grid[f.bottom]) <= x || grid[f.bottom][x] != '╰'); f.bottom++ {
			}
			frames[name] = append(frames[name], f)
		}
	}
	return frames
}

// position is the line and column of the first rune of label.
func position(t *testing.T, drawing, label string) (int, int) {
	t.Helper()
	for y, line := range strings.Split(drawing, "\n") {
		if before, _, found := strings.Cut(line, label); found {
			return y, len([]rune(before))
		}
	}
	t.Fatalf("no %q in:\n%s", label, drawing)
	return 0, 0
}

func (f frameArea) holds(y, x int) bool {
	return y > f.top && y < f.bottom && x > f.left && x < f.right
}

func TestGraphDrawsEachDirectoryAsABox(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/login-epic")
	r.mustRun("new", "web/auth/fix-login-bug", "--parent", "login-epic")
	r.mustRun("new", "web/split-login-form", "--parent", "login-epic", "--depends-on", "fix-login-bug")
	r.mustRun("new", "mobile/add-dark-mode", "--depends-on", "web/split-login-form")

	res := r.run("graph", "--full", "login-epic")

	equal(t, 0, res.exit)
	drawing, legend, _ := strings.Cut(res.output, "\n\n")
	frames := framesOf(t, drawing)
	equal(t, 3, len(frames))
	for _, name := range []string{"web", "auth", "mobile"} {
		equal(t, 1, len(frames[name]))
	}
	contains(t, drawing, "login-epic [open]")
	isTrue(t, !strings.Contains(drawing, "web/"), "a box shows only the file name of its issue")
	contains(t, legend, "╔═╗ web/auth/login-epic")
}

// TestGraphDirectoryBoxHoldsOnlyItsIssues draws two sibling directories with links in both
// directions between them, and one directory nested in one of them.
func TestGraphDirectoryBoxHoldsOnlyItsIssues(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "plugins/integration-plugins")
	r.mustRun("new", "plugins/plugin-protocol", "--parent", "integration-plugins")
	r.mustRun("new", "plugins/plugin-runner", "--parent", "integration-plugins", "--depends-on", "plugin-protocol")
	r.mustRun("new", "plugins/github/github-plugin", "--parent", "integration-plugins", "--depends-on", "plugin-protocol")
	r.mustRun("new", "github/intake-workflow", "--parent", "integration-plugins", "--depends-on", "github-plugin")
	r.mustRun("new", "github/sync-workflow", "--parent", "integration-plugins", "--depends-on", "plugin-runner")

	res := r.run("graph", "--full", "integration-plugins")

	equal(t, 0, res.exit)
	drawing, _, _ := strings.Cut(res.output, "\n\n")
	frames := framesOf(t, drawing)
	equal(t, 2, len(frames["github"]))
	plugins := frames["plugins"][0]
	nested, top := frames["github"][0], frames["github"][1]
	if !plugins.holds(nested.top, nested.left) {
		nested, top = top, nested
	}
	isTrue(t, plugins.holds(nested.top, nested.left) && plugins.holds(nested.bottom, nested.right), "plugins/github is inside plugins")
	isTrue(t, !plugins.holds(top.top, top.left) && !plugins.holds(top.bottom, top.right), "github is outside plugins")

	for label, inside := range map[string][]frameArea{
		"integration-plugins [open]": {plugins},
		"plugin-protocol [open]":     {plugins},
		"plugin-runner [open]":       {plugins},
		"github-plugin [open]":       {plugins, nested},
		"intake-workflow [open]":     {top},
		"sync-workflow [open]":       {top},
	} {
		y, x := position(t, drawing, label)
		for _, f := range []frameArea{plugins, nested, top} {
			equal(t, slices.Contains(inside, f), f.holds(y, x))
		}
	}

	// An arrow that crosses a border joins it with a solid character, so the border stays visible.
	grid := strings.Split(drawing, "\n")
	for _, list := range frames {
		for _, f := range list {
			for y := f.top; y <= f.bottom; y++ {
				line := []rune(grid[y])
				for x := f.left; x <= f.right && x < len(line); x++ {
					if y == f.top || y == f.bottom || x == f.left || x == f.right {
						isTrue(t, line[x] != '┄' && line[x] != '┆', "a frame border is not hidden by an arrow")
					}
				}
			}
		}
	}
}

func TestGraphWithoutDirectoriesDrawsNoFrame(t *testing.T) {
	r := setUpGraphIssues(t)

	res := r.run("graph", "epic-root-test")

	equal(t, 0, res.exit)
	isTrue(t, !strings.Contains(res.output, "╭"), "no directory box")
}

func TestGraphPrintsMermaidSubgraphs(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/login-epic")
	r.mustRun("new", "web/split-login-form", "--parent", "login-epic")

	res := r.run("graph", "--mermaid", "login-epic")

	equal(t, 0, res.exit)
	equalSlices(t, []string{
		"graph LR",
		`    subgraph g0["web"]`,
		`        n1["split-login-form [open]"]`,
		`        subgraph g1["auth"]`,
		`            n0["login-epic [open]"]`,
		"        end",
		"    end",
		"    n0 --> n1",
		"    classDef focus stroke-width:5px",
		"    class n0 focus",
		"    classDef available stroke-width:3px",
		"    class n1 available",
	}, res.lines())
}

// TestGraphArrowsTakeNoDetourWithoutACycle draws an epic whose sub-issues depend on each other in
// chains. Each arrow must reach its issue with one turn at most, never by a line back to the left.
func TestGraphArrowsTakeNoDetourWithoutACycle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "integration-plugins")
	r.mustRun("new", "plugin-protocol", "--parent", "integration-plugins")
	r.mustRun("new", "plugin-runner", "--parent", "integration-plugins", "--depends-on", "plugin-protocol")
	r.mustRun("new", "github-plugin", "--parent", "integration-plugins", "--depends-on", "plugin-protocol")
	r.mustRun("new", "sync-commands", "--parent", "integration-plugins", "--depends-on", "plugin-runner")
	r.mustRun("new", "intake-workflow", "--parent", "integration-plugins", "--depends-on", "github-plugin")
	r.mustRun("depends-on", "intake-workflow", "sync-commands")

	graph := graphJSON(t, r, "--full", "integration-plugins")
	routes, _ := routeEdges(layOut(graph), graph)

	equal(t, len(graph.Edges), len(routes))
	for _, route := range routes {
		isTrue(t, route.a == nil || route.b == nil, "the arrow from "+route.edge.From+" to "+route.edge.To+" takes no detour")
		isTrue(t, route.to.column > route.from.column, "the arrow from "+route.edge.From+" points right")
	}
}
