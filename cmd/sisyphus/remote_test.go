package main

import (
	"slices"
	"strings"
	"testing"
)

func TestRemoteNewIssuesHaveNoRemoteByDefault(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "remote-issue-test")

	equal(t, "", r.frontmatter("issues/open/remote-issue-test.md").get("remote"))
}

func TestNewSetsTheRemoteAsAQuotedURL(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "remote-issue-test", "--remote", "https://github.com/acme/widgets/issues/42")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	doc := r.frontmatter("issues/open/remote-issue-test.md")
	equal(t, "https://github.com/acme/widgets/issues/42", doc.get("remote"))
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool {
		return strings.HasPrefix(l, `remote: "https://github.com/acme/widgets/issues/42"`)
	}), "the remote is quoted")
}

func TestNewRejectsARemoteThatIsNotAURL(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "remote-issue-test", "--remote", "not-a-url")

	equal(t, 1, res.exit)
	contains(t, res.error, "'not-a-url' is not a URL")
	isTrue(t, !r.exists("issues/open/remote-issue-test.md"), "no issue is created")
}

func TestRemoteSetsAndClearsTheRemote(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "remote-issue-test", "--state", "in-progress", "--bookmark", "ai/work")

	set := r.run("remote", "remote-issue-test", "https://github.com/acme/widgets/issues/42")

	equal(t, 0, set.exit)
	equal(t, "issues/in-progress/remote-issue-test.md", strings.TrimSpace(set.output))
	doc := r.frontmatter("issues/in-progress/remote-issue-test.md")
	equal(t, "https://github.com/acme/widgets/issues/42", doc.get("remote"))
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, "remote: ") && strings.Contains(l, " # ") }), "the remote line keeps its comment")

	clear := r.run("remote", "remote-issue-test", "--clear")

	equal(t, 0, clear.exit)
	equal(t, "", r.frontmatter("issues/in-progress/remote-issue-test.md").get("remote"))
}

func TestRemoteRejectsBothAURLAndClear(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "remote-issue-test")

	res := r.run("remote", "remote-issue-test", "https://github.com/acme/widgets/issues/42", "--clear")

	equal(t, 1, res.exit)
	contains(t, res.error, "Give a remote URL or --clear, but not both")
}

func TestRemoteRequiresAURLOrClear(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "remote-issue-test")

	res := r.run("remote", "remote-issue-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "Give a remote URL or --clear, but not both")
}

func TestRemoteRejectsAnInvalidURL(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "remote-issue-test")

	res := r.run("remote", "remote-issue-test", "not-a-url")

	equal(t, 1, res.exit)
	contains(t, res.error, "'not-a-url' is not a URL")
}

func TestRemoteReportsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("remote", "no-such-issue", "https://github.com/acme/widgets/issues/42")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'no-such-issue'")
}
