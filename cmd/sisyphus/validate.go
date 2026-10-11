package main

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// keyPrefix reports whether line starts like a frontmatter field: a key of letters, "_" or "-",
// then a colon.
func keyPrefix(line string) bool {
	key, _, ok := strings.Cut(line, ":")
	return ok && key != "" && strings.Trim(key, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-") == ""
}

// issueProblems returns what is wrong with an issue file that sisyphus could read: frontmatter
// lines that sisyphus cannot read as one "key: value" field, values that open a quote, a list, or a
// map and do not close it, a state field that does not match the state directory, and enum fields
// with a value that is not valid. It does not check relationships or wikilinks; sisyphus lint does.
func issueProblems(issue issueFile, doc *document) []string {
	var problems []string
	for _, line := range doc.front {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !keyPrefix(line) {
			problems = append(problems, fmt.Sprintf("the frontmatter line %q is not a \"key: value\" field", line))
			continue
		}
		match := fieldPattern.FindStringSubmatch(line)
		if match == nil {
			problems = append(problems, fmt.Sprintf("the frontmatter line %q has a value that sisyphus cannot read", line))
			continue
		}
		if problem := valueProblem(strings.TrimSpace(match[2])); problem != "" {
			problems = append(problems, fmt.Sprintf("the value of '%s' %s", match[1], problem))
		}
	}
	if recorded := doc.get("state"); recorded != issue.state {
		problems = append(problems, fmt.Sprintf("its state field is '%s', but it is in issues/%s/", recorded, issue.state))
	}
	for _, field := range []struct {
		key     string
		choices []string
		empty   bool
	}{
		{"priority", priorities, false},
		{"effort", efforts, false},
		{"resolution", resolutions, true},
	} {
		value := doc.get(field.key)
		if value == "" && field.empty {
			continue
		}
		if !slices.Contains(field.choices, value) {
			problems = append(problems, fmt.Sprintf("its %s '%s' is not one of: %s", field.key, value, strings.Join(field.choices, ", ")))
		}
	}
	return problems
}

// valueProblem checks the shape of one frontmatter value: a double-quoted value must be a valid
// string, and a value that opens a quote, a flow list, or a flow map must close it.
func valueProblem(value string) string {
	switch {
	case value == "":
		return ""
	case strings.HasPrefix(value, `"`):
		var unquoted string
		if len(value) < 2 || !strings.HasSuffix(value, `"`) || json.Unmarshal([]byte(value), &unquoted) != nil {
			return "is not a valid double-quoted string"
		}
	case strings.HasPrefix(value, "'"):
		if len(value) < 2 || !strings.HasSuffix(value, "'") {
			return "opens a single quote and does not close it"
		}
	case strings.HasPrefix(value, "["):
		if !strings.HasSuffix(value, "]") {
			return "opens a list with '[' and does not close it with ']'"
		}
	case strings.HasPrefix(value, "{"):
		if !strings.HasSuffix(value, "}") {
			return "opens a map with '{' and does not close it with '}'"
		}
	}
	return ""
}

// warnIssueProblems writes one warning for an issue file that has problems, or that sisyphus cannot
// read at all. warnings can be nil, which means no warnings.
func warnIssueProblems(x *issueIndex, issue issueFile, doc *document, err error, warnings io.Writer) {
	if warnings == nil {
		return
	}
	path := relative(x.root, issue.path)
	if err != nil {
		warn(warnings, fmt.Sprintf("%s cannot be read: %s", path, err))
		return
	}
	if problems := issueProblems(issue, doc); len(problems) > 0 {
		warn(warnings, fmt.Sprintf("%s: %s. Fix the file by hand.", path, strings.Join(problems, "; ")))
	}
}
