package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	stateDirectoryList = "issues/open, issues/in-progress, or issues/closed"
	noBookmarkWarning  = "The issue is in-progress but has no bookmark. Set it with --bookmark."
)

// newOptions holds the values of "sisyphus new". An empty resolution and a nil pointer mean that the option was not given.
type newOptions struct {
	name, state, resolution, priority, effort, tags string
	title, bookmark, deferredFrom, parent           *string
	owner, approver, workspace                      *string
	remote, dependsOn, context                      *string
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

// baseName is the last part of a full issue name, which is also its file name without ".md".
func baseName(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

func newIssue(root string, o newOptions, warnings io.Writer) (string, error) {
	if !validName(o.name) {
		return "", fmt.Errorf("Invalid issue name '%s'. Use 2-6 lowercase words in kebab-case, for example fix-login-bug, optionally below kebab-case directories, for example web/auth/fix-login-bug.", o.name)
	}
	resolution := resolveResolution(o.state, o.resolution)
	if err := checkResolution(o.state, resolution, "Set --resolution only when the state is closed."); err != nil {
		return "", err
	}

	if existing := exactIssues(root, o.name); len(existing) > 0 {
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
		value, err := parentValueFor(root, *o.parent, o.name, warnings)
		if err != nil {
			return "", err
		}
		parentValue = value
	}
	var dependsOnValue []string
	if o.dependsOn != nil {
		value, err := dependsOnValueFor(root, *o.dependsOn, o.name, warnings)
		if err != nil {
			return "", err
		}
		dependsOnValue = []string{value}
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
		title = *o.title
	}
	deferred := ""
	if o.deferredFrom != nil {
		value, err := deferredValue(root, *o.deferredFrom)
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
	if o.dependsOn != nil {
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
	priority, effort, tags               *string
	metadata                             map[string]string
}

func updateIssue(root, reference, state string, o updateOptions, warnings io.Writer) (string, error) {
	name, current, doc, err := loadIssue(root, reference)
	if err != nil {
		return "", err
	}
	if current.state == "closed" && state != "closed" {
		return "", fmt.Errorf("Issue '%s' is closed. Do not reopen a closed issue. Create a new issue and link to [[%s]].", name, name)
	}
	resolution := resolveResolution(state, o.resolution)
	if err := checkResolution(state, resolution, "Set --resolution only when the new state is closed."); err != nil {
		return "", err
	}
	if recorded := doc.get("state"); recorded != current.state {
		warn(warnings, fmt.Sprintf("The issue was in issues/%s/ but its state field was '%s'. The update corrects both.", current.state, recorded))
	}

	applyState(doc, state, resolution)
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
		doc.set("tags", formatTags(*o.tags))
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
		var blocked []string
		for _, dep := range dependents(root, name) {
			if dep.state != "closed" {
				blocked = append(blocked, dep.name)
			}
		}
		if len(blocked) > 0 {
			warn(warnings, fmt.Sprintf("Issue '%s' is closing, but these issues still depend on it: %s.", name, strings.Join(blocked, ", ")))
		}
	}

	destination := issuePath(root, state, name)
	if err := saveIssue(root, doc, current.path, destination); err != nil {
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

func setRemote(root, reference string, remote *string, warnings io.Writer) (string, error) {
	_, issue, doc, err := loadIssue(root, reference)
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

func parentValueFor(root, reference, child string, warnings io.Writer) (string, error) {
	if issueName(reference) == "" {
		return "", fmt.Errorf("'%s' does not name an issue.", reference)
	}
	match, found, err := matchIssue(root, issueName(reference))
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
	for ancestor := parentName(root, parent); ancestor != "" && !seen[ancestor]; ancestor = parentName(root, ancestor) {
		if ancestor == child {
			return "", fmt.Errorf("Issue '%s' is already below '%s'. A parent of '%s' cannot be one of its sub-issues.", parent, child, child)
		}
		seen[ancestor] = true
	}
	return quote("[[" + parent + "]]"), nil
}

// parentName returns the full name of the parent of the issue name, or "" when it has none.
func parentName(root, name string) string {
	match, found, err := matchIssue(root, name)
	if err != nil || !found {
		return ""
	}
	doc, err := loadDocument(match.path)
	if err != nil {
		return ""
	}
	return canonicalName(root, doc.get("parent"))
}

type subIssue struct {
	name  string
	state string
}

func subIssues(root, name string) []subIssue {
	var result []subIssue
	for _, issue := range allIssues(root) {
		doc, err := loadDocument(issue.path)
		if err != nil || canonicalName(root, doc.get("parent")) != name {
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

// parseMetadata reads a frontmatter value like `{session-id: "abc123", note: "a short note"}` into
// a map. Like tags and the other list-shaped fields, a value cannot contain a comma.
func parseMetadata(value string) map[string]string {
	value = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "{"), "}")
	m := map[string]string{}
	for _, item := range strings.Split(value, ",") {
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
	name, issue, doc, err := loadIssue(root, reference)
	if err != nil {
		return "", err
	}
	current := parseDependsOn(doc.get("depends-on"))
	switch {
	case blocking == nil: // clear is guaranteed true by the caller: clear every dependency
		current = nil
	case clear:
		target := canonicalName(root, *blocking)
		current = slices.DeleteFunc(current, func(n string) bool { return canonicalName(root, n) == target })
	default:
		added, err := dependsOnValueFor(root, *blocking, name, warnings)
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
func dependsOnValueFor(root, reference, child string, warnings io.Writer) (string, error) {
	if issueName(reference) == "" {
		return "", fmt.Errorf("'%s' does not name an issue.", reference)
	}
	match, found, err := matchIssue(root, issueName(reference))
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
	if dependsOnReaches(root, blocking, child, map[string]bool{}) {
		return "", fmt.Errorf("Issue '%s' already depends on '%s'. Adding this dependency would create a cycle.", blocking, child)
	}
	if match.state == "closed" {
		warn(warnings, fmt.Sprintf("The issue '%s' is closed.", blocking))
	}
	return blocking, nil
}

// dependsOnReaches reports whether target is reachable from start by following depends-on edges.
func dependsOnReaches(root, start, target string, seen map[string]bool) bool {
	if start == target {
		return true
	}
	if seen[start] {
		return false
	}
	seen[start] = true
	match, found, err := matchIssue(root, start)
	if err != nil || !found {
		return false
	}
	doc, err := loadDocument(match.path)
	if err != nil {
		return false
	}
	for _, next := range parseDependsOn(doc.get("depends-on")) {
		if dependsOnReaches(root, canonicalName(root, next), target, seen) {
			return true
		}
	}
	return false
}

// dependents returns every issue whose depends-on list includes name.
func dependents(root, name string) []subIssue {
	var result []subIssue
	for _, issue := range allIssues(root) {
		doc, err := loadDocument(issue.path)
		if err != nil {
			continue
		}
		if slices.ContainsFunc(parseDependsOn(doc.get("depends-on")), func(dep string) bool { return canonicalName(root, dep) == name }) {
			result = append(result, subIssue{name: issue.name, state: issue.state})
		}
	}
	return result
}

func issuePath(root, state, name string) string {
	return filepath.Join(root, "issues", state, filepath.FromSlash(name)+".md")
}

// allIssues returns every issue in the state directories and their subdirectories, sorted by name.
func allIssues(root string) []issueFile {
	var result []issueFile
	for _, state := range states {
		directory := filepath.Join(root, "issues", state)
		_ = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !hasSuffixFold(entry.Name(), ".md") {
				return nil
			}
			result = append(result, issueFile{state: state, path: path, name: trimMarkdownExtension(relative(directory, path))})
			return nil
		})
	}
	slices.SortStableFunc(result, func(a, b issueFile) int { return strings.Compare(a.name, b.name) })
	return result
}

// exactIssues returns the issue files whose full name is name, one for each state directory that has it.
func exactIssues(root, name string) []issueFile {
	var matches []issueFile
	for _, issue := range allIssues(root) {
		if issue.name == name {
			matches = append(matches, issue)
		}
	}
	return matches
}

// issueMatches returns the issue files that name gives. A full name gives its own issue, even when
// it is also the end of a longer name. Otherwise name gives every issue whose full name ends with
// it, so that a bare file name gives the issue in any subdirectory.
func issueMatches(root, name string) []issueFile {
	if exact := exactIssues(root, name); len(exact) > 0 {
		return exact
	}
	var matches []issueFile
	for _, issue := range allIssues(root) {
		if strings.HasSuffix(issue.name, "/"+name) {
			matches = append(matches, issue)
		}
	}
	return matches
}

// matchIssue finds the one issue that name gives. found is false when no issue matches. A name
// that gives more than one issue, or one issue in more than one state directory, is an error.
func matchIssue(root, name string) (issueFile, bool, error) {
	matches := issueMatches(root, name)
	if len(matches) == 0 {
		return issueFile{}, false, nil
	}
	var names, paths []string
	for _, m := range matches {
		if !slices.Contains(names, m.name) {
			names = append(names, m.name)
		}
		paths = append(paths, relative(root, m.path))
	}
	if len(names) > 1 {
		return issueFile{}, false, fmt.Errorf("'%s' matches more than one issue: %s. Use the full name, for example %s.", name, strings.Join(names, ", "), names[0])
	}
	if len(matches) > 1 {
		return issueFile{}, false, fmt.Errorf("Issue '%s' exists in more than one state directory: %s. Remove the duplicate.", names[0], strings.Join(paths, ", "))
	}
	return matches[0], true, nil
}

// canonicalName returns the full name of the one issue that reference gives. If no single issue
// matches, it returns the name in reference unchanged.
func canonicalName(root, reference string) string {
	name := issueName(reference)
	if match, found, err := matchIssue(root, name); err == nil && found {
		return match.name
	}
	return name
}

// nameClashes returns every Markdown file in the repo (sorted, relative to root) whose name without
// ".md" equals name, case-insensitively. An issue name must be unique across all of them, not just
// among other issues.
func nameClashes(root, name string) []string {
	var clashes []string
	for _, file := range repoFiles(root) {
		base := filepath.Base(file)
		if hasSuffixFold(base, ".md") && strings.EqualFold(trimMarkdownExtension(base), name) {
			clashes = append(clashes, relative(root, file))
		}
	}
	slices.Sort(clashes)
	return clashes
}

var slugWordPattern = regexp.MustCompile(`[a-z0-9]+`)

// slugify turns text into a kebab-case name of 2-6 words, the shape namePattern requires, for
// example "Fix the login bug!" -> "fix-the-login-bug".
func slugify(text string) (string, error) {
	words := slugWordPattern.FindAllString(strings.ToLower(text), -1)
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
	if len(nameClashes(root, base)) == 0 {
		return base, nil
	}
	words := strings.Split(base, "-")
	for n := 2; ; n++ {
		suffixed := words
		if len(words)+1 > 6 {
			suffixed = words[:5]
		}
		candidate := strings.Join(append(append([]string{}, suffixed...), strconv.Itoa(n)), "-")
		if len(nameClashes(root, candidate)) == 0 {
			return candidate, nil
		}
	}
}

func findIssue(root, name string) (issueFile, error) {
	match, found, err := matchIssue(root, name)
	if err != nil {
		return issueFile{}, err
	}
	if !found {
		return issueFile{}, fmt.Errorf("No issue named '%s' in %s.", name, stateDirectoryList)
	}
	return match, nil
}

// loadIssue finds the issue that reference gives and returns its full name, its file, and its document.
func loadIssue(root, reference string) (string, issueFile, *document, error) {
	issue, err := findIssue(root, issueName(reference))
	if err != nil {
		return "", issueFile{}, nil, err
	}
	doc, err := loadDocument(issue.path)
	return issue.name, issue, doc, err
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
	if usesGit(root) && gitTracks(root, from) {
		return gitMove(root, from, to)
	}
	return os.Rename(from, to)
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

// formatTags parses a comma-separated list into the frontmatter's bracketed form, for example
// "plan, widget-scheduler ,,repo" -> "[plan, widget-scheduler, repo]".
func formatTags(raw string) string {
	return formatList(parseList(raw))
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
// any other value with "/" is a bookmark, written as it is.
func deferredValue(root, reference string) (string, error) {
	r := strings.TrimSpace(reference)
	if strings.HasPrefix(r, "[[") {
		return quote(r), nil
	}
	match, found, err := matchIssue(root, issueName(r))
	if err != nil {
		return "", err
	}
	switch {
	case found:
		return quote("[[" + match.name + "]]"), nil
	case strings.Contains(r, "/"):
		return quote(r), nil
	default:
		return quote("[[" + issueName(r) + "]]"), nil
	}
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
