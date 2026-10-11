package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	states      = []string{"open", "in-progress", "closed"}
	resolutions = []string{"completed", "abandoned"}
	priorities  = []string{"critical", "high", "medium", "low"}
	efforts     = []string{"small", "medium", "large", "x-large"}
	// builtinTags are suggested in the template and in --help. Tags stay free-form: these are not enforced.
	builtinTags = []string{"bug", "feature"}
	namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+){1,5}$`)
	// directoryPattern is the shape of each directory in a full issue name such as web/auth/fix-login-bug.
	directoryPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

const (
	openWorkFieldsError = "An open issue has no owner and no bookmark: nobody works on it yet. Leave out --owner and --bookmark, or give the state in-progress."
	stateDirectoryList  = "issues/open, issues/in-progress, or issues/closed"
	noBookmarkWarning   = "The issue is in-progress but has no bookmark. Set it with --bookmark."
)

// newOptions holds the values of "sisyphus new". An empty resolution and a nil pointer mean that the option was not given.
type newOptions struct {
	name, state, resolution, priority, effort, tags string
	title, bookmark, deferredFrom, parent           *string
	owner, approver, workspace                      *string
	remote, context                                 *string
	dependsOn                                       []string
	metadata                                        map[string]string
}

// issueFile is one issue on disk. name is the full issue name: the path below issues/<state>/,
// without ".md", for example web/auth/fix-login-bug.
type issueFile struct {
	state string
	path  string
	name  string
}

// validName reports whether name is a valid full issue name: 2-6 kebab-case words, optionally
// below kebab-case directories.
func validName(name string) bool {
	parts := strings.Split(name, "/")
	for _, directory := range parts[:len(parts)-1] {
		if !directoryPattern.MatchString(directory) {
			return false
		}
	}
	return namePattern.MatchString(parts[len(parts)-1])
}

// warnOpenDependencies warns when an issue is in progress while some of its depends-on issues are
// not closed. depends-on means "must close before this one can start". It is a warning, not an
// error, because a person may start blocked work on purpose.
func warnOpenDependencies(x *issueIndex, name string, doc *document, warnings io.Writer) {
	if open := openDependencies(x, doc); len(open) > 0 {
		warn(warnings, fmt.Sprintf("Issue '%s' is in progress, but it depends on issues that are not closed: %s.", name, strings.Join(open, ", ")))
	}
}

// checkOpenWorkFields fails when an open issue would get an owner or a bookmark. Those fields
// describe active work, so sisyphus clears them for an open issue; a value that the user gave must
// not be dropped without a message.
func checkOpenWorkFields(state string, owner, bookmark *string) error {
	if state == "open" && (valueOrEmpty(owner) != "" || valueOrEmpty(bookmark) != "") {
		return errors.New(openWorkFieldsError)
	}
	return nil
}

// setTitleHeading sets the first "# " heading of the body, outside code blocks, to title. A body
// with no such heading does not change.
func setTitleHeading(doc *document, title string) {
	for _, index := range outsideCodeFences(doc.body) {
		if headingPattern.MatchString(doc.body[index]) && headingLevel(doc.body[index]) == 1 {
			doc.body[index] = "# " + title
			return
		}
	}
}

// validateTitle rejects a title that the frontmatter, the "# Title" heading, and the list table
// cannot show on one line: an empty title, or one with a line break, a tab, or another control
// character.
func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("The title is empty. Give a title, or leave out --title to make one from the name.")
	}
	if i := strings.IndexFunc(title, unicode.IsControl); i >= 0 {
		return fmt.Errorf("The title has a line break, a tab, or another control character (%q). Give the title on one line.", title[i:i+1])
	}
	return nil
}

// baseName is the last part of a full issue name, which is also its file name without ".md".
func baseName(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

func newIssue(root string, o newOptions, warnings io.Writer) (string, error) {
	x := loadIndex(root)
	if !validName(o.name) {
		return "", fmt.Errorf("Invalid issue name '%s'. Use 2-6 lowercase words in kebab-case, for example fix-login-bug, optionally below kebab-case directories, for example web/auth/fix-login-bug.", o.name)
	}
	resolution := resolveResolution(o.state, o.resolution)
	if err := checkResolution(o.state, resolution, "Set --resolution only when the state is closed."); err != nil {
		return "", err
	}
	if err := validateTags(o.tags); err != nil {
		return "", err
	}
	if err := checkOpenWorkFields(o.state, o.owner, o.bookmark); err != nil {
		return "", err
	}

	if existing := x.exact[o.name]; len(existing) > 0 {
		return "", fmt.Errorf("Issue '%s' already exists: %s.", o.name, relative(root, existing[0].path))
	}
	if clashes := nameClashes(root, baseName(o.name)); len(clashes) > 0 {
		advice := fmt.Sprintf("Link to the full name, [[%s]].", o.name)
		if o.name == baseName(o.name) {
			advice = "Put the issue in a subdirectory to give it a unique full name."
		}
		warn(warnings, fmt.Sprintf("Other files also have the name '%s': %s. A link to [[%s]] is ambiguous. %s",
			baseName(o.name), strings.Join(clashes, ", "), baseName(o.name), advice))
	}

	parentValue := ""
	if o.parent != nil {
		value, err := parentValueFor(x, *o.parent, o.name, warnings)
		if err != nil {
			return "", err
		}
		parentValue = value
	}
	var dependsOnValue []string
	for _, reference := range o.dependsOn {
		if strings.TrimSpace(reference) == "" {
			continue
		}
		value, err := dependsOnValueFor(x, strings.TrimSpace(reference), o.name, warnings)
		if err != nil {
			return "", err
		}
		if !slices.Contains(dependsOnValue, value) {
			dependsOnValue = append(dependsOnValue, value)
		}
	}

	doc, err := loadDocument(filepath.Join(root, "issues", "TEMPLATE.md"))
	if err != nil {
		return "", err
	}
	doc.front = slices.DeleteFunc(doc.front, func(line string) bool {
		return strings.HasPrefix(strings.TrimSpace(line), "#")
	})

	title := capitalize(strings.ReplaceAll(baseName(o.name), "-", " "))
	if o.title != nil {
		if err := validateTitle(*o.title); err != nil {
			return "", err
		}
		title = *o.title
	}
	deferred := ""
	if o.deferredFrom != nil {
		value, err := deferredValue(x, *o.deferredFrom, warnings)
		if err != nil {
			return "", err
		}
		deferred = value
	}
	remoteField := ""
	if o.remote != nil {
		value, err := remoteValue(*o.remote)
		if err != nil {
			return "", err
		}
		remoteField = value
	}

	doc.set("title", quote(title))
	applyState(doc, o.state, resolution)
	doc.set("priority", o.priority)
	doc.set("effort", o.effort)
	doc.set("tags", formatTags(o.tags))
	doc.set("created", today())
	doc.set("owner", valueOrEmpty(o.owner))
	doc.set("approver", valueOrEmpty(o.approver))
	doc.set("bookmark", valueOrEmpty(o.bookmark))
	if o.workspace != nil {
		addToList(doc, "workspaces", *o.workspace)
	}
	if len(o.metadata) > 0 {
		doc.set("metadata", formatMetadata(o.metadata))
	}
	doc.set("deferred-from", deferred)
	doc.set("parent", parentValue)
	doc.set("remote", remoteField)
	if len(dependsOnValue) > 0 {
		doc.set("depends-on", formatDependsOn(dependsOnValue))
	}
	for i, line := range doc.body {
		if line == "# <Title>" {
			doc.body[i] = "# " + title
		}
	}
	if o.context != nil {
		if replaced, ok := replaceSection(doc.body, "Context", strings.Split(*o.context, "\n")); ok {
			doc.body = replaced
		}
	}

	destination := issuePath(root, o.state, o.name)
	if err := saveIssue(root, doc, destination, destination); err != nil {
		return "", err
	}
	if o.state == "in-progress" && o.bookmark == nil {
		warn(warnings, noBookmarkWarning)
	}
	if o.state == "in-progress" {
		warnOpenDependencies(x, o.name, doc, warnings)
	}
	return relative(root, destination), nil
}

// updateOptions holds the values of "sisyphus update" that are not positional arguments. A nil pointer means the
// flag was not given. owner and bookmark describe the current, active work on the issue: they are cleared when
// the issue returns to open. workspaces is a history of every workspace that has worked on the issue, so an
// entry is appended, never cleared. approver is not tied to active work and is only ever set explicitly.
// metadata entries are merged into the existing map (added, or overwritten by key), never cleared automatically:
// it is a place for an agent to leave arbitrary notes, such as a session id, so work can be resumed with
// context later, including after the issue closes or reopens.
type updateOptions struct {
	resolution                           string
	bookmark, owner, approver, workspace *string
	priority, effort, tags, title        *string
	addTags, removeTags                  string
	metadata                             map[string]string
}

// updateIssue changes the issue that reference gives. An empty state keeps the current state.
func updateIssue(root, reference, state string, o updateOptions, warnings io.Writer) (string, error) {
	x := loadIndex(root)
	name, current, doc, err := x.load(reference)
	if err != nil {
		return "", err
	}
	if state == "" {
		state = current.state
	}
	if current.state == "closed" && state != "closed" {
		return "", fmt.Errorf("Issue '%s' is closed. Do not reopen a closed issue. Create a new issue and link to [[%s]].", name, name)
	}
	resolution := resolveResolution(state, o.resolution)
	if err := checkResolution(state, resolution, "Set --resolution only when the new state is closed."); err != nil {
		return "", err
	}
	if err := checkOpenWorkFields(state, o.owner, o.bookmark); err != nil {
		return "", err
	}
	if recorded := doc.get("state"); recorded != current.state {
		warn(warnings, fmt.Sprintf("The issue was in issues/%s/ but its state field was '%s'. The update corrects both.", current.state, recorded))
	}

	// When the state does not change and no --resolution is given, the resolution and the closed
	// date stay: applyState would reset an abandoned issue to completed, and its closed date to today.
	if keep := state == current.state && o.resolution == "" && doc.get("state") == state; !keep {
		applyState(doc, state, resolution)
	}
	if state == "open" {
		doc.set("bookmark", "")
		doc.set("owner", "")
	} else {
		if o.bookmark != nil {
			doc.set("bookmark", *o.bookmark)
		}
		if o.owner != nil {
			doc.set("owner", *o.owner)
		}
	}
	if len(o.metadata) > 0 {
		merged := parseMetadata(doc.get("metadata"))
		for k, v := range o.metadata {
			merged[k] = v
		}
		doc.set("metadata", formatMetadata(merged))
	}
	if o.workspace != nil {
		addToList(doc, "workspaces", *o.workspace)
	}
	if o.approver != nil {
		doc.set("approver", *o.approver)
	}
	if o.priority != nil {
		if err := oneOf("priority", *o.priority, priorities); err != nil {
			return "", err
		}
		doc.set("priority", *o.priority)
	}
	if o.effort != nil {
		if err := oneOf("effort", *o.effort, efforts); err != nil {
			return "", err
		}
		doc.set("effort", *o.effort)
	}
	if o.tags != nil {
		if err := validateTags(*o.tags); err != nil {
			return "", err
		}
		doc.set("tags", formatTags(*o.tags))
	}
	if o.addTags != "" || o.removeTags != "" {
		if err := validateTags(o.addTags); err != nil {
			return "", err
		}
		tags := parseList(doc.get("tags"))
		for _, tag := range parseList(formatTags(o.addTags)) {
			if !containsFold(tags, tag) {
				tags = append(tags, tag)
			}
		}
		removed := parseList(o.removeTags)
		tags = slices.DeleteFunc(tags, func(tag string) bool { return containsFold(removed, tag) })
		doc.set("tags", formatList(tags))
	}
	if o.title != nil {
		if err := validateTitle(*o.title); err != nil {
			return "", err
		}
		doc.set("title", quote(*o.title))
		setTitleHeading(doc, *o.title)
	}
	if state == "in-progress" && doc.get("bookmark") == "" {
		warn(warnings, noBookmarkWarning)
	}
	if state == "in-progress" {
		warnOpenDependencies(x, name, doc, warnings)
	}
	if state == "closed" && current.state != "closed" {
		var open []string
		for _, sub := range subIssues(x, name) {
			if sub.state != "closed" {
				open = append(open, sub.name)
			}
		}
		if len(open) > 0 {
			warn(warnings, fmt.Sprintf("Issue '%s' has sub-issues that are not closed: %s.", name, strings.Join(open, ", ")))
		}
		var blocked []string
		for _, dep := range dependents(x, name) {
			if dep.state != "closed" {
				blocked = append(blocked, dep.name)
			}
		}
		if len(blocked) > 0 {
			warn(warnings, fmt.Sprintf("Issue '%s' is closing, but these issues still depend on it: %s.", name, strings.Join(blocked, ", ")))
		}
		if section, ok := sectionIn(doc.body, "Resolution"); ok && isPlaceholder(section) {
			warn(warnings, fmt.Sprintf("The Resolution section of '%s' holds no text. Complete it with: sisyphus edit %s resolution \"<text>\"", name, name))
		}
	}

	destination := issuePath(root, state, name)
	if err := saveIssue(root, doc, current.path, destination); err != nil {
		return "", err
	}
	return relative(root, destination), nil
}

func setParent(root, reference string, parent *string, warnings io.Writer) (string, error) {
	x := loadIndex(root)
	name, issue, doc, err := x.load(reference)
	if err != nil {
		return "", err
	}
	value := ""
	if parent != nil {
		if value, err = parentValueFor(x, *parent, name, warnings); err != nil {
			return "", err
		}
	}
	doc.set("parent", value)
	if err := doc.save(issue.path); err != nil {
		return "", err
	}
	return relative(root, issue.path), nil
}

func setRemote(root, reference string, remote *string, warnings io.Writer) (string, error) {
	x := loadIndex(root)
	_, issue, doc, err := x.load(reference)
	if err != nil {
		return "", err
	}
	value := ""
	if remote != nil {
		if value, err = remoteValue(*remote); err != nil {
			return "", err
		}
	}
	doc.set("remote", value)
	if err := doc.save(issue.path); err != nil {
		return "", err
	}
	return relative(root, issue.path), nil
}

// editSection replaces the text of the section named heading in an issue's body, or with appendText adds
// text at the end of the section. An append to a section that holds only the template placeholder
// replaces the placeholder.
func editSection(root, reference, heading, text string, appendText bool) (string, error) {
	x := loadIndex(root)
	name, issue, doc, err := x.load(reference)
	if err != nil {
		return "", err
	}
	lines := splitLines(strings.TrimSpace(text))
	if len(lines) == 0 {
		return "", errors.New("The text is empty. Give the text for the section.")
	}
	start, _, ok := sectionBounds(doc.body, heading)
	if !ok || headingLevel(doc.body[start]) < 2 {
		return "", fmt.Errorf("Issue '%s' has no section '%s'. Its sections are: %s.", name, heading, strings.Join(sectionNames(doc.body), ", "))
	}
	// A heading at the level of the section or above would end the section inside the text, and make
	// a second section that a later edit could change instead of the real one.
	level := headingLevel(doc.body[start])
	for _, index := range outsideCodeFences(lines) {
		if headingPattern.MatchString(lines[index]) && headingLevel(lines[index]) <= level {
			return "", fmt.Errorf("The text has the heading %q, which would split the section '%s'. Use a heading with more than %d '#', or no heading.", lines[index], heading, level)
		}
	}
	if appendText {
		if current, _ := sectionIn(doc.body, heading); !isPlaceholder(current) {
			lines = slices.Concat(trimBlankLines(current), []string{""}, lines)
		}
	}
	doc.body, _ = replaceSection(doc.body, heading, lines)
	if err := doc.save(issue.path); err != nil {
		return "", err
	}
	return relative(root, issue.path), nil
}

// sectionNames returns the headings of the sections that editSection can change: every heading below the title.
func sectionNames(lines []string) []string {
	var names []string
	for _, index := range outsideCodeFences(lines) {
		if match := headingPattern.FindStringSubmatch(lines[index]); match != nil && headingLevel(lines[index]) >= 2 {
			names = append(names, match[1])
		}
	}
	return names
}

// isPlaceholder reports whether a section holds only the template's "<...>" text, or nothing.
func isPlaceholder(lines []string) bool {
	text := strings.TrimSpace(strings.Join(lines, "\n"))
	return text == "" || strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">") && strings.Count(text, ">\n") == 0
}

func trimBlankLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// remoteValue loosely validates and quotes a remote reference: it must be a URL with a scheme and a host.
func remoteValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("'%s' is not a URL.", raw)
	}
	return quote(raw), nil
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

func parentValueFor(x *issueIndex, reference, child string, warnings io.Writer) (string, error) {
	if issueName(reference) == "" {
		return "", fmt.Errorf("'%s' does not name an issue.", reference)
	}
	match, found, err := x.match(issueName(reference))
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("The parent issue '%s' does not exist in %s.", issueName(reference), stateDirectoryList)
	}
	parent := match.name
	if parent == child {
		return "", fmt.Errorf("Issue '%s' cannot be its own parent.", child)
	}
	if match.state == "closed" {
		warn(warnings, fmt.Sprintf("The parent issue '%s' is closed.", parent))
	}

	seen := map[string]bool{parent: true}
	for ancestor := parentName(x, parent); ancestor != "" && !seen[ancestor]; ancestor = parentName(x, ancestor) {
		if ancestor == child {
			return "", fmt.Errorf("Issue '%s' is already below '%s'. A parent of '%s' cannot be one of its sub-issues.", parent, child, child)
		}
		seen[ancestor] = true
	}
	return quote("[[" + parent + "]]"), nil
}

// parentName returns the full name of the parent of the issue name, or "" when it has none.
func parentName(x *issueIndex, name string) string {
	match, found, err := x.match(name)
	if err != nil || !found {
		return ""
	}
	doc, err := x.document(match)
	if err != nil {
		return ""
	}
	return x.canonical(doc.get("parent"))
}

type subIssue struct {
	name  string
	state string
}

func subIssues(x *issueIndex, name string) []subIssue {
	var result []subIssue
	for _, issue := range x.issues {
		doc, err := x.document(issue)
		if err != nil || x.canonical(doc.get("parent")) != name {
			continue
		}
		result = append(result, subIssue{name: issue.name, state: issue.state})
	}
	return result
}

// parseDependsOn reads a depends-on frontmatter value, for example `["[[a]]", "[[b]]"]`, into issue
// names, for example ["a", "b"].
func parseDependsOn(value string) []string {
	var names []string
	for _, item := range parseList(value) {
		if name := issueName(unquote(item)); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// formatDependsOn writes issue names back as a depends-on frontmatter value of quoted wikilinks.
func formatDependsOn(names []string) string {
	items := make([]string, len(names))
	for i, name := range names {
		items[i] = quote("[[" + name + "]]")
	}
	return formatList(items)
}

// parseMetadata reads a frontmatter value like `{session-id: "abc123", note: "a, b"}` into a map. It
// splits the map only on commas outside quotes, so that a quoted value can hold commas, colons, and
// escaped quotes: what formatMetadata writes, parseMetadata reads back the same.
func parseMetadata(value string) map[string]string {
	value = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "{"), "}")
	m := map[string]string{}
	for _, item := range splitOutsideQuotes(value, ',') {
		key, val, ok := strings.Cut(item, ":")
		if !ok {
			continue
		}
		if key = strings.TrimSpace(key); key != "" {
			m[key] = unquote(strings.TrimSpace(val))
		}
	}
	return m
}

// splitOutsideQuotes splits text on each separator that is not inside a double-quoted string (with
// backslash escapes) or a single-quoted string.
func splitOutsideQuotes(text string, separator rune) []string {
	var parts []string
	var quote rune
	escaped := false
	start := 0
	for i, r := range text {
		switch {
		case escaped:
			escaped = false
		case quote == '"' && r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == separator:
			parts = append(parts, text[start:i])
			start = i + utf8.RuneLen(r)
		}
	}
	return append(parts, text[start:])
}

// formatMetadata writes a map back as a frontmatter value, with keys sorted for a deterministic diff.
func formatMetadata(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	items := make([]string, len(keys))
	for i, k := range keys {
		items[i] = k + ": " + quote(m[k])
	}
	return "{" + strings.Join(items, ", ") + "}"
}

func setDependsOn(root, reference string, blocking *string, clear bool, warnings io.Writer) (string, error) {
	x := loadIndex(root)
	name, issue, doc, err := x.load(reference)
	if err != nil {
		return "", err
	}
	current := parseDependsOn(doc.get("depends-on"))
	switch {
	case blocking == nil: // clear is guaranteed true by the caller: clear every dependency
		current = nil
	case clear:
		target := x.canonical(*blocking)
		kept := slices.DeleteFunc(slices.Clone(current), func(n string) bool { return x.canonical(n) == target })
		if len(kept) == len(current) {
			has := "no dependencies"
			if len(current) > 0 {
				has = "these dependencies: " + strings.Join(current, ", ")
			}
			warn(warnings, fmt.Sprintf("Issue '%s' does not depend on '%s', so nothing was removed. It has %s.", name, issueName(*blocking), has))
			return relative(root, issue.path), nil
		}
		current = kept
	default:
		added, err := dependsOnValueFor(x, *blocking, name, warnings)
		if err != nil {
			return "", err
		}
		if !slices.Contains(current, added) {
			current = append(current, added)
		}
	}
	doc.set("depends-on", formatDependsOn(current))
	if err := doc.save(issue.path); err != nil {
		return "", err
	}
	return relative(root, issue.path), nil
}

// dependsOnValueFor validates that child can depend on reference: it must name an existing issue,
// not be child itself, and not already (directly or transitively) depend on child, which would form
// a cycle.
func dependsOnValueFor(x *issueIndex, reference, child string, warnings io.Writer) (string, error) {
	if issueName(reference) == "" {
		return "", fmt.Errorf("'%s' does not name an issue.", reference)
	}
	match, found, err := x.match(issueName(reference))
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("The issue '%s' does not exist in %s.", issueName(reference), stateDirectoryList)
	}
	blocking := match.name
	if blocking == child {
		return "", fmt.Errorf("Issue '%s' cannot depend on itself.", child)
	}
	if dependsOnReaches(x, blocking, child, map[string]bool{}) {
		return "", fmt.Errorf("Issue '%s' already depends on '%s'. Adding this dependency would create a cycle.", blocking, child)
	}
	if match.state == "closed" {
		warn(warnings, fmt.Sprintf("The issue '%s' is closed.", blocking))
	}
	return blocking, nil
}

// dependsOnReaches reports whether target is reachable from start by following depends-on edges.
func dependsOnReaches(x *issueIndex, start, target string, seen map[string]bool) bool {
	if start == target {
		return true
	}
	if seen[start] {
		return false
	}
	seen[start] = true
	match, found, err := x.match(start)
	if err != nil || !found {
		return false
	}
	doc, err := x.document(match)
	if err != nil {
		return false
	}
	for _, next := range parseDependsOn(doc.get("depends-on")) {
		if dependsOnReaches(x, x.canonical(next), target, seen) {
			return true
		}
	}
	return false
}

// dependents returns every issue whose depends-on list includes name.
func dependents(x *issueIndex, name string) []subIssue {
	var result []subIssue
	for _, issue := range x.issues {
		doc, err := x.document(issue)
		if err != nil {
			continue
		}
		if slices.ContainsFunc(parseDependsOn(doc.get("depends-on")), func(dep string) bool { return x.canonical(dep) == name }) {
			result = append(result, subIssue{name: issue.name, state: issue.state})
		}
	}
	return result
}

func issuePath(root, state, name string) string {
	return filepath.Join(root, "issues", state, filepath.FromSlash(name)+".md")
}

// nameClashes returns every Markdown file in the repo (sorted, relative to root) whose name without
// ".md" equals name, case-insensitively. A link to that bare name is ambiguous.
func nameClashes(root, name string) []string {
	return markdownNames(root)[strings.ToLower(name)]
}

// markdownNames reads the repo once and maps each Markdown file name, without ".md" and in lower
// case, to the sorted paths of the files that have it, relative to root.
func markdownNames(root string) map[string][]string {
	names := map[string][]string{}
	for _, file := range repoFiles(root) {
		base := filepath.Base(file)
		if hasSuffixFold(base, ".md") {
			key := strings.ToLower(trimMarkdownExtension(base))
			names[key] = append(names[key], relative(root, file))
		}
	}
	for _, paths := range names {
		slices.Sort(paths)
	}
	return names
}

var slugWordPattern = regexp.MustCompile(`[a-z0-9]+`)

// asciiSpellings covers the Latin letters that Unicode does not decompose into a base letter and a mark.
var asciiSpellings = strings.NewReplacer(
	"ß", "ss", "æ", "ae", "œ", "oe", "ø", "o", "ł", "l", "đ", "d", "ð", "d", "þ", "th",
	"ı", "i", "ħ", "h", "ŋ", "ng", "ŧ", "t",
)

// fillerWords add no meaning to an issue name. slugify drops them before it keeps 6 words, so that
// the 6 words carry the meaning of a long title.
var fillerWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "of": true, "to": true,
	"in": true, "on": true, "at": true, "by": true, "for": true, "with": true, "from": true,
	"into": true, "is": true, "are": true, "was": true, "were": true, "be": true,
}

// fillerWordList is fillerWords in sorted order, for help text.
func fillerWordList() string {
	words := make([]string, 0, len(fillerWords))
	for w := range fillerWords {
		words = append(words, w)
	}
	slices.Sort(words)
	return strings.Join(words, ", ")
}

// toASCII lowercases text and spells its letters in ASCII where it can: accents go, and
// asciiSpellings covers letters with no decomposition. A character with no ASCII form stays, and
// slugWordPattern then treats it as a separator.
func toASCII(text string) string {
	// NFKD can produce capitals, for example "™" -> "TM", so lowercase again after it.
	decomposed := strings.ToLower(norm.NFKD.String(asciiSpellings.Replace(strings.ToLower(text))))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, decomposed)
}

// slugify turns text into a kebab-case ASCII name of 2-6 words, the shape namePattern requires, for
// example "Fix the naïve parser!" -> "fix-naive-parser". Filler words drop first, unless fewer than
// 2 words would remain.
func slugify(text string) (string, error) {
	words := slugWordPattern.FindAllString(toASCII(text), -1)
	if meaningful := slices.DeleteFunc(slices.Clone(words), func(w string) bool { return fillerWords[w] }); len(meaningful) >= 2 {
		words = meaningful
	}
	if len(words) > 6 {
		words = words[:6]
	}
	if len(words) < 2 {
		return "", fmt.Errorf("'%s' does not have enough words to make a valid issue name.", text)
	}
	return strings.Join(words, "-"), nil
}

// uniqueSlug makes a slug from text and, if it already names a file in the repo, appends -2, -3, and
// so on until it does not, trimming words from the base as needed to stay within namePattern's 6-word
// limit. The result always matches namePattern.
func uniqueSlug(root, text string) (string, error) {
	base, err := slugify(text)
	if err != nil {
		return "", err
	}
	names := markdownNames(root)
	if len(names[base]) == 0 {
		return base, nil
	}
	words := strings.Split(base, "-")
	for n := 2; ; n++ {
		suffixed := words
		if len(words)+1 > 6 {
			suffixed = words[:5]
		}
		candidate := strings.Join(append(append([]string{}, suffixed...), strconv.Itoa(n)), "-")
		if len(names[candidate]) == 0 {
			return candidate, nil
		}
	}
}

// saveIssue writes the document to "from" and moves it to "to", the directory of its state. It
// moves the file through git when git tracks it, so the move lands in the git index as a rename
// instead of as an untracked delete-and-add. A file that git does not track gets a plain rename:
// git mv refuses it. In a colocated jj repo that is every file that jj has not committed yet, and
// jj detects the rename from content when it next snapshots the working copy.
func saveIssue(root string, doc *document, from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	if err := doc.save(from); err != nil {
		return err
	}
	if from == to {
		return nil
	}
	var err error
	if usesGit(root) && gitTracks(root, from) {
		err = gitMove(root, from, to)
	} else {
		err = os.Rename(from, to)
	}
	if err != nil {
		return err
	}
	removeEmptyDirectories(root, filepath.Dir(from))
	return nil
}

// removeEmptyDirectories removes dir and each directory above it that is now empty, up to but not
// including the state directory, for example issues/open. os.Remove fails for a directory that is
// not empty, which ends the walk.
func removeEmptyDirectories(root, dir string) {
	issues := filepath.Join(root, "issues")
	for filepath.Dir(dir) != issues && dir != issues && strings.HasPrefix(dir, issues) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// usesGit reports whether root is the root of a git repo (colocated git+jj, or git alone).
func usesGit(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

// gitTracks reports whether the git index of root has the file at path.
func gitTracks(root, path string) bool {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", "--", path)
	cmd.Dir = root
	return cmd.Run() == nil
}

func gitMove(root, from, to string) error {
	cmd := exec.Command("git", "mv", from, to)
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git mv %s %s: %w: %s", relative(root, from), relative(root, to), err, strings.TrimSpace(string(output)))
	}
	return nil
}

// parseList reads a frontmatter value such as "[a, b]" into its items. An empty or missing value is an empty list.
func parseList(value string) []string {
	value = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "["), "]")
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func formatList(items []string) string {
	return "[" + strings.Join(items, ", ") + "]"
}

// validateTags rejects a tag with whitespace in a comma-separated list. Obsidian reads the tags
// field as tags, and a tag there cannot contain a space.
func validateTags(raw string) error {
	for _, tag := range parseList(raw) {
		if strings.IndexFunc(tag, unicode.IsSpace) >= 0 {
			return fmt.Errorf("The tag '%s' has a space. A tag is one word; use '%s'.", tag, strings.Join(strings.Fields(tag), "-"))
		}
	}
	return nil
}

// formatTags parses a comma-separated list into the frontmatter's bracketed form, in lowercase, for
// example "plan, Widget-Scheduler ,,repo" -> "[plan, widget-scheduler, repo]". Tags are stored in
// lowercase, like the built-in tags, and compared without regard to case (see containsFold).
func formatTags(raw string) string {
	return formatList(parseList(strings.ToLower(raw)))
}

// containsFold reports whether items contains value, without regard to case.
func containsFold(items []string, value string) bool {
	return slices.ContainsFunc(items, func(item string) bool { return strings.EqualFold(item, value) })
}

// addToList appends value to the frontmatter list field key, unless it is already there.
func addToList(doc *document, key, value string) {
	items := parseList(doc.get(key))
	if !slices.Contains(items, value) {
		items = append(items, value)
	}
	doc.set(key, formatList(items))
}

func checkResolution(state, resolution, unexpectedMessage string) error {
	if state != "closed" && resolution != "" {
		return errors.New(unexpectedMessage)
	}
	return nil
}

// resolveResolution defaults an empty resolution to "completed" when closing, the common case, so
// --resolution is only needed to mark an issue "abandoned". An explicit --resolution always wins.
func resolveResolution(state, resolution string) string {
	if state == "closed" && resolution == "" {
		return "completed"
	}
	return resolution
}

// deferredValue writes a deferred-from reference: an issue becomes a wikilink to its full name, and
// any other value with "/" is a bookmark, written as it is. A wikilink, or a value without "/", that
// names no issue is still written, as a wikilink, with a warning: it may name an issue that was
// removed on purpose, but it is more often a typo.
func deferredValue(x *issueIndex, reference string, warnings io.Writer) (string, error) {
	r := strings.TrimSpace(reference)
	explicitLink := strings.HasPrefix(r, "[[")
	match, found, err := x.match(issueName(r))
	if err != nil {
		return "", err
	}
	switch {
	case found && !explicitLink:
		return quote("[[" + match.name + "]]"), nil
	case found:
		return quote(r), nil
	case strings.Contains(r, "/") && !explicitLink:
		return quote(r), nil
	}
	warn(warnings, fmt.Sprintf("--deferred-from names no issue: '%s'. It is written as the link [[%s]]. A bookmark must contain '/'.", issueName(r), issueName(r)))
	if explicitLink {
		return quote(r), nil
	}
	return quote("[[" + issueName(r) + "]]"), nil
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

// issueView is every frontmatter field of an issue, plus its body, for "sisyphus show".
type issueView struct {
	Name         string            `json:"name"`
	Title        string            `json:"title"`
	State        string            `json:"state"`
	Resolution   string            `json:"resolution,omitempty"`
	Priority     string            `json:"priority"`
	Effort       string            `json:"effort"`
	Tags         []string          `json:"tags,omitempty"`
	Created      string            `json:"created"`
	Closed       string            `json:"closed,omitempty"`
	Owner        string            `json:"owner,omitempty"`
	Approver     string            `json:"approver,omitempty"`
	Bookmark     string            `json:"bookmark,omitempty"`
	Workspaces   []string          `json:"workspaces,omitempty"`
	DeferredFrom string            `json:"deferred-from,omitempty"`
	Parent       string            `json:"parent,omitempty"`
	DependsOn    []string          `json:"depends-on,omitempty"`
	Remote       string            `json:"remote,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Body         string            `json:"body"`
}

func newIssueView(name string, doc *document) issueView {
	return issueView{
		Name:         name,
		Title:        doc.get("title"),
		State:        doc.get("state"),
		Resolution:   doc.get("resolution"),
		Priority:     doc.get("priority"),
		Effort:       doc.get("effort"),
		Tags:         parseList(doc.get("tags")),
		Created:      doc.get("created"),
		Closed:       doc.get("closed"),
		Owner:        doc.get("owner"),
		Approver:     doc.get("approver"),
		Bookmark:     doc.get("bookmark"),
		Workspaces:   parseList(doc.get("workspaces")),
		DeferredFrom: doc.get("deferred-from"),
		Parent:       doc.get("parent"),
		DependsOn:    parseDependsOn(doc.get("depends-on")),
		Remote:       doc.get("remote"),
		Metadata:     parseMetadata(doc.get("metadata")),
		Body:         strings.Join(doc.body, "\n"),
	}
}
