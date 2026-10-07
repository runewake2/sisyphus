package main

import (
	"strings"
	"testing"
)

func TestSlugMakesAKebabCaseNameFromText(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("slug", "Fix the login bug!")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "fix-the-login-bug", strings.TrimSpace(res.output))
}

func TestSlugTrimsToSixWords(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("slug", "one two three four five six seven eight")

	equal(t, 0, res.exit)
	equal(t, "one-two-three-four-five-six", strings.TrimSpace(res.output))
}

func TestSlugRejectsTextWithFewerThanTwoWords(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("slug", "solo")

	equal(t, 1, res.exit)
	contains(t, res.error, "does not have enough words")
}

func TestSlugAddsASuffixWhenTheNameAlreadyExists(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-the-login-bug")

	res := r.run("slug", "Fix The Login Bug")

	equal(t, 0, res.exit)
	equal(t, "fix-the-login-bug-2", strings.TrimSpace(res.output))
}

func TestSlugSuffixDoesNotExceedSixWords(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "one-two-three-four-five-six")

	res := r.run("slug", "one two three four five six")

	equal(t, 0, res.exit)
	equal(t, "one-two-three-four-five-2", strings.TrimSpace(res.output))
}
