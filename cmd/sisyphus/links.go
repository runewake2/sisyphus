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

func repoCandidates(root string) []candidate {
	var result []candidate
	for _, file := range repoFiles(root) {
		rel := relative(root, file)
		keys := []string{rel}
		if hasSuffixFold(rel, ".md") {
			keys = append(keys, trimMarkdownExtension(rel))
		}
		result = append(result, candidate{file: file, relative: rel, keys: keys})
	}
	slices.SortFunc(result, func(a, b candidate) int { return strings.Compare(a.relative, b.relative) })
	return result
}

// matchTarget follows Obsidian: match by file name, with or without ".md", in any directory.
// A target with "/" matches the end of a path. An exact path from the repo root wins.
func matchTarget(candidates []candidate, target string, all bool) []candidate {
	target = strings.TrimLeft(target, "/")
	if target == "" {
		return nil
	}
	suffix := "/" + target
	var exact, partial []candidate
	for _, c := range candidates {
		if slices.ContainsFunc(c.keys, func(key string) bool { return strings.EqualFold(key, target) }) {
			exact = append(exact, c)
		}
		if slices.ContainsFunc(c.keys, func(key string) bool { return hasSuffixFold(key, suffix) }) {
			partial = append(partial, c)
		}
	}
	if all {
		seen := map[string]bool{}
		var result []candidate
		for _, c := range append(exact, partial...) {
			if !seen[c.relative] {
				seen[c.relative] = true
				result = append(result, c)
			}
		}
		return result
	}
	if len(exact) > 0 {
		return exact
	}
	return partial
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
	candidates := repoCandidates(root)
	file, err := findDocument(candidates, document)
	if err != nil {
		return "", nil, err
	}
	fileRelative := relative(root, file)
	content, err := os.ReadFile(file)
	if err != nil {
		return "", nil, err
	}
	lines := splitLines(string(content))

	headingsOf := map[string][]string{file: headingsIn(lines)}
	hasHeading := func(file, heading string) bool {
		if _, ok := headingsOf[file]; !ok {
			headingsOf[file] = headings(file)
		}
		return slices.ContainsFunc(headingsOf[file], func(h string) bool { return strings.EqualFold(h, heading) })
	}

	rows := []linkRow{}
	for _, found := range wikilinks(lines) {
		link := parseLink(found.link)
		var matches []candidate
		if link.target == "" && link.sameFile {
			for _, c := range candidates {
				if c.relative == fileRelative {
					matches = append(matches, c)
				}
			}
		} else {
			matches = matchTarget(candidates, link.target, false)
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
			!hasHeading(matches[0].file, link.heading) {
			status = "missing-heading"
		}
		rows = append(rows, linkRow{
			Line:     found.line,
			Link:     found.link,
			Status:   status,
			Resolved: strings.Join(paths(matches, absolute), ", "),
		})
	}
	return fileRelative, rows, nil
}

func findDocument(candidates []candidate, document string) (string, error) {
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
