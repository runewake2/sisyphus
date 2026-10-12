package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type candidate struct {
	file     string
	relative string
	keys     []string
}

type linkRow struct {
	Line     int    `json:"line"`
	Link     string `json:"link"`
	Status   string `json:"status"`
	Resolved string `json:"resolved"`
}

type foundLink struct {
	line int
	link string
}

// wikilink is a parsed link. sameFile is true for a link such as [[#Heading]], which points into its own document.
type wikilink struct {
	target   string
	heading  string
	sameFile bool
}

var (
	wikilinkPattern   = regexp.MustCompile(`\[\[([^\[\]\r\n]+)\]\]`)
	inlineCodePattern = regexp.MustCompile("`[^`]*`")
	headingPattern    = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*#*\s*$`)
)

// parseLink accepts [[target#heading|text]], #name, or a bare target.
func parseLink(reference string) wikilink {
	r := strings.TrimSpace(reference)
	if !strings.HasPrefix(r, "[[") && strings.HasPrefix(r, "#") {
		r = r[1:]
	}
	r = strings.TrimPrefix(r, "[[")
	r = strings.TrimSuffix(r, "]]")
	r, _, _ = strings.Cut(r, "|")
	target, heading, _ := strings.Cut(r, "#")
	return wikilink{
		target:   strings.TrimSpace(target),
		heading:  strings.TrimSpace(heading),
		sameFile: strings.HasPrefix(r, "#"),
	}
}

func linkTarget(reference string) string {
	return parseLink(reference).target
}

// issueName returns the issue name in reference: the link target without ".md". A path to an issue
// file loses everything up to and including issues/<state>/, so that only the full name remains.
func issueName(reference string) string {
	target := trimMarkdownExtension(strings.Trim(strings.ReplaceAll(linkTarget(reference), `\`, "/"), "/"))
	target = strings.TrimPrefix(target, "./")
	for _, state := range states {
		marker := "issues/" + state + "/"
		if i := strings.LastIndex("/"+target, "/"+marker); i >= 0 {
			return target[i+len(marker):]
		}
	}
	return target
}

// candidateSet is every file of the repo that a wikilink can point to, sorted by path, with maps that
// find the files for a link target without a scan of every file.
type candidateSet struct {
	list []candidate
	// exact maps a lowercase key (a path from the root, with or without ".md") to the indexes of the
	// candidates that have it.
	exact map[string][]int
	// suffix maps the lowercase part of a key after each "/" to the indexes of the candidates whose key
	// ends with "/" and that part, so that a target with "/" matches the end of a path.
	suffix map[string][]int
	// byRelative maps a path from the root to the index of its candidate.
	byRelative map[string]int
	// lookups counts the calls of matchTarget. The maps exact and suffix are built only after
	// scanLookups calls: for a few links, one scan of the list costs less than building them.
	lookups int
}

// scanLookups is the number of link lookups that scan the candidate list before the maps are built.
const scanLookups = 64

// repoCandidates finds every file that a wikilink of the tracker at root can point to: every file of
// the working copy that holds root, as Obsidian does with a vault at the root of the repo. A tracker
// below another one ("sisyphus init --nested") so resolves links to the documents of the whole repo.
// Paths stay relative to root: a file outside it has a path that starts with "../".
func repoCandidates(root string) *candidateSet {
	set := &candidateSet{byRelative: map[string]int{}}
	for _, file := range repoFiles(vaultRoot(root)) {
		rel := relative(root, file)
		keys := []string{rel}
		if hasSuffixFold(rel, ".md") {
			keys = append(keys, trimMarkdownExtension(rel))
		}
		set.list = append(set.list, candidate{file: file, relative: rel, keys: keys})
	}
	slices.SortFunc(set.list, func(a, b candidate) int { return strings.Compare(a.relative, b.relative) })
	for i, c := range set.list {
		set.byRelative[c.relative] = i
	}
	return set
}

// vaultRoot returns the directory whose files wikilinks resolve against: the root of the jj or git
// working copy that holds root, or root itself when there is none.
func vaultRoot(root string) string {
	if _, workingCopy, ok := findSharedState(root); ok {
		return workingCopy
	}
	return root
}

// outside reports whether a path relative to the tracker root is outside it.
func outside(relativePath string) bool {
	return relativePath == ".." || strings.HasPrefix(relativePath, "../")
}

// buildMaps fills exact and suffix, for a command that looks up many links.
func (set *candidateSet) buildMaps() {
	set.exact, set.suffix = map[string][]int{}, map[string][]int{}
	for i, c := range set.list {
		for _, key := range c.keys {
			lower := strings.ToLower(key)
			set.exact[lower] = appendIndex(set.exact[lower], i)
			for j := 0; j < len(lower); j++ {
				if lower[j] == '/' {
					set.suffix[lower[j+1:]] = appendIndex(set.suffix[lower[j+1:]], i)
				}
			}
		}
	}
}

// appendIndex appends i to indexes unless it is already the last one: the indexes of one candidate are
// added together, so this keeps each candidate once.
func appendIndex(indexes []int, i int) []int {
	if n := len(indexes); n > 0 && indexes[n-1] == i {
		return indexes
	}
	return append(indexes, i)
}

// scan finds the indexes of the candidates that target matches exactly and at the end of a path, with
// one pass over the list, for a command that looks up only a few links.
func (set *candidateSet) scan(target string) (exact, partial []int) {
	suffix := "/" + target
	for i, c := range set.list {
		if slices.ContainsFunc(c.keys, func(key string) bool { return strings.EqualFold(key, target) }) {
			exact = append(exact, i)
		}
		if slices.ContainsFunc(c.keys, func(key string) bool { return hasSuffixFold(key, suffix) }) {
			partial = append(partial, i)
		}
	}
	return exact, partial
}

// pick returns the candidates at indexes, in index order (the sorted order of paths).
func (s *candidateSet) pick(indexes []int) []candidate {
	result := make([]candidate, len(indexes))
	for i, index := range indexes {
		result[i] = s.list[index]
	}
	return result
}

// matchTarget follows Obsidian: match by file name, with or without ".md", in any directory.
// A target with "/" matches the end of a path. An exact path from the repo root wins.
func matchTarget(set *candidateSet, target string, all bool) []candidate {
	target = strings.TrimLeft(target, "/")
	if target == "" {
		return nil
	}
	lower := strings.ToLower(target)
	set.lookups++
	if set.exact == nil && set.lookups > scanLookups {
		set.buildMaps()
	}
	var exact, partial []int
	if set.exact != nil {
		exact, partial = set.exact[lower], set.suffix[lower]
	} else {
		exact, partial = set.scan(target)
	}
	if all {
		merged := slices.Concat(exact, partial)
		slices.Sort(merged)
		return set.pick(slices.Compact(merged))
	}
	if len(exact) > 0 {
		return set.pick(exact)
	}
	return set.pick(partial)
}

func resolveLink(root, link string, all, absolute bool) ([]string, error) {
	target := linkTarget(link)
	if target == "" {
		return nil, fmt.Errorf("'%s' does not contain a link target.", link)
	}
	matches := matchTarget(repoCandidates(root), target, all)
	if len(matches) == 0 {
		return nil, fmt.Errorf("No file matches '%s'.", link)
	}
	if !all && len(matches) > 1 {
		return nil, fmt.Errorf(
			"'%s' is ambiguous. It matches: %s. Add more of the path, for example [[%s]], or use --all.",
			link, joinRelative(matches), trimMarkdownExtension(matches[0].relative))
	}
	return paths(matches, absolute), nil
}

func listLinks(root, document string, absolute bool) (string, []linkRow, error) {
	checker := newLinkChecker(root)
	file, err := findDocument(checker.candidates, document)
	if err != nil {
		return "", nil, err
	}
	rows, err := checker.rows(file, absolute)
	return relative(root, file), rows, err
}

// linkChecker resolves the wikilinks of documents against one walk of the repo, and reads the
// headings of each target file once, so that sisyphus lint can check every document cheaply.
type linkChecker struct {
	root       string
	candidates *candidateSet
	headingsOf map[string][]string
}

func newLinkChecker(root string) *linkChecker {
	return &linkChecker{root: root, candidates: repoCandidates(root), headingsOf: map[string][]string{}}
}

func (c *linkChecker) hasHeading(file, heading string) bool {
	if _, ok := c.headingsOf[file]; !ok {
		c.headingsOf[file] = headings(file)
	}
	return slices.ContainsFunc(c.headingsOf[file], func(h string) bool { return strings.EqualFold(h, heading) })
}

// rows returns one row for each wikilink in file, with its status: ok, missing, ambiguous, or
// missing-heading.
func (c *linkChecker) rows(file string, absolute bool) ([]linkRow, error) {
	trace(file)
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return c.rowsOf(file, splitLines(string(content)), absolute), nil
}

// rowsOf is rows for a file whose lines the caller already has, for example from the issue index.
func (c *linkChecker) rowsOf(file string, lines []string, absolute bool) []linkRow {
	fileRelative := relative(c.root, file)
	c.headingsOf[file] = headingsIn(lines)

	rows := []linkRow{}
	for _, found := range wikilinks(lines) {
		link := parseLink(found.link)
		var matches []candidate
		if link.target == "" && link.sameFile {
			if index, ok := c.candidates.byRelative[fileRelative]; ok {
				matches = append(matches, c.candidates.list[index])
			}
		} else {
			matches = matchTarget(c.candidates, link.target, false)
		}
		status := "ambiguous"
		switch len(matches) {
		case 0:
			status = "missing"
		case 1:
			status = "ok"
		}
		if status == "ok" &&
			link.heading != "" &&
			!strings.HasPrefix(link.heading, "^") &&
			hasSuffixFold(matches[0].relative, ".md") &&
			!c.hasHeading(matches[0].file, link.heading) {
			status = "missing-heading"
		}
		rows = append(rows, linkRow{
			Line:     found.line,
			Link:     found.link,
			Status:   status,
			Resolved: strings.Join(paths(matches, absolute), ", "),
		})
	}
	return rows
}

func findDocument(candidates *candidateSet, document string) (string, error) {
	if path, err := filepath.Abs(document); err == nil && fileExists(path) {
		return path, nil
	}
	matches := matchTarget(candidates, linkTarget(document), false)
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("No file matches '%s'.", document)
	case 1:
		return matches[0].file, nil
	default:
		return "", fmt.Errorf("'%s' is ambiguous. It matches: %s.", document, joinRelative(matches))
	}
}

// wikilinks skips fenced code blocks and inline code, because Obsidian ignores links there.
func wikilinks(lines []string) []foundLink {
	var found []foundLink
	for _, index := range outsideCodeFences(lines) {
		text := inlineCodePattern.ReplaceAllString(lines[index], "")
		for _, match := range wikilinkPattern.FindAllStringSubmatch(text, -1) {
			found = append(found, foundLink{line: index + 1, link: "[[" + match[1] + "]]"})
		}
	}
	return found
}

func headings(file string) []string {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	return headingsIn(splitLines(string(content)))
}

func headingsIn(lines []string) []string {
	var result []string
	for _, index := range outsideCodeFences(lines) {
		if match := headingPattern.FindStringSubmatch(lines[index]); match != nil {
			result = append(result, match[1])
		}
	}
	return result
}

// sectionBounds finds the line range of the section named heading (matched case-insensitively, the
// same way a wikilink's #heading is matched): start is the heading line's index, end is the index of
// the next heading at the same or a shallower level, or len(lines). ok is false when no heading matches.
func sectionBounds(lines []string, heading string) (start, end int, ok bool) {
	end = len(lines)
	var level int
	for _, index := range outsideCodeFences(lines) {
		match := headingPattern.FindStringSubmatch(lines[index])
		if match == nil {
			continue
		}
		if !ok {
			if strings.EqualFold(match[1], heading) {
				start, level, ok = index, headingLevel(lines[index]), true
			}
			continue
		}
		if headingLevel(lines[index]) <= level {
			return start, index, true
		}
	}
	return start, end, ok
}

// sectionIn returns the lines of the section named heading, from just after the heading line to just
// before the next heading at the same or a shallower level, or the end of lines. ok is false when no
// heading matches.
func sectionIn(lines []string, heading string) (section []string, ok bool) {
	start, end, ok := sectionBounds(lines, heading)
	if !ok {
		return nil, false
	}
	return lines[start+1 : end], true
}

// replaceSection returns lines with the section named heading replaced by replacement, keeping the
// heading line itself and the blank line before and after the section. ok is false when no heading
// matches, in which case lines is returned unchanged.
func replaceSection(lines []string, heading string, replacement []string) (result []string, ok bool) {
	start, end, ok := sectionBounds(lines, heading)
	if !ok {
		return lines, false
	}
	result = append(result, lines[:start+1]...)
	result = append(result, "")
	result = append(result, replacement...)
	result = append(result, "")
	result = append(result, lines[end:]...)
	return result, true
}

// headingLevel counts the leading '#' characters of a heading line, for example 2 for "## Summary".
func headingLevel(line string) int {
	return len(line) - len(strings.TrimLeft(line, "#"))
}

func outsideCodeFences(lines []string) []int {
	var result []int
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if !inFence {
			result = append(result, i)
		}
	}
	return result
}

func paths(matches []candidate, absolute bool) []string {
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		if absolute {
			result = append(result, m.file)
		} else {
			result = append(result, m.relative)
		}
	}
	return result
}

func joinRelative(matches []candidate) string {
	return strings.Join(paths(matches, false), ", ")
}
