package main

import (
	"path/filepath"
	"slices"
	"strings"
)

// listRow is one issue in "sisyphus list".
type listRow struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Priority string   `json:"priority"`
	Owner    string   `json:"owner,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// listOptions holds the filters of "sisyphus list". An empty slice or string means the filter is not given.
// states and priorities narrow to one of the given values; tags narrows to any of the given tags. owner and
// parent narrow to an exact match. blockedOnly narrows to issues with a dependency that is not yet closed.
// Every given filter applies together (AND).
type listOptions struct {
	states, priorities, tags []string
	owner, parent            string
	blockedOnly              bool
}

// isBlocked reports whether doc depends on any issue that is not closed.
func isBlocked(root string, doc *document) bool {
	for _, dep := range parseDependsOn(doc.get("depends-on")) {
		matches := issueMatches(root, dep)
		if len(matches) == 1 && matches[0].state != "closed" {
			return true
		}
	}
	return false
}

// matchedIssue is one issue that matched a set of frontmatter filters, before formatting as a row.
type matchedIssue struct {
	file string
	doc  *document
}

// matchingIssues returns every issue matching o's frontmatter filters, unsorted. With no state
// filter, only open and in-progress issues are matched.
func matchingIssues(root string, o listOptions) []matchedIssue {
	stateFilter := o.states
	if len(stateFilter) == 0 {
		stateFilter = []string{"open", "in-progress"}
	}
	parentFilter := ""
	if o.parent != "" {
		parentFilter = issueName(o.parent)
	}

	var matches []matchedIssue
	for _, state := range stateFilter {
		files, _ := filepath.Glob(filepath.Join(root, "issues", state, "*.md"))
		for _, file := range files {
			doc, err := loadDocument(file)
			if err != nil {
				continue
			}
			if len(o.priorities) > 0 && !slices.Contains(o.priorities, doc.get("priority")) {
				continue
			}
			if len(o.tags) > 0 && !anyOf(parseList(doc.get("tags")), o.tags) {
				continue
			}
			if o.owner != "" && doc.get("owner") != o.owner {
				continue
			}
			if parentFilter != "" && issueName(doc.get("parent")) != parentFilter {
				continue
			}
			if o.blockedOnly && !isBlocked(root, doc) {
				continue
			}
			matches = append(matches, matchedIssue{file: file, doc: doc})
		}
	}
	return matches
}

// listIssues returns every issue that matches o, sorted by state, then priority, then name.
func listIssues(root string, o listOptions) []listRow {
	matches := matchingIssues(root, o)
	rows := make([]listRow, 0, len(matches))
	for _, m := range matches {
		rows = append(rows, listRow{
			Name:     trimMarkdownExtension(filepath.Base(m.file)),
			Title:    m.doc.get("title"),
			State:    m.doc.get("state"),
			Priority: m.doc.get("priority"),
			Owner:    m.doc.get("owner"),
			Tags:     parseList(m.doc.get("tags")),
		})
	}
	slices.SortFunc(rows, func(a, b listRow) int {
		return compareByStatePriorityName(a.State, b.State, a.Priority, b.Priority, a.Name, b.Name)
	})
	return rows
}

// compareByStatePriorityName orders issues by state (open, in-progress, closed), then priority
// (critical..low), then name, the sort sisyphus list and sisyphus search both use.
func compareByStatePriorityName(stateA, stateB, priorityA, priorityB, nameA, nameB string) int {
	if d := slices.Index(states, stateA) - slices.Index(states, stateB); d != 0 {
		return d
	}
	if d := slices.Index(priorities, priorityA) - slices.Index(priorities, priorityB); d != 0 {
		return d
	}
	return strings.Compare(nameA, nameB)
}

// anyOf reports whether haystack contains any element of needles.
func anyOf(haystack, needles []string) bool {
	for _, needle := range needles {
		if slices.Contains(haystack, needle) {
			return true
		}
	}
	return false
}
