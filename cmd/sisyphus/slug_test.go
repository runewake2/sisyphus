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
	equal(t, "fix-login-bug", strings.TrimSpace(res.output))
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
	r.mustRun("new", "fix-login-bug")

	res := r.run("slug", "Fix The Login Bug")

	equal(t, 0, res.exit)
	equal(t, "fix-login-bug-2", strings.TrimSpace(res.output))
}

func TestSlugSuffixDoesNotExceedSixWords(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "one-two-three-four-five-six")

	res := r.run("slug", "one two three four five six")

	equal(t, 0, res.exit)
	equal(t, "one-two-three-four-five-2", strings.TrimSpace(res.output))
}

func TestSlugSpellsLettersInASCII(t *testing.T) {
	cases := map[string]string{
		"Café crème: fix the naïve parser": "cafe-creme-fix-naive-parser",
		"Straße über Ærø":                  "strasse-uber-aero",
		"Łódź œuvre déjà vu":               "lodz-oeuvre-deja-vu",
		"ＦＵＬＬＷＩＤＴＨ ｔｅｘｔ":                   "fullwidth-text",
	}
	for text, want := range cases {
		t.Run(text, func(t *testing.T) {
			r := newTestRepo(t)

			res := r.mustRun("slug", text)

			equal(t, want, strings.TrimSpace(res.output))
		})
	}
}

func TestSlugOutputIsAlwaysASCII(t *testing.T) {
	r := newTestRepo(t)

	res := r.mustRun("slug", "Fix Привет 解析 parser 🐛 bug ™")

	equal(t, "fix-parser-bug-tm", strings.TrimSpace(res.output))
}

func TestSlugDropsFillerWordsBeforeItKeepsSixWords(t *testing.T) {
	r := newTestRepo(t)

	res := r.mustRun("slug", "Build the issue index once per command for a repo")

	equal(t, "build-issue-index-once-per-command", strings.TrimSpace(res.output))
}

func TestSlugKeepsFillerWordsWhenTooFewWordsRemain(t *testing.T) {
	r := newTestRepo(t)

	res := r.mustRun("slug", "To be or not")

	equal(t, "to-be-or-not", strings.TrimSpace(res.output))
}
