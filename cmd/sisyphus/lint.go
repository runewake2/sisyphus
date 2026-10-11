package main

import (
	"fmt"
	"slices"
	"strings"
)

// lintProblem is one problem that sisyphus lint found: the file (relative to the repo root), the
// line for a wikilink problem, and what is wrong.
type lintProblem struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Problem string `json:"problem"`
}

func (p lintProblem) String() string {
	if p.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", p.File, p.Line, p.Problem)
	}
	return fmt.Sprintf("%s: %s", p.File, p.Problem)
}

// lintResult is what sisyphus lint checked and what it found.
type lintResult struct {
	Issues    int           `json:"issues"`
	Documents int           `json:"documents"`
	Problems  []lintProblem `json:"problems"`
}

// lintRepo checks every issue and every Markdown file of the repo, and never writes. For each
// issue: the file can be read, its frontmatter has a valid shape and valid values (issueProblems),
// its full name is in one state directory only, its tags have no spaces, its parent exists with no
// cycle above it, and each depends-on issue exists with no cycle. For each Markdown file: every
// wikilink resolves to one file, and to an existing heading when it names one.
func lintRepo(root string) lintResult {
	x := loadIndex(root)
	var problems []lintProblem
	add := func(file string, line int, format string, args ...any) {
		problems = append(problems, lintProblem{File: file, Line: line, Problem: fmt.Sprintf(format, args...)})
	}

	for _, issue := range x.issues {
		file := relative(root, issue.path)
		doc, err := x.document(issue)
		if err != nil {
			add(file, 0, "the file cannot be read: %s", err)
			continue
		}
		for _, problem := range issueProblems(issue, doc) {
			add(file, 0, "%s", problem)
		}
		if copies := x.exact[issue.name]; len(copies) > 1 && copies[0].path == issue.path {
			var paths []string
			for _, c := range copies {
				paths = append(paths, relative(root, c.path))
			}
			add(file, 0, "the issue '%s' is in more than one state directory: %s", issue.name, strings.Join(paths, ", "))
		}
		if err := validateTags(doc.get("tags")); err != nil {
			add(file, 0, "%s", strings.TrimSuffix(lowerFirst(err.Error()), "."))
		}
		lintParent(x, issue, doc, func(format string, args ...any) { add(file, 0, format, args...) })
		lintDependsOn(x, issue, doc, func(format string, args ...any) { add(file, 0, format, args...) })
	}

	checker := newLinkChecker(root)
	documents := 0
	for _, c := range checker.candidates {
		if !hasSuffixFold(c.relative, ".md") {
			continue
		}
		documents++
		rows, err := checker.rows(c.file, false)
		if err != nil {
			add(c.relative, 0, "the file cannot be read: %s", err)
			continue
		}
		for _, row := range rows {
			if row.Status == "ok" {
				continue
			}
			detail := ""
			if row.Resolved != "" {
				detail = " (" + row.Resolved + ")"
			}
			add(c.relative, row.Line, "the link %s is %s%s", row.Link, row.Status, detail)
		}
	}

	slices.SortStableFunc(problems, func(a, b lintProblem) int {
		if d := strings.Compare(a.File, b.File); d != 0 {
			return d
		}
		return a.Line - b.Line
	})
	if problems == nil {
		problems = []lintProblem{}
	}
	return lintResult{Issues: len(x.issues), Documents: documents, Problems: problems}
}

// lintParent reports a parent that does not name one existing issue, and a parent chain that comes
// back to the issue.
func lintParent(x *issueIndex, issue issueFile, doc *document, report func(string, ...any)) {
	value := doc.get("parent")
	if value == "" {
		return
	}
	match, found, err := x.match(issueName(value))
	switch {
	case err != nil:
		report("its parent %s: %s", value, lowerFirst(strings.TrimSuffix(err.Error(), ".")))
		return
	case !found:
		report("its parent %s does not exist", value)
		return
	}
	seen := map[string]bool{issue.name: true}
	for ancestor := match.name; ancestor != ""; ancestor = parentName(x, ancestor) {
		if seen[ancestor] {
			report("its parent chain has a cycle: it comes back to '%s'", ancestor)
			return
		}
		seen[ancestor] = true
	}
}

// lintDependsOn reports a depends-on entry that does not name one existing issue, a dependency on
// the issue itself, and a dependency cycle through the issue.
func lintDependsOn(x *issueIndex, issue issueFile, doc *document, report func(string, ...any)) {
	for _, dependency := range parseDependsOn(doc.get("depends-on")) {
		match, found, err := x.match(dependency)
		switch {
		case err != nil:
			report("its depends-on entry '%s': %s", dependency, lowerFirst(strings.TrimSuffix(err.Error(), ".")))
		case !found:
			report("its depends-on entry '%s' does not exist", dependency)
		case match.name == issue.name:
			report("it depends on itself")
		case dependsOnReaches(x, match.name, issue.name, map[string]bool{}):
			report("its dependencies form a cycle: '%s' depends back on it", match.name)
		}
	}
}

// lowerFirst lowercases the first letter of an error message, so that it reads inside a sentence.
func lowerFirst(text string) string {
	if text == "" {
		return text
	}
	return strings.ToLower(text[:1]) + text[1:]
}
