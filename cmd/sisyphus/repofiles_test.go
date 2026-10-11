package main

import (
	"os"
	"slices"
	"testing"
)

func relativeFiles(r *testRepo) []string {
	var rels []string
	for _, file := range repoFiles(r.root) {
		rels = append(rels, relative(r.root, file))
	}
	slices.Sort(rels)
	return rels
}

func TestRepoFilesInAGitRepoFollowGitignore(t *testing.T) {
	r := newTestRepo(t)
	r.write(".gitignore", "vendor/\n*.log\n")
	r.write("vendor/lib/README.md", "# Vendored\n")
	r.write("debug.log", "x\n")
	r.write("design/tracked.md", "# Tracked\n")
	r.write("design/gone.md", "# Gone\n")
	r.write("docs/.gitignore", "*.md\n!keep.md\n")
	r.write("docs/drop.md", "# Drop\n")
	r.write("docs/keep.md", "# Keep\n")
	r.write("build/tracked.md", "# Tracked in an ignored directory\n")
	r.git("init", "-q")
	r.git("add", "-A")
	r.git("add", "-f", "build/tracked.md")
	r.write(".gitignore", "vendor/\n*.log\nbuild/\n")
	r.write("build/untracked.md", "# Untracked in an ignored directory\n")
	r.git("-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "init")
	r.write("design/untracked.md", "# Untracked\n")
	if err := os.Remove(r.pathOf("design/gone.md")); err != nil {
		t.Fatal(err)
	}

	files := relativeFiles(r)

	isTrue(t, slices.Contains(files, "design/tracked.md"), "a tracked file is listed")
	isTrue(t, slices.Contains(files, "design/untracked.md"), "an untracked file that is not ignored is listed")
	isTrue(t, !slices.Contains(files, "design/gone.md"), "a deleted file is not listed")
	isTrue(t, !slices.Contains(files, "vendor/lib/README.md"), "an ignored directory is not listed")
	isTrue(t, !slices.Contains(files, "debug.log"), "an ignored file is not listed")
	isTrue(t, !slices.Contains(files, "docs/drop.md"), "a nested .gitignore applies")
	isTrue(t, slices.Contains(files, "docs/keep.md"), "a negated pattern applies")
	isTrue(t, slices.Contains(files, "build/tracked.md"), "a tracked file in an ignored directory is listed")
	isTrue(t, !slices.Contains(files, "build/untracked.md"), "an untracked file in an ignored directory is not listed")
	equal(t, 1, r.run("resolve", "[[vendor/lib/README]]").exit)
}

func TestRepoFilesWithoutGitFollowSimpleGitignoreRules(t *testing.T) {
	r := newTestRepo(t)
	r.write(".gitignore", "# build output\nvendor\n*.log\n/dist\ndocs/build/\n!keep.log\n")
	r.write("vendor/README.md", "# V\n")
	r.write("sub/vendor/README.md", "# V\n")
	r.write("app.log", "x\n")
	r.write("dist/out.md", "# D\n")
	r.write("sub/dist/kept.md", "# D\n")
	r.write("docs/build/page.md", "# B\n")
	r.write("docs/guide.md", "# G\n")
	r.write("node_modules/x/README.md", "# N\n")

	files := relativeFiles(r)

	for _, ignored := range []string{"vendor/README.md", "sub/vendor/README.md", "app.log", "dist/out.md", "docs/build/page.md", "node_modules/x/README.md"} {
		isTrue(t, !slices.Contains(files, ignored), ignored+" is ignored")
	}
	for _, kept := range []string{"sub/dist/kept.md", "docs/guide.md", "README.md", ".gitignore"} {
		isTrue(t, slices.Contains(files, kept), kept+" is listed")
	}
}
