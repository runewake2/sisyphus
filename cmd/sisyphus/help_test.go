package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryCommandHasExamplesFromItsDocsPage(t *testing.T) {
	root := newRootCommand(findRoot)
	for _, cmd := range root.Commands() {
		if cmd.Name() == "help" || cmd.Hidden {
			continue
		}
		examples := commandExamples[cmd.Name()]
		if len(examples) < 2 || len(examples) > 4 {
			t.Errorf("%s has %d examples, not 2-4", cmd.Name(), len(examples))
		}
		page, err := os.ReadFile(filepath.Join("..", "..", "docs", "sisyphus-"+cmd.Name()+".md"))
		if err != nil {
			t.Errorf("%s: %v", cmd.Name(), err)
			continue
		}
		for _, example := range examples {
			if !strings.Contains(string(page), example+"\n") {
				t.Errorf("docs/sisyphus-%s.md has no line %q", cmd.Name(), example)
			}
		}
	}
}

func TestRootHelpGroupsTheCommandsOnLinesThatFit(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("--help")

	equal(t, 0, res.exit)
	for _, group := range commandGroups {
		if !strings.Contains(res.output, group.title) {
			t.Errorf("no group %q in the help", group.title)
		}
	}
	if strings.Contains(res.output, "Additional Commands") {
		t.Errorf("a command has no group:\n%s", res.output)
	}
	for _, line := range strings.Split(res.output, "\n") {
		// The command and flag lines; cobra does not wrap the description paragraph.
		if strings.HasPrefix(line, "  ") && len(line) > 80 {
			t.Errorf("help line longer than 80 characters: %q", line)
		}
	}
}
