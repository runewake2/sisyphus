package main

import (
	"fmt"
	"maps"
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

	files := map[string]string{}
	parents := map[string]string{}
	dependencies := map[string][]string{}
	// Read each issue in full once: the issue checks use the frontmatter, and the link check below uses
	// the same lines instead of reading the file again.
	x.preload(x.all(), true)
	for _, issue := range x.all() {
		file := relative(root, issue.path)
		doc, err := x.document(issue)
		if err != nil {
			add(file, 0, "the file cannot be read: %s", err)
			continue
		}
		for _, problem := range issueProblems(issue, doc) {
			add(file, 0, "%s", problem)
		}
		if copies := x.exactOf(issue.name); len(copies) > 1 && copies[0].path == issue.path {
			var paths []string
			for _, c := range copies {
				paths = append(paths, relative(root, c.path))
			}
			add(file, 0, "the issue '%s' is in more than one state directory: %s", issue.name, strings.Join(paths, ", "))
		}
		if err := validateTags(doc.get("tags")); err != nil {
			add(file, 0, "%s", strings.TrimSuffix(lowerFirst(err.Error()), "."))
		}
		report := func(format string, args ...any) { add(file, 0, format, args...) }
		files[issue.name] = file
		if parent := lintParent(x, doc, report); parent != "" {
			parents[issue.name] = parent
		}
		dependencies[issue.name] = lintDependsOn(x, issue, doc, report)
	}
	// One pass each over the whole graph finds every cycle, and reports it once, on the file of its
	// first issue. A search from every issue would be quadratic on a long chain.
	for _, cycle := range parentCycles(parents) {
		add(files[cycle[0]], 0, "its parent chain is a cycle: %s", strings.Join(append(cycle, cycle[0]), " → "))
	}
	for _, cycle := range dependencyCycles(dependencies) {
		add(files[cycle[0]], 0, "its dependencies form a cycle: %s", strings.Join(append(cycle, cycle[0]), " → "))
	}

	checker := newLinkChecker(root)
	documents := 0
	for _, c := range checker.candidates.list {
		// A nested tracker checks only its own documents; their links resolve in the whole repo.
		if !hasSuffixFold(c.relative, ".md") || outside(c.relative) {
			continue
		}
		documents++
		var rows []linkRow
		if loaded, ok := x.docs[c.file]; ok && loaded.err == nil && !loaded.doc.headerOnly {
			rows = checker.rowsOf(c.file, loaded.doc.lines(), false)
		} else {
			var err error
			if rows, err = checker.rows(c.file, false); err != nil {
				add(c.relative, 0, "the file cannot be read: %s", err)
				continue
			}
		}
		for _, row := range rows {
			if row.Status == "ok" {
				continue
			}
			if row.Status == "alias-in-table" {
				add(c.relative, row.Line, "the link %s is in a table and has an alias: Markdown reads its | as the end of the cell. Remove the alias", row.Link)
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
	return lintResult{Issues: len(x.all()), Documents: documents, Problems: problems}
}

// lintParent reports a parent that does not name one existing issue, and returns the full name of the
// parent that it names, or "" when there is none. Cycles are found later, for the whole graph.
func lintParent(x *issueIndex, doc *document, report func(string, ...any)) string {
	value := doc.get("parent")
	if value == "" {
		return ""
	}
	match, found, err := x.match(issueName(value))
	switch {
	case err != nil:
		report("its parent %s: %s", value, lowerFirst(strings.TrimSuffix(err.Error(), ".")))
		return ""
	case !found:
		report("its parent %s does not exist", value)
		return ""
	}
	return match.name
}

// lintDependsOn reports a depends-on entry that does not name one existing issue, and a dependency
// on the issue itself, and returns the full names of the other issues that the entries name.
func lintDependsOn(x *issueIndex, issue issueFile, doc *document, report func(string, ...any)) []string {
	var names []string
	for _, dependency := range parseDependsOn(doc.get("depends-on")) {
		match, found, err := x.match(dependency)
		switch {
		case err != nil:
			report("its depends-on entry '%s': %s", dependency, lowerFirst(strings.TrimSuffix(err.Error(), ".")))
		case !found:
			report("its depends-on entry '%s' does not exist", dependency)
		case match.name == issue.name:
			report("it depends on itself")
		default:
			names = append(names, match.name)
		}
	}
	return names
}

// parentCycles returns each cycle of the parent links once, as its issues in order, starting with the
// first name in sort order. Each issue has at most one parent, so one walk with three states finds
// every cycle in linear time.
func parentCycles(parents map[string]string) [][]string {
	const (
		unseen = iota
		onPath
		done
	)
	state := map[string]int{}
	var cycles [][]string
	for _, start := range slices.Sorted(maps.Keys(parents)) {
		var path []string
		at := map[string]int{}
		name := start
		for name != "" && state[name] == unseen {
			state[name] = onPath
			at[name] = len(path)
			path = append(path, name)
			name = parents[name]
		}
		if name != "" && state[name] == onPath {
			cycles = append(cycles, rotateToFirst(path[at[name]:]))
		}
		for _, n := range path {
			state[n] = done
		}
	}
	return cycles
}

// dependencyCycles returns one cycle for each group of issues that depend on each other, found with
// Tarjan's strongly connected components in linear time. Each cycle starts with the first name of its
// group in sort order, and is a shortest way back to it.
func dependencyCycles(edges map[string][]string) [][]string {
	index, low := map[string]int{}, map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var groups [][]string
	var visit func(string)
	visit = func(v string) {
		index[v], low[v] = len(index), len(index)
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range edges[v] {
			if _, seen := index[w]; !seen {
				visit(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] == index[v] {
			var group []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				group = append(group, w)
				if w == v {
					break
				}
			}
			if len(group) > 1 {
				groups = append(groups, group)
			}
		}
	}
	for _, v := range slices.Sorted(maps.Keys(edges)) {
		if _, seen := index[v]; !seen {
			visit(v)
		}
	}
	var cycles [][]string
	for _, group := range groups {
		cycles = append(cycles, shortestCycle(edges, group))
	}
	slices.SortFunc(cycles, func(a, b []string) int { return strings.Compare(a[0], b[0]) })
	return cycles
}

// shortestCycle returns a shortest path from the first name of group back to itself, through edges
// inside the group, without the repeated first name at the end.
func shortestCycle(edges map[string][]string, group []string) []string {
	inGroup := map[string]bool{}
	for _, name := range group {
		inGroup[name] = true
	}
	start := slices.Min(group)
	from := map[string]string{start: ""}
	queue := []string{start}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, w := range slices.Sorted(slices.Values(edges[v])) {
			if w == start {
				var path []string
				for n := v; n != ""; n = from[n] {
					path = append(path, n)
				}
				slices.Reverse(path)
				return path
			}
			if _, seen := from[w]; !seen && inGroup[w] {
				from[w] = v
				queue = append(queue, w)
			}
		}
	}
	return group
}

// rotateToFirst rotates a cycle so that it starts with its first name in sort order.
func rotateToFirst(cycle []string) []string {
	i := slices.Index(cycle, slices.Min(cycle))
	return slices.Concat(cycle[i:], cycle[:i])
}

// lowerFirst lowercases the first letter of an error message, so that it reads inside a sentence.
func lowerFirst(text string) string {
	if text == "" {
		return text
	}
	return strings.ToLower(text[:1]) + text[1:]
}
