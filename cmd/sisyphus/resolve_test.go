package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveResolvesALink(t *testing.T) {
	cases := []struct{ link, expected string }{
		{"widget-scheduler", "design/widget-scheduler.md"},
		{"[[widget-scheduler]]", "design/widget-scheduler.md"},
		{"[[widget-scheduler#Status API|the API]]", "design/widget-scheduler.md"},
		{"[[design/example/README]]", "design/example/README.md"},
		{"[[example/README]]", "design/example/README.md"},
		{"[[Design/Example/readme]]", "design/example/README.md"},
		{"[[WIDGET-SCHEDULER]]", "design/widget-scheduler.md"},
		{"[[basic-plan.yaml]]", "design/example/plan/basic-plan.yaml"},
		{"[[0.0.0]]", "changelog/0.0.0.md"},
	}
	for _, c := range cases {
		t.Run(c.link, func(t *testing.T) {
			r := newTestRepo(t)

			res := r.run("resolve", c.link)

			equal(t, 0, res.exit)
			equal(t, "", res.error)
			equalSlices(t, []string{c.expected}, res.lines())
		})
	}
}

func TestResolvePrefersAnExactPathFromTheRepoRoot(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("resolve", "[[README]]")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"README.md"}, res.lines())
}

func TestResolveReportsAnAmbiguousLink(t *testing.T) {
	r := newTestRepo(t)
	r.write("a/dup-doc.md", "# A\n")
	r.write("b/dup-doc.md", "# B\n")

	res := r.run("resolve", "dup-doc")

	equal(t, 1, res.exit)
	contains(t, res.error, "is ambiguous")
	contains(t, res.error, "a/dup-doc.md, b/dup-doc.md")
	contains(t, res.error, "[[a/dup-doc]]")
}

func TestResolvePrintsEveryMatchWithAll(t *testing.T) {
	r := newTestRepo(t)
	r.write("a/dup-doc.md", "# A\n")
	r.write("b/dup-doc.md", "# B\n")

	res := r.run("resolve", "dup-doc", "--all")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"a/dup-doc.md", "b/dup-doc.md"}, res.lines())
}

func TestResolvePrintsAbsolutePaths(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("resolve", "widget-scheduler", "--absolute")

	equal(t, 0, res.exit)
	want, _ := filepath.Abs(r.pathOf("design/widget-scheduler.md"))
	got, _ := filepath.Abs(strings.TrimSpace(res.output))
	equal(t, want, got)
}

func TestResolveFindsAnIssueAfterItChangesState(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "moving-issue-test")
	r.mustRun("update", "moving-issue-test", "in-progress", "--bookmark", "samw/ai/work")

	res := r.run("resolve", "[[moving-issue-test]]")

	equalSlices(t, []string{"issues/in-progress/moving-issue-test.md"}, res.lines())
}

func TestResolveReportsAMissingFile(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("resolve", "[[nothing-here]]")

	equal(t, 1, res.exit)
	contains(t, res.error, "No file matches")
}

func TestResolveReportsALinkWithoutATarget(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("resolve", "[[]]")

	equal(t, 1, res.exit)
	contains(t, res.error, "does not contain a link target")
}

func TestResolveIgnoresBuildOutput(t *testing.T) {
	r := newTestRepo(t)
	r.write("commands/tool/bin/Debug/widget-scheduler.md", "# Copy\n")
	r.write("commands/tool/obj/widget-scheduler.md", "# Copy\n")

	res := r.run("resolve", "widget-scheduler")

	equalSlices(t, []string{"design/widget-scheduler.md"}, res.lines())
}
