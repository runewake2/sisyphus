---
title: "Add \"sisyphus search\" to search issues by frontmatter filters and content, with partial (section-scoped) results"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/search-command # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-search-command] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
agent-session:          # AI agent session id of the current agent working on the issue, if available.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Add "sisyphus search" to search issues by frontmatter filters and content, with partial (section-scoped) results

## Summary

Add `sisyphus search` to find issues by combining frontmatter filters (state, priority, tags,
owner, parent) with a free-text query over an issue's title and body. The query, and the content
returned for each match, can both be scoped to one section (for example `## Summary`) instead of
the whole issue.

## Context

Part of [[issue-manager-gaps]]. [[list-issues-command]] already covers pure frontmatter filtering
and explicitly deferred free-text search as "a different, simpler feature" — this issue is that
feature, extended with section scoping. Share the frontmatter-filter logic with `list` rather than
duplicating it; whichever of the two lands first should expose that matching as a function the
other can call.

Section scoping needs a way to slice a document's body into named sections by heading. There is no
such helper yet, but the building blocks exist: `headingPattern` and `outsideCodeFences`
(`cmd/sisyphus/links.go`) already find heading lines outside fenced code, and `hasHeading` already
matches a heading name case-insensitively (the same way `[[doc#Heading]]` link resolution does). A
new helper can walk headings at the top level (`##`) to find each section's start and end line.

Suggested shape:

```bash
sisyphus search "scheduler"                         # title + body of every issue, case-insensitive
sisyphus search "scheduler" --section summary       # search only within each issue's "## Summary"
sisyphus search --state open --priority high        # frontmatter filters only, no text query
sisyphus search "scheduler" --state open --tags cli # text query AND frontmatter filters
sisyphus search "scheduler" --section summary --json
```

The query is optional: with none given, `search` behaves like `list` restricted by whatever
frontmatter filters are given. `--section` does two things at once: it restricts matching to that
section's text, and it restricts the content returned for each match to that section, instead of
the whole body.

## Acceptance criteria

- [ ] A text query matches case-insensitively against an issue's title and body.
- [ ] The frontmatter filters [[list-issues-command]] defines (`--state`, `--priority`, `--tags`,
      `--owner`, `--parent`) also apply here, and combine with the text query and with each other
      using AND semantics.
- [ ] `--section <heading>` restricts the text query to that section only, matching the heading
      name case-insensitively (consistent with how `sisyphus links` matches `#heading` targets).
- [ ] With `--section`, each result's returned content is only that section's text, not the whole
      body.
- [ ] Without `--section`, each result returns the whole body (or a documented alternative such as a
      matching snippet with surrounding context).
- [ ] The text query is optional; frontmatter filters alone still work, equivalent to `list` with
      the same flags.
- [ ] A query or filter combination that matches nothing returns an empty result, not an error.
- [ ] `--json` prints machine-readable results, consistent with `list`/`show`/`links`'s `--json`.
- [ ] Covered by tests: text-only, frontmatter-only, combined, `--section` (both the match-scoping
      and the content-scoping), case-insensitivity, and the no-match case.

## Out of scope

- A query language with boolean operators (AND/OR/NOT) beyond combining separate flags; revisit
  only if plain substring matching plus filters turns out not to be enough.
- Fuzzy or relevance-ranked matching; this is case-insensitive substring matching, not scoring.
- Searching Markdown files outside `issues/` (design docs, changelog); scope matches
  [[list-issues-command]].

## Notes

- 2026-10-07: Implemented as planned. Refactored `list.go` to expose `matchingIssues` (frontmatter
  filtering) and `compareByStatePriorityName` (the shared sort), so `search` reuses both instead of
  duplicating them. Added `sectionIn` to `links.go`, next to the existing `headingPattern`/
  `outsideCodeFences`/heading-matching machinery it builds on. `--section` restricts the query
  haystack to just that section's text (no title) when given; without it, the query matches title +
  body, per the issue's "restricts the text query to that section only" wording.

## Resolution

Completed: `sisyphus search` combines `list`'s frontmatter filters with a case-insensitive text
query over title and body, optionally scoped (both the query and the returned content) to one
section. See [[0.0.6]].
