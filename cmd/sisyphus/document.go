package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
)

// document is a Markdown file with YAML frontmatter. Frontmatter is edited line by line, so that inline comments stay.
// bom is true for a file that starts with a UTF-8 byte order mark, and crlf for a file whose first
// line ends with CRLF; render writes both back, so that a change touches only the lines it must.
// headerOnly is true for a document that loadHeader read only up to the end of its frontmatter: its
// body is empty, and save refuses it, so that a partial read can never become a partial file.
type document struct {
	front      []string
	body       []string
	bom        bool
	crlf       bool
	headerOnly bool
}

// utf8BOM is the UTF-8 byte order mark that some editors write at the start of a file.
const utf8BOM = "\ufeff"

// field is one frontmatter line split into its parts: "key: value  # comment".
type field struct {
	key, value, comment string
}

// parseField splits a frontmatter line into its key, its value, and its inline comment, without a
// regular expression, because commands parse every line of thousands of issues. ok is false when the
// line does not start with a key of letters, "_", and "-", and a colon. A double-quoted value (with
// backslash escapes) or a single-quoted value counts only when nothing but spaces and a comment
// follows it; otherwise the value is the text up to the first "#", without trailing spaces. It gives
// the same result as the pattern ^([A-Za-z_-]+):\s*("(?:[^"\\]|\\.)*"|'[^']*'|[^#]*?)\s*(#.*)?$.
func parseField(line string) (field, bool) {
	colon := strings.IndexByte(line, ':')
	if colon <= 0 {
		return field{}, false
	}
	key := line[:colon]
	for i := 0; i < len(key); i++ {
		if c := key[i]; !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '-') {
			return field{}, false
		}
	}
	rest := strings.TrimLeft(line[colon+1:], asciiSpace)
	if end := quotedEnd(rest); end > 0 {
		after := strings.TrimLeft(rest[end:], asciiSpace)
		if after == "" || after[0] == '#' {
			return field{key: key, value: rest[:end], comment: after}, true
		}
	}
	if hash := strings.IndexByte(rest, '#'); hash >= 0 {
		return field{key: key, value: strings.TrimRight(rest[:hash], asciiSpace), comment: rest[hash:]}, true
	}
	return field{key: key, value: strings.TrimRight(rest, asciiSpace)}, true
}

// asciiSpace is the whitespace that frontmatter parsing skips, the same as \s in Go's regexp.
const asciiSpace = " \t\n\f\r"

// quotedEnd returns the length of the quoted string at the start of text, with its quotes, or 0 when
// text does not start with a complete double-quoted or single-quoted string.
func quotedEnd(text string) int {
	if text == "" {
		return 0
	}
	switch text[0] {
	case '"':
		for i := 1; i < len(text); i++ {
			switch text[i] {
			case '\\':
				if i+1 >= len(text) || text[i+1] == '\n' {
					return 0
				}
				i++
			case '"':
				return i + 1
			}
		}
	case '\'':
		if i := strings.IndexByte(text[1:], '\''); i >= 0 {
			return i + 2
		}
	}
	return 0
}

func parseDocument(content string) (*document, error) {
	bom := strings.HasPrefix(content, utf8BOM)
	firstEnd := strings.IndexByte(content, '\n')
	crlf := firstEnd > 0 && content[firstEnd-1] == '\r'
	lines := splitLines(strings.TrimPrefix(content, utf8BOM))
	if len(lines) == 0 || lines[0] != "---" {
		return nil, errors.New("The file does not start with YAML frontmatter ('---').")
	}
	end := slices.Index(lines[1:], "---") + 1
	if end == 0 {
		return nil, errors.New("The YAML frontmatter has no closing '---'.")
	}
	// The capacity limit makes an append to front copy, so it cannot overwrite body.
	return &document{front: lines[1:end:end], body: lines[end+1:], bom: bom, crlf: crlf}, nil
}

func loadDocument(path string) (*document, error) {
	trace(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseDocument(string(content))
}

// splitLines splits content into lines, without a leading byte order mark.
func splitLines(content string) []string {
	lines := strings.Split(strings.ReplaceAll(strings.TrimPrefix(content, utf8BOM), "\r\n", "\n"), "\n")
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

// lines returns every line of the file of a full document, as splitLines gives them: the frontmatter
// between its "---" lines, then the body.
func (d *document) lines() []string {
	return slices.Concat([]string{"---"}, d.front, []string{"---"}, d.body)
}

func (d *document) render() string {
	all := append([]string{"---"}, d.front...)
	all = append(all, "---")
	all = append(all, d.body...)
	newline := "\n"
	if d.crlf {
		newline = "\r\n"
	}
	text := strings.Join(all, newline) + newline
	if d.bom {
		text = utf8BOM + text
	}
	return text
}

func (d *document) save(path string) error {
	if d.headerOnly {
		return fmt.Errorf("internal error: %s was read without its body and cannot be written", path)
	}
	return writeFileAtomically(path, []byte(d.render()))
}

// loadHeader reads a file only up to the line that closes its frontmatter, for a command that does not
// need the body. A file without frontmatter gives the same error as loadDocument.
func loadHeader(path string) (*document, error) {
	trace(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var header strings.Builder
	for delimiters := 0; delimiters < 2; {
		line, err := reader.ReadString('\n')
		header.WriteString(line)
		trimmed := strings.TrimRight(line, "\r\n")
		if header.Len() == len(line) {
			trimmed = strings.TrimPrefix(trimmed, utf8BOM)
		}
		if trimmed == "---" {
			delimiters++
		} else if delimiters == 0 {
			break // not frontmatter: parseDocument gives the error
		}
		if err != nil {
			break
		}
	}
	doc, err := parseDocument(header.String())
	if err != nil {
		return nil, err
	}
	doc.headerOnly = true
	return doc, nil
}

func (d *document) get(key string) string {
	for _, line := range d.front {
		if !strings.HasPrefix(line, key+":") {
			continue
		}
		f, ok := parseField(line)
		if !ok {
			return ""
		}
		return unquote(f.value)
	}
	return ""
}

// unquote reverses quote: it unwraps a double- or single-quoted frontmatter value, or returns value
// unchanged if it is not quoted. An invalid double-quoted value becomes "", matching document.get.
func unquote(value string) string {
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
		if f, ok := parseField(line); ok {
			comment = f.comment
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
	// A field that the frontmatter does not have is added only to set a value, so that a template
	// without an optional field (for example bookmark) keeps working, and clearing it adds nothing.
	if value != "" {
		d.front = append(d.front, replacement)
	}
}
