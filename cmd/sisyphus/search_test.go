package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func setUpSearchIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "widget-scheduler-test", "--priority", "high", "--tags", "scheduler")
	r.write("issues/open/widget-scheduler-test.md", strings.Replace(
		r.read("issues/open/widget-scheduler-test.md"),
		"<One or two sentences. Tell what must change and why.>",
		"Make the widget scheduler faster under load.",
		1,
	))
	r.mustRun("new", "other-issue-test", "--priority", "low")
	r.write("issues/open/other-issue-test.md", strings.Replace(
		r.read("issues/open/other-issue-test.md"),
		"<One or two sentences. Tell what must change and why.>",
		"This is about something else entirely.",
		1,
	))
	return r
}

func TestSearchMatchesTheQueryCaseInsensitivelyInTitleAndBody(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "WIDGET SCHEDULER")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "widget-scheduler-test")
	isTrue(t, !strings.Contains(res.output, "other-issue-test"), "the other issue does not match")
}

func TestSearchMatchesTheTitleEvenWhenTheBodyDoesNot(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "SCHEDULER TEST") // substring of the auto-generated title, not of the body

	equal(t, 0, res.exit)
	contains(t, res.output, "widget-scheduler-test")
}

func TestSearchWithNoQueryAppliesOnlyFrontmatterFilters(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "--priority", "low")

	equal(t, 0, res.exit)
	contains(t, res.output, "other-issue-test")
	isTrue(t, !strings.Contains(res.output, "widget-scheduler-test"), "the high-priority issue is excluded")
}

func TestSearchCombinesQueryAndFrontmatterFilters(t *testing.T) {
	r := setUpSearchIssues(t)

	matches := r.run("search", "scheduler", "--tags", "scheduler")
	equal(t, 0, matches.exit)
	contains(t, matches.output, "widget-scheduler-test")

	noMatches := r.run("search", "scheduler", "--tags", "nonexistent-tag")
	equal(t, 0, noMatches.exit)
	equal(t, "", strings.TrimSpace(noMatches.output))
}

func TestSearchSectionRestrictsTheQueryToThatSection(t *testing.T) {
	r := setUpSearchIssues(t)

	inSummary := r.run("search", "widget scheduler", "--section", "summary")
	equal(t, 0, inSummary.exit)
	contains(t, inSummary.output, "widget-scheduler-test")

	inAcceptanceCriteria := r.run("search", "widget scheduler", "--section", "acceptance criteria")
	equal(t, 0, inAcceptanceCriteria.exit)
	equal(t, "", strings.TrimSpace(inAcceptanceCriteria.output))
}

func TestSearchSectionRestrictsTheReturnedContent(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "--section", "summary", "--json", "--tags", "scheduler")

	equal(t, 0, res.exit)
	var rows []searchRow
	if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	equal(t, 1, len(rows))
	equal(t, "Make the widget scheduler faster under load.", rows[0].Content)
	equal(t, "summary", rows[0].Section)
	isTrue(t, !strings.Contains(rows[0].Content, "## Context"), "only the Summary section is returned")
}

func TestSearchReturnsNothingForAQueryThatMatchesNoIssue(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "no-such-text-anywhere")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "", strings.TrimSpace(res.output))
}

func TestSearchRejectsAnInvalidStateOrPriority(t *testing.T) {
	cases := []struct{ flag, value string }{
		{"--state", "done"},
		{"--priority", "urgent"},
	}
	for _, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			r := setUpSearchIssues(t)

			res := r.run("search", c.flag, c.value)

			equal(t, 1, res.exit)
		})
	}
}

func TestSearchFailsForASectionThatNoIssueAndNoTemplateHas(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "--section", "summry", "widget")

	equal(t, 1, res.exit)
	contains(t, res.error, "named 'summry'")
	contains(t, res.error, "Summary, Context, Acceptance criteria")
}

func TestSearchAcceptsATemplateSectionInAnyCase(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.run("search", "--section", "SUMMARY", "faster")

	equal(t, 0, res.exit)
	contains(t, res.output, "widget-scheduler-test")
}

func TestSearchAcceptsASectionThatOnlyAnIssueHas(t *testing.T) {
	r := setUpSearchIssues(t)
	r.write("issues/open/other-issue-test.md", r.read("issues/open/other-issue-test.md")+"\n## Design\n\nA custom section.\n")

	res := r.run("search", "--section", "design", "custom")

	equal(t, 0, res.exit)
	contains(t, res.output, "other-issue-test")
}

func TestSearchReadyFiltersLikeList(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.mustRun("new", "second-issue-test", "--depends-on", "first-issue-test")

	res := r.mustRun("search", "--ready", "issue")

	contains(t, res.output, "first-issue-test")
	isTrue(t, !strings.Contains(res.output, "second-issue-test"), "a blocked issue is not ready")
}

func TestSearchReturnsTheMatchedLinesWithTheirSections(t *testing.T) {
	r := setUpSearchIssues(t)

	var rows []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "faster", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	equal(t, 1, len(rows))
	equal(t, "", rows[0].Content)
	equal(t, 1, len(rows[0].Matches))
	equal(t, "Summary", rows[0].Matches[0].Section)
	equal(t, "Make the widget scheduler faster under load.", rows[0].Matches[0].Text)
}

func TestSearchTextShowsTheMatchedLines(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.mustRun("search", "faster")

	contains(t, res.output, "widget-scheduler-test: ")
	contains(t, res.output, "  Summary: Make the widget scheduler faster under load.\n")
	notContains(t, res.output, "## Context")
}

func TestSearchDoesNotMatchTemplateText(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.mustRun("search", "changelog version") // only in the placeholder of Resolution

	equal(t, "", strings.TrimSpace(res.output))
}

func TestSearchContentReturnsTheBodyWithoutTemplateText(t *testing.T) {
	r := setUpSearchIssues(t)

	var rows []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "faster", "--content", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	contains(t, rows[0].Content, "## Summary\n\nMake the widget scheduler faster under load.")
	notContains(t, rows[0].Content, "## Context")
	equal(t, 0, len(rows[0].Matches))
}

func TestSearchCutsLongLinesAroundTheMatch(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "long-line-test", "--context", strings.Repeat("filler ", 100)+"needle "+strings.Repeat("filler ", 100))

	var rows []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "needle", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	text := rows[0].Matches[0].Text
	contains(t, text, "needle")
	isTrue(t, strings.HasPrefix(text, "…") && strings.HasSuffix(text, "…"), "the cut ends are marked: "+text)
	isTrue(t, len([]rune(text)) <= matchWidth+2, "the line is cut")
}

func TestSearchCountsTheMatchedLinesAfterTheFirstFive(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "many-lines-test", "--context", strings.Repeat("needle\n\n", 8))

	var rows []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "needle", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	equal(t, maxMatches, len(rows[0].Matches))
	equal(t, 3, rows[0].MoreMatches)
}

func TestListAndSearchLimitTheRowsAndTellTheRest(t *testing.T) {
	r := setUpSearchIssues(t)
	r.mustRun("new", "third-issue-test")

	list := r.mustRun("list", "--limit", "2", "--json")
	search := r.mustRun("search", "--limit", "1")

	var rows []listRow
	if err := json.Unmarshal([]byte(list.output), &rows); err != nil {
		t.Fatal(err)
	}
	equal(t, 2, len(rows))
	contains(t, list.error, "Showing 2 of 3 issues.")
	contains(t, search.error, "Showing 1 of 3 issues.")
}

func TestSearchMatchesAllWordsInAnyOrder(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "update-docs-site", "--title", "Update the docs site theme")
	r.mustRun("new", "docs-site-search", "--title", "Add search to the docs site")

	res := r.mustRun("search", "docs site theme update")

	contains(t, res.output, "update-docs-site")
	notContains(t, res.output, "docs-site-search")
}

func TestSearchWordsCanBeInTheTitleAndTheBody(t *testing.T) {
	r := setUpSearchIssues(t)

	res := r.mustRun("search", "load test") // "test" is only in the title, "load" only in the body

	contains(t, res.output, "widget-scheduler-test")
}

func TestSearchQuotesMatchAnExactPhrase(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "update-docs-site", "--title", "Update the docs site theme")
	r.mustRun("new", "theme-site-docs", "--title", "Theme the site docs")

	phrase := r.mustRun("search", `"docs site"`)
	mixed := r.mustRun("search", `theme "the docs"`)

	contains(t, phrase.output, "update-docs-site")
	notContains(t, phrase.output, "theme-site-docs")
	contains(t, mixed.output, "update-docs-site")
	notContains(t, mixed.output, "theme-site-docs")
}

func TestSearchShowsTheLinesOfEachWord(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff", "--title", "Add retry backoff", "--context", "The client retries at once.\n\nA cache hides the error.")

	var rows []searchRow
	if err := json.Unmarshal([]byte(r.mustRun("search", "cache retries", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	equal(t, 2, len(rows[0].Matches))
	equal(t, "The client retries at once.", rows[0].Matches[0].Text)
	equal(t, "A cache hides the error.", rows[0].Matches[1].Text)
}

func TestQueryTerms(t *testing.T) {
	equalSlices(t, []string{"login", "session cache", "bug"}, queryTerms(`Login "Session  Cache" bug`))
	equalSlices(t, []string{"open", "quote"}, queryTerms(`open "quote`))
	equal(t, 0, len(queryTerms(`  "" `)))
}
