package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/runewake2/sisyphus"
)

func TestLinkTargetFindsTheTargetOfALink(t *testing.T) {
	cases := []struct{ reference, expected string }{
		{"widget-scheduler", "widget-scheduler"},
		{"[[widget-scheduler]]", "widget-scheduler"},
		{"[[widget-scheduler#Status API]]", "widget-scheduler"},
		{"[[widget-scheduler|text]]", "widget-scheduler"},
		{"[[widget-scheduler#Status API|text]]", "widget-scheduler"},
		{"#issue-name", "issue-name"},
		{"[[#Heading]]", ""},
		{"  [[design/example/README]]  ", "design/example/README"},
	}
	for _, c := range cases {
		t.Run(c.reference, func(t *testing.T) {
			equal(t, c.expected, linkTarget(c.reference))
		})
	}
}

func TestIssueNameFindsTheNameOfAnIssue(t *testing.T) {
	cases := []struct{ reference, expected string }{
		{"issue-name", "issue-name"},
		{"[[issue-name]]", "issue-name"},
		{"#issue-name", "issue-name"},
		{"issues/open/issue-name.md", "issue-name"},
		{`issues\open\issue-name.md`, "issue-name"},
	}
	for _, c := range cases {
		t.Run(c.reference, func(t *testing.T) {
			equal(t, c.expected, issueName(c.reference))
		})
	}
}

func TestDocumentParsesFrontmatterAndBody(t *testing.T) {
	doc, err := parseDocument("---\ntitle: \"A \\\"quoted\\\" title\"\nstate: open   # comment\nname: 'single'\n---\n\n# Body\n")
	if err != nil {
		t.Fatal(err)
	}

	equal(t, `A "quoted" title`, doc.get("title"))
	equal(t, "open", doc.get("state"))
	equal(t, "single", doc.get("name"))
	equal(t, "", doc.get("missing"))
	equalSlices(t, []string{"", "# Body"}, doc.body)
}

func TestDocumentParsesWindowsLineEndings(t *testing.T) {
	doc, err := parseDocument("---\r\nstate: open\r\n---\r\n# Body\r\n")
	if err != nil {
		t.Fatal(err)
	}

	equal(t, "open", doc.get("state"))
	equalSlices(t, []string{"# Body"}, doc.body)
}

func TestDocumentSetKeepsTheCommentColumn(t *testing.T) {
	doc, _ := parseDocument("---\nstate: open            # open | closed\n---\n")

	doc.set("state", "in-progress")

	equalSlices(t, []string{"state: in-progress     # open | closed"}, doc.front)
}

func TestDocumentSetKeepsOneSpaceBeforeACommentWhenTheValueIsLong(t *testing.T) {
	doc, _ := parseDocument("---\nbookmark: # comment\n---\n")

	doc.set("bookmark", "samw/ai/a-long-workspace-name")

	equalSlices(t, []string{"bookmark: samw/ai/a-long-workspace-name # comment"}, doc.front)
}

func TestDocumentSetClearsAValue(t *testing.T) {
	doc, _ := parseDocument("---\nclosed: 2026-10-06 # date\n---\n")

	doc.set("closed", "")

	equalSlices(t, []string{"closed:            # date"}, doc.front)
}

func TestDocumentSetAddsAMissingKey(t *testing.T) {
	doc, _ := parseDocument("---\nstate: open\n---\n")

	doc.set("bookmark", "samw/ai/work")

	equalSlices(t, []string{"state: open", "bookmark: samw/ai/work"}, doc.front)
}

func TestDocumentSetAddsKeysWithoutChangingTheBody(t *testing.T) {
	doc, _ := parseDocument("---\nstate: open\n---\nfirst body line\nsecond body line\n")

	doc.set("bookmark", "samw/ai/work")
	doc.set("parent", `"[[parent-issue]]"`)

	equal(t, "---\nstate: open\nbookmark: samw/ai/work\nparent: \"[[parent-issue]]\"\n---\nfirst body line\nsecond body line\n", doc.render())
}

func TestDocumentRenderRoundTrips(t *testing.T) {
	const content = "---\nstate: open # comment\n---\n\n# Body\n"
	doc, _ := parseDocument(content)

	equal(t, content, doc.render())
}

func TestDocumentRejectsBadFrontmatter(t *testing.T) {
	cases := []struct{ content, message string }{
		{"# No frontmatter\n", "does not start with YAML frontmatter"},
		{"---\nstate: open\n", "no closing '---'"},
	}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			_, err := parseDocument(c.content)

			if err == nil {
				t.Fatal("want an error")
			}
			contains(t, err.Error(), c.message)
		})
	}
}

func TestVersionPrintsTheCompiledVersion(t *testing.T) {
	r := newTestRepo(t)
	r.write("VERSION", "9.9.9\n")

	res := r.run("--version")

	equal(t, 0, res.exit)
	isTrue(t, sisyphus.Version() != "", "the compiled version is not empty")
	equal(t, sisyphus.Version(), strings.TrimSpace(res.output))
	equal(t, strings.TrimSpace(readSourceFile(t, "VERSION")), strings.TrimSpace(res.output))
}

func TestVersionWorksOutsideARepo(t *testing.T) {
	var output, errors bytes.Buffer
	notARepo := func() (string, error) { return "", fmt.Errorf("not a repo") }

	exit := run([]string{"--version"}, notARepo, &output, &errors)

	equal(t, 0, exit)
	equal(t, sisyphus.Version(), strings.TrimSpace(output.String()))
}
