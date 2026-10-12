package main

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// listRow is one issue in "sisyphus list".
type listRow struct {
	Name     string            `json:"name"`
	Title    string            `json:"title"`
	State    string            `json:"state"`
	Priority string            `json:"priority"`
	Owner    string            `json:"owner,omitempty"`
	Bookmark string            `json:"bookmark,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// metadataFilter is one --metadata key=value of list or search. any is true for a key with no value.
type metadataFilter struct {
	key, value string
	any        bool
}

// parseMetadataFilters reads the values of --metadata: key=value, or key alone for any value.
func parseMetadataFilters(values []string) ([]metadataFilter, error) {
	var filters []metadataFilter
	for _, value := range values {
		key, val, hasValue := strings.Cut(value, "=")
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("--metadata %q has no key. Use key=value, or key for any value, for example --metadata delegate=merge-queue.", value)
		}
		filters = append(filters, metadataFilter{key: key, value: strings.TrimSpace(val), any: !hasValue})
	}
	return filters, nil
}

// matchesMetadata reports whether metadata has every filter.
func matchesMetadata(metadata map[string]string, filters []metadataFilter) bool {
	for _, f := range filters {
		value, ok := metadata[f.key]
		if !ok || !f.any && value != f.value {
			return false
		}
	}
	return true
}

// listOptions holds the filters of "sisyphus list". An empty slice or string means the filter is not given.
// states and priorities narrow to one of the given values; tags narrows to any of the given tags. owner and
// parent narrow to an exact match. blockedOnly narrows to issues with a dependency that is not yet closed.
// readyOnly narrows to the issues that sisyphus graph calls available: not closed, and with no
// dependency and no sub-issue that is not closed, so work on them can start now. It also leaves out
// an issue that another workspace has claimed (see claimStore).
// Every given filter applies together (AND).
// warnings, when not nil, gets one warning for each issue in the state filter that has an invalid
// file (see issueProblems), so that bad data is not shown as if it were good.
type listOptions struct {
	states, priorities, tags []string
	owner, parent            string
	// metadata narrows to issues whose metadata has every pair; a pair with no value matches any
	// value of its key.
	metadata               []metadataFilter
	blockedOnly, readyOnly bool
	warnings               io.Writer
	// withBody reads each matched issue in full; without it, only the frontmatter is read.
	withBody bool
}

// nonEmpty returns m, or nil when it is empty, so that JSON leaves it out.
func nonEmpty(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	return m
}

// isBlocked reports whether doc depends on any issue that is not closed.
func isBlocked(x *issueIndex, doc *document) bool {
	return len(openDependencies(x, doc)) > 0
}

// openDependencies returns the full names of the depends-on issues of doc that are not closed.
func openDependencies(x *issueIndex, doc *document) []string {
	var open []string
	for _, dep := range parseDependsOn(doc.get("depends-on")) {
		match, found, err := x.match(dep)
		if err == nil && found && match.state != "closed" {
			open = append(open, match.name)
		}
	}
	return open
}

// parentsWithOpenSubIssues returns the full names of the issues that have a sub-issue that is not
// closed. Such an issue waits on its sub-issues, so it is not ready.
func parentsWithOpenSubIssues(x *issueIndex) map[string]bool {
	parents := map[string]bool{}
	x.preload(x.all(), false)
	for _, issue := range x.all() {
		if issue.state == "closed" {
			continue
		}
		if doc, err := x.header(issue); err == nil {
			if parent := x.canonical(doc.get("parent")); parent != "" {
				parents[parent] = true
			}
		}
	}
	return parents
}

// matchedIssue is one issue that matched a set of frontmatter filters, before formatting as a row.
type matchedIssue struct {
	name string
	doc  *document
}

// matchingIssues returns every issue matching o's frontmatter filters, unsorted. With no state
// filter, only open and in-progress issues are matched.
func matchingIssues(x *issueIndex, o listOptions) []matchedIssue {
	stateFilter := o.states
	if len(stateFilter) == 0 {
		stateFilter = []string{"open", "in-progress"}
	}
	parentFilter := ""
	if o.parent != "" {
		parentFilter = x.canonical(o.parent)
	}
	var waitsOnSubIssues map[string]bool
	var claims claimStore
	if o.readyOnly {
		waitsOnSubIssues = parentsWithOpenSubIssues(x)
		claims = openClaims(x.root)
	}

	var inStates []issueFile
	for _, issue := range x.all() {
		if slices.Contains(stateFilter, issue.state) {
			inStates = append(inStates, issue)
		}
	}
	x.preload(inStates, o.withBody)
	var matches []matchedIssue
	for _, issue := range inStates {
		load := x.header
		if o.withBody {
			load = x.document
		}
		doc, err := load(issue)
		warnIssueProblems(x, issue, doc, err, o.warnings)
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
		if len(o.metadata) > 0 && !matchesMetadata(parseMetadata(doc.get("metadata")), o.metadata) {
			continue
		}
		if parentFilter != "" && x.canonical(doc.get("parent")) != parentFilter {
			continue
		}
		if o.blockedOnly && !isBlocked(x, doc) {
			continue
		}
		if o.readyOnly && (issue.state == "closed" || isBlocked(x, doc) || waitsOnSubIssues[issue.name]) {
			continue
		}
		if o.readyOnly && issue.state != "in-progress" {
			// Another workspace started the issue, and its work has not merged yet.
			if _, claimed := claims.claimedElsewhere(issue.name); claimed {
				continue
			}
		}
		matches = append(matches, matchedIssue{name: issue.name, doc: doc})
	}
	return matches
}

// checkParentFilter fails when o.parent is given and does not name exactly one issue, so that a
// typo is not shown as an epic with no sub-issues.
func checkParentFilter(x *issueIndex, o listOptions) error {
	if o.parent == "" {
		return nil
	}
	_, err := x.find(issueName(o.parent))
	return err
}

// listIssues returns every issue that matches o, sorted by state, then priority, then name.
func listIssues(root string, o listOptions) ([]listRow, error) {
	x := loadIndex(root)
	if err := checkParentFilter(x, o); err != nil {
		return nil, err
	}
	matches := matchingIssues(x, o)
	rows := make([]listRow, 0, len(matches))
	for _, m := range matches {
		rows = append(rows, listRow{
			Name:     m.name,
			Title:    m.doc.get("title"),
			State:    m.doc.get("state"),
			Priority: m.doc.get("priority"),
			Owner:    m.doc.get("owner"),
			Bookmark: m.doc.get("bookmark"),
			Tags:     parseList(m.doc.get("tags")),
			Metadata: nonEmpty(parseMetadata(m.doc.get("metadata"))),
		})
	}
	slices.SortFunc(rows, func(a, b listRow) int {
		return compareByStatePriorityName(a.State, b.State, a.Priority, b.Priority, a.Name, b.Name)
	})
	return rows, nil
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

// anyOf reports whether haystack contains any element of needles, without regard to case. It
// compares tags, which Obsidian treats without regard to case.
func anyOf(haystack, needles []string) bool {
	for _, needle := range needles {
		if containsFold(haystack, needle) {
			return true
		}
	}
	return false
}
