package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
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
	docs    map[string]loadedDocument
}

type loadedDocument struct {
	doc *document
	err error
}

// traceRead is nil except in tests, which count the reads of one command with it.
var traceRead func(path string)

// loadIndex reads the state directories and their subdirectories once. Issues are sorted by full name.
func loadIndex(root string) *issueIndex {
	x := &issueIndex{
		root:    root,
		exact:   map[string][]issueFile{},
		shorter: map[string][]issueFile{},
		docs:    map[string]loadedDocument{},
	}
	for _, state := range states {
		directory := filepath.Join(root, "issues", state)
		if traceRead != nil {
			traceRead(directory)
		}
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
	return x
}

// matches returns the issue files that name gives. A full name gives its own issue, even when it is
// also the end of a longer name. Otherwise name gives every issue whose full name ends with it, so
// that a bare file name gives the issue in any subdirectory.
func (x *issueIndex) matches(name string) []issueFile {
	if exact := x.exact[name]; len(exact) > 0 {
		return exact
	}
	return x.shorter[name]
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

// load finds the issue that reference gives and returns its full name, its file, and its document.
func (x *issueIndex) load(reference string) (string, issueFile, *document, error) {
	issue, err := x.find(issueName(reference))
	if err != nil {
		return "", issueFile{}, nil, err
	}
	doc, err := x.document(issue)
	return issue.name, issue, doc, err
}
