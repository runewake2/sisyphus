package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// document is a Markdown file with YAML frontmatter. Frontmatter is edited line by line, so that inline comments stay.
type document struct {
	front []string
	body  []string
}

var fieldPattern = regexp.MustCompile(`^([A-Za-z_-]+):\s*("(?:[^"\\]|\\.)*"|'[^']*'|[^#]*?)\s*(#.*)?$`)

func parseDocument(content string) (*document, error) {
	lines := splitLines(content)
	if len(lines) == 0 || lines[0] != "---" {
		return nil, errors.New("The file does not start with YAML frontmatter ('---').")
	}
	end := slices.Index(lines[1:], "---") + 1
	if end == 0 {
		return nil, errors.New("The YAML frontmatter has no closing '---'.")
	}
	// The capacity limit makes an append to front copy, so it cannot overwrite body.
	return &document{front: lines[1:end:end], body: lines[end+1:]}, nil
}

func loadDocument(path string) (*document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseDocument(string(content))
}

func splitLines(content string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// quote returns a JSON string, which is also a valid YAML double-quoted string.
func quote(value string) string {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
	return strings.TrimSuffix(buffer.String(), "\n")
}

func (d *document) render() string {
	all := append([]string{"---"}, d.front...)
	all = append(all, "---")
	all = append(all, d.body...)
	return strings.Join(all, "\n") + "\n"
}

func (d *document) save(path string) error {
	return os.WriteFile(path, []byte(d.render()), 0o644)
}

func (d *document) get(key string) string {
	for _, line := range d.front {
		if !strings.HasPrefix(line, key+":") {
			continue
		}
		match := fieldPattern.FindStringSubmatch(line)
		if match == nil {
			return ""
		}
		value := strings.TrimSpace(match[2])
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			var unquoted string
			if json.Unmarshal([]byte(value), &unquoted) != nil {
				return ""
			}
			return unquoted
		}
		if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			return value[1 : len(value)-1]
		}
		return value
	}
	return ""
}

func (d *document) set(key, value string) {
	prefix := key + ":"
	replacement := prefix
	if value != "" {
		replacement = prefix + " " + value
	}
	for i, line := range d.front {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		comment := ""
		if match := fieldPattern.FindStringSubmatch(line); match != nil {
			comment = match[3]
		}
		if comment == "" {
			d.front[i] = replacement
			return
		}
		column := utf8.RuneCountInString(line) - utf8.RuneCountInString(comment)
		padding := max(column-utf8.RuneCountInString(replacement), 1)
		d.front[i] = replacement + strings.Repeat(" ", padding) + comment
		return
	}
	d.front = append(d.front, replacement)
}
