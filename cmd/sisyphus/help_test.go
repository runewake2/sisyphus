package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/pflag"
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

// pageSections are the level-2 headings of every command page, in order.
var pageSections = []string{"Usage", "Flags", "Behavior", "Output", "Examples"}

func TestEveryCommandPageHasTheSameSectionsAndItsFlags(t *testing.T) {
	root := newRootCommand(findRoot)
	for _, cmd := range root.Commands() {
		if cmd.Name() == "help" || cmd.Hidden {
			continue
		}
		page, err := os.ReadFile(filepath.Join("..", "..", "docs", "sisyphus-"+cmd.Name()+".md"))
		if err != nil {
			t.Errorf("%s: %v", cmd.Name(), err)
			continue
		}
		lines := splitLines(string(page))
		var sections []string
		flagsText := ""
		inFlags := false
		for _, index := range outsideCodeFences(lines) {
			line := lines[index]
			if strings.HasPrefix(line, "## ") {
				sections = append(sections, strings.TrimPrefix(line, "## "))
				inFlags = line == "## Flags"
			} else if inFlags {
				flagsText += line + "\n"
			}
		}
		equalSlices(t, pageSections, sections)
		cmd.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Name != "help" && !flag.Hidden && !strings.Contains(flagsText, "`--"+flag.Name+"`") {
				t.Errorf("docs/sisyphus-%s.md: --%s is not in the Flags table", cmd.Name(), flag.Name)
			}
		})
	}
}

// TestReadmeLinksWorkOnGitHub checks the README, which GitHub shows: GitHub does not render a
// wikilink, so the README has none outside code, and each relative Markdown link names a file.
func TestReadmeLinksWorkOnGitHub(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(content))
	for _, link := range wikilinks(lines) {
		t.Errorf("README.md:%d: the wikilink %s does not render on GitHub; use a Markdown link", link.line, link.link)
	}
	markdownLink := regexp.MustCompile(`\]\(([^)#]+)(#[^)]*)?\)`)
	for _, index := range outsideCodeFences(lines) {
		for _, match := range markdownLink.FindAllStringSubmatch(lines[index], -1) {
			target := match[1]
			if strings.Contains(target, "://") {
				continue
			}
			if _, err := os.Stat(filepath.Join("..", "..", target)); err != nil {
				t.Errorf("README.md:%d: the link target %s does not exist", index+1, target)
			}
		}
	}
}

// TestTheAgentGuideUsesCommandsAndFlagsThatExist checks each command in docs/agents.md, in inline code
// and on indented lines, because agents copy that text into their instructions.
func TestTheAgentGuideUsesCommandsAndFlagsThatExist(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents.md"))
	if err != nil {
		t.Fatal(err)
	}
	var commands []string
	for _, span := range regexp.MustCompile("`([^`\n]+)`").FindAllStringSubmatch(string(page), -1) {
		commands = append(commands, span[1])
	}
	for _, line := range strings.Split(string(page), "\n") {
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(line, "    ") && strings.HasPrefix(trimmed, "sisyphus ") {
			commands = append(commands, trimmed)
		}
	}
	root := newRootCommand(findRoot)
	checked := 0
	for _, command := range commands {
		words := strings.Fields(command)
		if len(words) < 2 || words[0] != "sisyphus" {
			continue
		}
		cmd, _, err := root.Find(words[1:2])
		if err != nil || cmd == root {
			t.Errorf("%q: no command %s", command, words[1])
			continue
		}
		checked++
		for _, word := range words[2:] {
			if !strings.HasPrefix(word, "-") {
				continue
			}
			name := strings.SplitN(word, "=", 2)[0]
			var flag *pflag.Flag
			if long, ok := strings.CutPrefix(name, "--"); ok {
				flag = cmd.Flags().Lookup(long)
				if flag == nil {
					flag = cmd.InheritedFlags().Lookup(long)
				}
			} else {
				flag = cmd.Flags().ShorthandLookup(strings.TrimPrefix(name, "-"))
			}
			if flag == nil {
				t.Errorf("%q: %s has no flag %s", command, words[1], name)
			}
		}
	}
	if checked < 10 {
		t.Errorf("checked only %d commands; the test does not find them", checked)
	}
}
