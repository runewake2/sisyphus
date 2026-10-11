package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// repoFiles returns every file of the repo that a wikilink can point to, as absolute paths, without
// the files that .gitignore ignores, so that an ignored directory such as vendor/ is never read. In a
// git repo, git decides what is ignored (see gitIgnored), with every rule of every .gitignore;
// otherwise the simple rules of the root .gitignore apply (see gitignoreRules). It never returns
// files below .git, .jj, or the other skippedDirectories.
func repoFiles(root string) []string {
	ignored, ok := gitIgnored(root)
	if !ok {
		ignored = gitignoreRules(root)
	}
	var files []string
	_ = filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil || file == root {
			return nil
		}
		rel := relative(root, file)
		if entry.IsDir() {
			if skippedDirectories[strings.ToLower(entry.Name())] || ignored(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if !ignored(rel, false) {
			files = append(files, file)
		}
		return nil
	})
	return files
}

// gitIgnored asks git once for the ignored paths of the working copy. With --directory, git gives an
// ignored directory as one path that ends with "/", and does not list what is in it. A directory that
// holds a tracked file is not given as a whole. ok is false when root is not the top of a git working
// copy or git cannot run.
func gitIgnored(root string) (func(rel string, isDir bool) bool, bool) {
	if !usesGit(root) {
		return nil, false
	}
	cmd := exec.Command("git", "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--directory")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	ignored := map[string]bool{}
	for _, rel := range bytes.Split(output, []byte{0}) {
		if len(rel) > 0 {
			ignored[string(rel)] = true
		}
	}
	return func(rel string, isDir bool) bool {
		if isDir {
			return ignored[rel+"/"]
		}
		return ignored[rel]
	}, true
}

// gitignoreRules reads the root .gitignore for a repo without git, and returns a function that tells
// whether a path (relative to the root, with "/") is ignored. It supports the common subset: blank
// lines and "#" comments, a name or glob that matches any path part ("vendor", "*.log"), a pattern
// with "/" that matches from the root ("/dist", "docs/build"), and a trailing "/" for a directory
// only. A negated pattern ("!keep") is not supported and is skipped.
func gitignoreRules(root string) func(rel string, isDir bool) bool {
	content, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return func(string, bool) bool { return false }
	}
	type rule struct {
		pattern           string
		anchored, dirOnly bool
	}
	var rules []rule
	for _, line := range splitLines(string(content)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		r := rule{dirOnly: strings.HasSuffix(line, "/")}
		line = strings.TrimSuffix(line, "/")
		r.anchored = strings.Contains(line, "/")
		r.pattern = strings.TrimPrefix(line, "/")
		if r.pattern != "" {
			rules = append(rules, r)
		}
	}
	return func(rel string, isDir bool) bool {
		for _, r := range rules {
			if r.dirOnly && !isDir {
				continue
			}
			if r.anchored {
				if ok, _ := path.Match(r.pattern, rel); ok {
					return true
				}
				continue
			}
			if ok, _ := path.Match(r.pattern, path.Base(rel)); ok {
				return true
			}
		}
		return false
	}
}
