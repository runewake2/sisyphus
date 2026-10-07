package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	states      = []string{"open", "in-progress", "closed"}
	resolutions = []string{"completed", "abandoned"}
	priorities  = []string{"critical", "high", "medium", "low"}
	efforts     = []string{"small", "medium", "large", "x-large"}
	namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+){1,5}$`)
)

const (
	stateDirectoryList = "issues/open, issues/in-progress, or issues/closed"
	noBookmarkWarning  = "The issue is in-progress but has no bookmark. Set it with --bookmark."
)

// newOptions holds the values of "sisyphus new". An empty resolution and a nil pointer mean that the option was not given.
type newOptions struct {
	name, state, resolution, priority, effort, tags string
	title, bookmark, deferredFrom, parent           *string
}

type issueFile struct {
	state string
	path  string
}

func newIssue(root string, o newOptions, warnings io.Writer) (string, error) {
	if !namePattern.MatchString(o.name) {
		return "", fmt.Errorf("Invalid issue name '%s'. Use 2-6 lowercase words in kebab-case, for example explicit-step-dependencies.", o.name)
	}
	if err := checkResolution(o.state, o.resolution,
		"A closed issue needs --resolution completed or --resolution abandoned.",
		"Set --resolution only when the state is closed."); err != nil {
		return "", err
	}

	var clashes []string
	for _, file := range repoFiles(root) {
		base := filepath.Base(file)
		if hasSuffixFold(base, ".md") && strings.EqualFold(trimMarkdownExtension(base), o.name) {
			clashes = append(clashes, relative(root, file))
		}
	}
	slices.Sort(clashes)
	if len(clashes) > 0 {
		return "", fmt.Errorf("The name '%s' is not unique. These files have the same name: %s.", o.name, strings.Join(clashes, ", "))
	}

	parentValue := ""
	if o.parent != nil {
		value, err := parentValueFor(root, *o.parent, o.name, warnings)
		if err != nil {
			return "", err
		}
		parentValue = value
	}

	doc, err := loadDocument(filepath.Join(root, "issues", "TEMPLATE.md"))
	if err != nil {
		return "", err
	}
	doc.front = slices.DeleteFunc(doc.front, func(line string) bool {
		return strings.HasPrefix(strings.TrimSpace(line), "#")
	})

	title := capitalize(strings.ReplaceAll(o.name, "-", " "))
	if o.title != nil {
		title = *o.title
	}
	var tags []string
	for _, tag := range strings.Split(o.tags, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	deferred := ""
	if o.deferredFrom != nil {
		deferred = deferredValue(*o.deferredFrom)
	}

	doc.set("title", quote(title))
	applyState(doc, o.state, o.resolution)
	doc.set("priority", o.priority)
	doc.set("effort", o.effort)
	doc.set("tags", "["+strings.Join(tags, ", ")+"]")
	doc.set("created", today())
	doc.set("bookmark", valueOrEmpty(o.bookmark))
	doc.set("deferred-from", deferred)
	doc.set("parent", parentValue)
	for i, line := range doc.body {
		if line == "# <Title>" {
			doc.body[i] = "# " + title
		}
	}

	destination := issuePath(root, o.state, o.name)
	if err := saveIssue(doc, destination, destination); err != nil {
		return "", err
	}
	if o.state == "in-progress" && o.bookmark == nil {
		warn(warnings, noBookmarkWarning)
	}
	return relative(root, destination), nil
}

func updateIssue(root, reference, state, resolution string, bookmark *string, warnings io.Writer) (string, error) {
	name, current, doc, err := loadIssue(root, reference)
	if err != nil {
		return "", err
	}
	if current.state == "closed" && state != "closed" {
		return "", fmt.Errorf("Issue '%s' is closed. Do not reopen a closed issue. Create a new issue and link to [[%s]].", name, name)
	}
	if err := checkResolution(state, resolution,
		"To close an issue, give --resolution completed or --resolution abandoned.",
		"Set --resolution only when the new state is closed."); err != nil {
		return "", err
	}
	if recorded := doc.get("state"); recorded != current.state {
		warn(warnings, fmt.Sprintf("The issue was in issues/%s/ but its state field was '%s'. The update corrects both.", current.state, recorded))
	}

	applyState(doc, state, resolution)
	if state == "open" {
		doc.set("bookmark", "")
	} else if bookmark != nil {
		doc.set("bookmark", *bookmark)
	}
	if state == "in-progress" && doc.get("bookmark") == "" {
		warn(warnings, noBookmarkWarning)
	}
	if state == "closed" {
		var open []string
		for _, sub := range subIssues(root, name) {
			if sub.state != "closed" {
				open = append(open, sub.name)
			}
		}
		if len(open) > 0 {
			warn(warnings, fmt.Sprintf("Issue '%s' has sub-issues that are not closed: %s.", name, strings.Join(open, ", ")))
		}
	}

	destination := issuePath(root, state, name)
	if err := saveIssue(doc, current.path, destination); err != nil {
		return "", err
	}
	return relative(root, destination), nil
}

func setParent(root, reference string, parent *string, warnings io.Writer) (string, error) {
	name, issue, doc, err := loadIssue(root, reference)
	if err != nil {
		return "", err
	}
	value := ""
	if parent != nil {
		if value, err = parentValueFor(root, *parent, name, warnings); err != nil {
			return "", err
		}
	}
	doc.set("parent", value)
	if err := doc.save(issue.path); err != nil {
		return "", err
	}
	return relative(root, issue.path), nil
}

// applyState sets the fields that follow from the state: only a closed issue has a resolution and a closed date.
func applyState(doc *document, state, resolution string) {
	closed := ""
	if state == "closed" {
		closed = today()
	}
	doc.set("state", state)
	doc.set("resolution", resolution)
	doc.set("closed", closed)
}

func parentValueFor(root, reference, child string, warnings io.Writer) (string, error) {
	parent := issueName(reference)
	if parent == "" {
		return "", fmt.Errorf("'%s' does not name an issue.", reference)
	}
	if parent == child {
		return "", fmt.Errorf("Issue '%s' cannot be its own parent.", child)
	}
	matches := issueMatches(root, parent)
	if len(matches) == 0 {
		return "", fmt.Errorf("The parent issue '%s' does not exist in %s.", parent, stateDirectoryList)
	}
	if matches[0].state == "closed" {
		warn(warnings, fmt.Sprintf("The parent issue '%s' is closed.", parent))
	}

	seen := map[string]bool{parent: true}
	for ancestor := parentName(root, parent); ancestor != "" && !seen[ancestor]; ancestor = parentName(root, ancestor) {
		if ancestor == child {
			return "", fmt.Errorf("Issue '%s' is already below '%s'. A parent of '%s' cannot be one of its sub-issues.", parent, child, child)
		}
		seen[ancestor] = true
	}
	return quote("[[" + parent + "]]"), nil
}

func parentName(root, name string) string {
	matches := issueMatches(root, name)
	if len(matches) != 1 {
		return ""
	}
	doc, err := loadDocument(matches[0].path)
	if err != nil {
		return ""
	}
	return issueName(doc.get("parent"))
}

type subIssue struct {
	name  string
	state string
}

func subIssues(root, name string) []subIssue {
	var result []subIssue
	for _, state := range states {
		files, _ := filepath.Glob(filepath.Join(root, "issues", state, "*.md"))
		for _, file := range files {
			doc, err := loadDocument(file)
			if err != nil || issueName(doc.get("parent")) != name {
				continue
			}
			result = append(result, subIssue{name: trimMarkdownExtension(filepath.Base(file)), state: state})
		}
	}
	slices.SortFunc(result, func(a, b subIssue) int { return strings.Compare(a.name, b.name) })
	return result
}

func issuePath(root, state, name string) string {
	return filepath.Join(root, "issues", state, name+".md")
}

func issueMatches(root, name string) []issueFile {
	var matches []issueFile
	for _, state := range states {
		if path := issuePath(root, state, name); fileExists(path) {
			matches = append(matches, issueFile{state: state, path: path})
		}
	}
	return matches
}

func findIssue(root, name string) (issueFile, error) {
	matches := issueMatches(root, name)
	switch len(matches) {
	case 0:
		return issueFile{}, fmt.Errorf("No issue named '%s' in %s.", name, stateDirectoryList)
	case 1:
		return matches[0], nil
	default:
		var paths []string
		for _, m := range matches {
			paths = append(paths, relative(root, m.path))
		}
		return issueFile{}, fmt.Errorf("Issue '%s' exists in more than one state directory: %s. Remove the duplicate.", name, strings.Join(paths, ", "))
	}
}

func loadIssue(root, reference string) (string, issueFile, *document, error) {
	name := issueName(reference)
	issue, err := findIssue(root, name)
	if err != nil {
		return "", issueFile{}, nil, err
	}
	doc, err := loadDocument(issue.path)
	return name, issue, doc, err
}

// saveIssue writes the document to "from" and moves it to "to", the directory of its state.
func saveIssue(doc *document, from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	if err := doc.save(from); err != nil {
		return err
	}
	if from != to {
		return os.Rename(from, to)
	}
	return nil
}

func checkResolution(state, resolution, missingMessage, unexpectedMessage string) error {
	if state == "closed" && resolution == "" {
		return errors.New(missingMessage)
	}
	if state != "closed" && resolution != "" {
		return errors.New(unexpectedMessage)
	}
	return nil
}

func deferredValue(reference string) string {
	r := strings.TrimSpace(reference)
	if strings.HasPrefix(r, "[[") || strings.Contains(r, "/") {
		return quote(r)
	}
	return quote("[[" + issueName(r) + "]]")
}

func capitalize(text string) string {
	if text == "" {
		return text
	}
	first, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(first)) + text[size:]
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
