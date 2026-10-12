// Package wikilink changes the target of wikilinks in Markdown text, for example [[unreleased-x]]
// to [[0.0.107]], when hephaestus renames a file. It finds links the way Obsidian and sisyphus do:
// not in code blocks and not in inline code. It knows nothing else about links; sisyphus lint checks
// that every link resolves.
package wikilink

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

var (
	linkPattern       = regexp.MustCompile(`\[\[([^\[\]\r\n]+)\]\]`)
	inlineCodePattern = regexp.MustCompile("`[^`]*`")
)

// Rename returns content with each wikilink whose target names a file in names changed to the new
// name, and the number of links that it changed. names maps an old file name, without directories
// and ".md", to the new one, for example "unreleased-fix-login" to "0.0.107". A target matches when
// its last path element, without ".md", is an old name, without regard to case. The new link has
// only the new name; the heading after "#" and the text after "|" stay.
func Rename(content string, names map[string]string) (string, int) {
	lower := map[string]string{}
	for old, renamed := range names {
		lower[strings.ToLower(old)] = renamed
	}
	lines := strings.SplitAfter(content, "\n")
	changed := 0
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		code := inlineCodePattern.FindAllStringIndex(line, -1)
		var rebuilt strings.Builder
		last := 0
		for _, match := range linkPattern.FindAllStringSubmatchIndex(line, -1) {
			if slices.ContainsFunc(code, func(span []int) bool { return match[0] >= span[0] && match[0] < span[1] }) {
				continue
			}
			inner := line[match[2]:match[3]]
			end := strings.IndexAny(inner, "#|")
			if end < 0 {
				end = len(inner)
			}
			target := strings.TrimSpace(inner[:end])
			name := strings.ToLower(strings.TrimSuffix(path.Base(strings.ReplaceAll(target, `\`, "/")), ".md"))
			renamed, ok := lower[name]
			if target == "" || !ok {
				continue
			}
			rebuilt.WriteString(line[last:match[0]])
			rebuilt.WriteString("[[" + renamed + inner[end:] + "]]")
			last = match[1]
			changed++
		}
		if last > 0 {
			rebuilt.WriteString(line[last:])
			lines[i] = rebuilt.String()
		}
	}
	return strings.Join(lines, ""), changed
}
