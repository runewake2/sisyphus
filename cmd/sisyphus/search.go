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
	// Content is the section of --section, or the body with --content. Otherwise it is left out, and
	// Matches has the lines that match the query.
	Content string      `json:"content,omitempty"`
	Matches []lineMatch `json:"matches,omitempty"`
	// MoreMatches counts the matched lines after the first maxMatches.
	MoreMatches int `json:"more-matches,omitempty"`
}

// lineMatch is one line of an issue that matches the query, with the heading of its section, or
// "title" for the title.
type lineMatch struct {
	Section string `json:"section"`
	Text    string `json:"text"`
}

const (
	// maxMatches is the largest number of matched lines that search returns for one issue.
	maxMatches = 5
	// matchWidth is the largest number of characters of one matched line. A longer line is cut
	// around the match.
	matchWidth = 200
)

// searchOptions holds the filters and query of "sisyphus search". filters are the same frontmatter
// filters sisyphus list uses. query is the optional free-text query. section, if given, restricts
// both the query (when one is given) and the returned content to that section of the body. content
// returns the whole body of each match instead of the matched lines.
type searchOptions struct {
	filters listOptions
	query   string
	section string
	content bool
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
			// Template text tells an author what to write; it is not a part of the issue.
			content = strings.TrimSpace(strings.Join(withoutTemplateSections(m.doc.body), "\n"))
			haystack = m.doc.get("title") + "\n" + content
		}
		if query != "" && !strings.Contains(strings.ToLower(haystack), query) {
			continue
		}
		row := searchRow{
			Name:     m.name,
			Title:    m.doc.get("title"),
			State:    m.doc.get("state"),
			Priority: m.doc.get("priority"),
			Owner:    m.doc.get("owner"),
			Tags:     parseList(m.doc.get("tags")),
			Section:  o.section,
		}
		switch {
		case o.section != "" || o.content:
			row.Content = content
		case query != "":
			matches := matchedLines(m.doc.get("title"), strings.Split(content, "\n"), query)
			if len(matches) > maxMatches {
				row.MoreMatches = len(matches) - maxMatches
				matches = matches[:maxMatches]
			}
			row.Matches = matches
		}
		rows = append(rows, row)
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

// matchedLines returns each line of the title and the body that contains query, without regard to
// case, with the heading of its section. A query that spans two lines of the body matches no single
// line; then the lines that contain any word of the query are returned.
func matchedLines(title string, body []string, query string) []lineMatch {
	find := func(match func(line string) bool) []lineMatch {
		var matches []lineMatch
		if match(strings.ToLower(title)) {
			matches = append(matches, lineMatch{Section: "title", Text: title})
		}
		section, inFence := "", false
		for _, line := range body {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				inFence = !inFence
			}
			if heading := headingPattern.FindStringSubmatch(line); heading != nil && !inFence {
				section = heading[1]
				if headingLevel(line) == 1 {
					continue // the title again
				}
			}
			if trimmed != "" && match(strings.ToLower(line)) {
				matches = append(matches, lineMatch{Section: section, Text: clip(trimmed, query)})
			}
		}
		return matches
	}
	matches := find(func(line string) bool { return strings.Contains(line, query) })
	if len(matches) == 0 {
		words := strings.Fields(query)
		matches = find(func(line string) bool {
			return slices.ContainsFunc(words, func(word string) bool { return strings.Contains(line, word) })
		})
	}
	return matches
}

// clip returns line, or a part of it of at most matchWidth characters around the first match of
// query, with "…" where text was cut.
func clip(line, query string) string {
	runes := []rune(line)
	if len(runes) <= matchWidth {
		return line
	}
	first, at := query, 0
	if words := strings.Fields(query); len(words) > 0 {
		first = words[0]
	}
	if index := strings.Index(strings.ToLower(line), first); index >= 0 && index <= len(line) {
		at = len([]rune(line[:index]))
	}
	start := max(0, at-matchWidth/3)
	end := min(len(runes), start+matchWidth)
	start = max(0, end-matchWidth)
	clipped := string(runes[start:end])
	if start > 0 {
		clipped = "…" + clipped
	}
	if end < len(runes) {
		clipped += "…"
	}
	return clipped
}
