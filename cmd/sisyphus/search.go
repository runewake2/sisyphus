package main

import (
	"path/filepath"
	"slices"
	"strings"
)

// searchRow is one matching issue in "sisyphus search".
type searchRow struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Priority string   `json:"priority"`
	Owner    string   `json:"owner,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Section  string   `json:"section,omitempty"`
	Content  string   `json:"content"`
}

// searchOptions holds the filters and query of "sisyphus search". filters are the same frontmatter
// filters sisyphus list uses. query is the optional free-text query. section, if given, restricts
// both the query (when one is given) and the returned content to that section of the body.
type searchOptions struct {
	filters listOptions
	query   string
	section string
}

// searchIssues finds every issue matching o, sorted the same way sisyphus list is.
func searchIssues(root string, o searchOptions) []searchRow {
	query := strings.ToLower(o.query)

	rows := []searchRow{}
	for _, m := range matchingIssues(root, o.filters) {
		var haystack, content string
		if o.section != "" {
			section, ok := sectionIn(m.doc.body, o.section)
			if !ok {
				continue
			}
			content = strings.TrimSpace(strings.Join(section, "\n"))
			haystack = content
		} else {
			content = strings.TrimSpace(strings.Join(m.doc.body, "\n"))
			haystack = m.doc.get("title") + "\n" + content
		}
		if query != "" && !strings.Contains(strings.ToLower(haystack), query) {
			continue
		}
		rows = append(rows, searchRow{
			Name:     trimMarkdownExtension(filepath.Base(m.file)),
			Title:    m.doc.get("title"),
			State:    m.doc.get("state"),
			Priority: m.doc.get("priority"),
			Owner:    m.doc.get("owner"),
			Tags:     parseList(m.doc.get("tags")),
			Section:  o.section,
			Content:  content,
		})
	}
	slices.SortFunc(rows, func(a, b searchRow) int {
		return compareByStatePriorityName(a.State, b.State, a.Priority, b.Priority, a.Name, b.Name)
	})
	return rows
}
