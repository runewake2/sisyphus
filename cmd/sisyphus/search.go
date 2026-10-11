package main

import (
	"fmt"
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

// searchIssues finds every issue matching o, sorted the same way sisyphus list is. A section that
// neither issues/TEMPLATE.md nor any filtered issue has is an error, because it is almost always a
// typo: without the error, the search would only find nothing.
func searchIssues(root string, o searchOptions) ([]searchRow, error) {
	query := strings.ToLower(o.query)

	x := loadIndex(root)
	o.filters.withBody = true
	if err := checkParentFilter(x, o.filters); err != nil {
		return nil, err
	}
	rows := []searchRow{}
	sectionFound := false
	for _, m := range matchingIssues(x, o.filters) {
		var haystack, content string
		if o.section != "" {
			section, ok := sectionIn(m.doc.body, o.section)
			if !ok {
				continue
			}
			sectionFound = true
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
			Name:     m.name,
			Title:    m.doc.get("title"),
			State:    m.doc.get("state"),
			Priority: m.doc.get("priority"),
			Owner:    m.doc.get("owner"),
			Tags:     parseList(m.doc.get("tags")),
			Section:  o.section,
			Content:  content,
		})
	}
	if o.section != "" && !sectionFound {
		if err := checkTemplateSection(root, o.section); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(rows, func(a, b searchRow) int {
		return compareByStatePriorityName(a.State, b.State, a.Priority, b.Priority, a.Name, b.Name)
	})
	return rows, nil
}

// checkTemplateSection returns an error that lists the sections of issues/TEMPLATE.md when the
// template has no section named heading. A template that cannot be read gives no error.
func checkTemplateSection(root, heading string) error {
	template, err := loadDocument(filepath.Join(root, "issues", "TEMPLATE.md"))
	if err != nil {
		return nil
	}
	if start, _, ok := sectionBounds(template.body, heading); ok && headingLevel(template.body[start]) >= 2 {
		return nil
	}
	return fmt.Errorf("No issue and no section of issues/TEMPLATE.md is named '%s'. The sections are: %s.",
		heading, strings.Join(sectionNames(template.body), ", "))
}
