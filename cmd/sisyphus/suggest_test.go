package main

import "testing"

func TestAnUnknownNameSuggestsIssuesThatShareWords(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "update-docs-site", "--title", "Update the docs site theme")
	r.mustRun("new", "fix-login-bug", "--title", "Fix the login bug")

	res := r.run("show", "docs-site-theme")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'docs-site-theme'")
	contains(t, res.error, "Did you mean: update-docs-site (Update the docs site theme)?")
	notContains(t, res.error, "fix-login-bug")
}

func TestSuggestionsMatchNearWords(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "output-conventions-doc", "--title", "Document the output")

	res := r.run("start", "output-convention-docs")

	contains(t, res.error, "Did you mean: output-conventions-doc")
}

func TestSuggestionsAreAtMostThreeAndPutOpenIssuesFirst(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "login-bug-one", "--state", "closed")
	r.mustRun("new", "login-bug-two")
	r.mustRun("new", "login-bug-three")
	r.mustRun("new", "login-bug-four")

	res := r.run("show", "login-bug")

	contains(t, res.error, "Did you mean: login-bug-four (Login bug four), login-bug-three (Login bug three), login-bug-two (Login bug two)?")
	notContains(t, res.error, "login-bug-one")
}

func TestAClosedSuggestionSaysThatItIsClosed(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff", "--state", "closed")

	res := r.run("show", "backoff-retry")

	contains(t, res.error, "retry-backoff (closed: Retry backoff)")
}

func TestAnUnknownNameWithNoSharedWordSuggestsNothing(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	res := r.run("show", "dark-mode")

	equal(t, 1, res.exit)
	notContains(t, res.error, "Did you mean")
}
