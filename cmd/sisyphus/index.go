package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

// issueIndex is every issue that one command can see. A command builds it once, so that the command
// reads each state directory once and each issue file at most once. It is never kept between
// commands: each command must see the edits that other sessions make in the meantime.
type issueIndex struct {
	root   string
	issues []issueFile
	// exact maps a full name to its files, one for each state directory that has it.
	exact map[string][]issueFile
	// shorter maps each shorter end of a full name, such as the bare file name, to its issues.
	shorter map[string][]issueFile
	walked  bool
	docs    map[string]loadedDocument
	// headers holds documents that were read only up to the end of their frontmatter.
	headers map[string]loadedDocument
}

type loadedDocument struct {
	doc *document
	err error
}

// traceRead is nil except in tests, which count the reads of one command with it.
var traceRead func(path string)

// traceMu serializes the calls of traceRead, because preload reads files on several goroutines.
var traceMu sync.Mutex

func trace(path string) {
	if traceRead != nil {
		traceMu.Lock()
		defer traceMu.Unlock()
		traceRead(path)
	}
}

// preloadMinimum is the smallest number of files that preload reads on a pool of goroutines. For fewer
// files, starting the workers costs more than it saves.
const preloadMinimum = 32

// preload reads the files of issues that the command has not read yet, with a bounded pool of
// goroutines, and keeps them in the same caches that document and header use, so that a command that
// needs many issues does not wait for one file at a time. Each file is still read once. withBody
// reads whole files; without it, only the frontmatter.
func (x *issueIndex) preload(issues []issueFile, withBody bool) {
	cache, load := x.headers, loadHeader
	if withBody {
		cache, load = x.docs, loadDocument
	}
	var todo []issueFile
	for _, issue := range issues {
		_, full := x.docs[issue.path]
		_, cached := cache[issue.path]
		if !full && !cached {
			todo = append(todo, issue)
		}
	}
	if len(todo) < preloadMinimum {
		return
	}
	results := make([]loadedDocument, len(todo))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), 16) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				doc, err := load(todo[i].path)
				results[i] = loadedDocument{doc: doc, err: err}
			}
		}()
	}
	for i := range todo {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for i, issue := range todo {
		cache[issue.path] = results[i]
	}
}

// loadIndex reads the state directories and their subdirectories once. Issues are sorted by full name.
func loadIndex(root string) *issueIndex {
	return &issueIndex{
		root:    root,
		exact:   map[string][]issueFile{},
		shorter: map[string][]issueFile{},
		docs:    map[string]loadedDocument{},
		headers: map[string]loadedDocument{},
	}
}

// walk reads the state directories and their subdirectories, once, on the first call. A command that
// only looks up issues by full name never needs it (see exactOf). Issues are sorted by full name.
func (x *issueIndex) walk() {
	if x.walked {
		return
	}
	x.walked = true
	for _, state := range states {
		directory := filepath.Join(x.root, "issues", state)
		trace(directory)
		_ = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !hasSuffixFold(entry.Name(), ".md") {
				return nil
			}
			x.issues = append(x.issues, issueFile{state: state, path: path, name: trimMarkdownExtension(relative(directory, path))})
			return nil
		})
	}
	slices.SortStableFunc(x.issues, func(a, b issueFile) int { return strings.Compare(a.name, b.name) })
	for _, issue := range x.issues {
		x.exact[issue.name] = append(x.exact[issue.name], issue)
		for i, r := range issue.name {
			if r == '/' {
				end := issue.name[i+1:]
				x.shorter[end] = append(x.shorter[end], issue)
			}
		}
	}
}

// all returns every issue, sorted by full name.
func (x *issueIndex) all() []issueFile {
	x.walk()
	return x.issues
}

// exactOf returns the files of the issue with the full name name, one for each state directory that
// has it. Before the walk, it looks at issues/<state>/<name>.md directly for a valid (lowercase) name,
// so that a command that gets a full name does not read every state directory.
func (x *issueIndex) exactOf(name string) []issueFile {
	if x.walked || !validName(name) {
		x.walk()
		return x.exact[name]
	}
	var found []issueFile
	for _, state := range states {
		path := issuePath(x.root, state, name)
		if fileExists(path) {
			found = append(found, issueFile{state: state, path: path, name: name})
		}
	}
	return found
}

// matches returns the issue files that name gives. A full name gives its own issue, even when it is
// also the end of a longer name. Otherwise name gives every issue whose full name ends with it, so
// that a bare file name gives the issue in any subdirectory. A name that matches nothing is tried
// again in lowercase: valid issue names are lowercase, and wikilinks ignore case, as in Obsidian.
func (x *issueIndex) matches(name string) []issueFile {
	if exact := x.exactOf(name); len(exact) > 0 {
		return exact
	}
	x.walk()
	if shorter := x.shorter[name]; len(shorter) > 0 {
		return shorter
	}
	if lower := strings.ToLower(name); lower != name {
		return x.matches(lower)
	}
	return nil
}

// match finds the one issue that name gives. found is false when no issue matches. A name that gives
// more than one issue, or one issue in more than one state directory, is an error.
func (x *issueIndex) match(name string) (issueFile, bool, error) {
	matches := x.matches(name)
	if len(matches) == 0 {
		return issueFile{}, false, nil
	}
	var names, paths []string
	for _, m := range matches {
		if !slices.Contains(names, m.name) {
			names = append(names, m.name)
		}
		paths = append(paths, relative(x.root, m.path))
	}
	if len(names) > 1 {
		return issueFile{}, false, fmt.Errorf("'%s' matches more than one issue: %s. Use the full name, for example %s.", name, strings.Join(names, ", "), names[0])
	}
	if len(matches) > 1 {
		return issueFile{}, false, fmt.Errorf("Issue '%s' exists in more than one state directory: %s. Remove the duplicate.", names[0], strings.Join(paths, ", "))
	}
	return matches[0], true, nil
}

func (x *issueIndex) find(name string) (issueFile, error) {
	match, found, err := x.match(name)
	if err != nil {
		return issueFile{}, err
	}
	if !found {
		return issueFile{}, fmt.Errorf("No issue named '%s' in %s.", name, stateDirectoryList)
	}
	return match, nil
}

// canonical returns the full name of the one issue that reference gives. If no single issue matches,
// it returns the name in reference unchanged.
func (x *issueIndex) canonical(reference string) string {
	name := issueName(reference)
	if match, found, err := x.match(name); err == nil && found {
		return match.name
	}
	return name
}

// document parses the file of issue on the first call, and returns the same document after that. A
// caller that changes the document changes it for the rest of the command.
func (x *issueIndex) document(issue issueFile) (*document, error) {
	if loaded, ok := x.docs[issue.path]; ok {
		return loaded.doc, loaded.err
	}
	doc, err := loadDocument(issue.path)
	x.docs[issue.path] = loadedDocument{doc: doc, err: err}
	return doc, err
}

// header returns the frontmatter of issue, for a caller that does not use the body: the full document
// when the command already read it, or else a document read only up to the end of its frontmatter.
// A caller must not save it; save refuses a header.
func (x *issueIndex) header(issue issueFile) (*document, error) {
	if loaded, ok := x.docs[issue.path]; ok {
		return loaded.doc, loaded.err
	}
	if loaded, ok := x.headers[issue.path]; ok {
		return loaded.doc, loaded.err
	}
	doc, err := loadHeader(issue.path)
	x.headers[issue.path] = loadedDocument{doc: doc, err: err}
	return doc, err
}

// load finds the issue that reference gives and returns its full name, its file, and its document.
func (x *issueIndex) load(reference string) (string, issueFile, *document, error) {
	issue, err := x.find(issueName(reference))
	if err != nil {
		return "", issueFile{}, nil, err
	}
	doc, err := x.document(issue)
	return issue.name, issue, doc, err
}
