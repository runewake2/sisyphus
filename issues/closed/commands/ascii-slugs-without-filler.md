---
title: "Make slug output ASCII and drop filler words"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/slug-ascii # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-slug-ascii] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "5f1f06e4-afca-4277-8f3c-b8c818789ed6"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Make slug output ASCII and drop filler words

## Summary

Make sisyphus slug spell letters in ASCII and drop filler words before it keeps the first 6 words.

## Context

sisyphus slug kept runs of ASCII letters and digits and dropped every other character, so 'Café crème: fix the naïve parser' gave caf-cr-me-fix-the-na. It kept the first 6 words, so 'Build the issue index once per command' gave build-the-issue-index-once-per. Decided 2026-10-08 by the human: spell letters in ASCII (é to e), drop filler words before the 6-word cut, and always output ASCII.

## Acceptance criteria

- [x] Accented letters become their base letter (NFKD, then marks removed), and ß, æ, œ, ø, ł, đ, ð, þ, ı, ħ, ŋ, and ŧ get an ASCII spelling.
- [x] The output is always ASCII; a character with no ASCII form separates words.
- [x] Filler words drop before the 6-word cut, unless fewer than 2 words would remain.
- [x] The help, the MCP description, and [[sisyphus-slug]] state the rules.

## Out of scope

Transliteration of non-Latin scripts, such as Cyrillic or CJK. Their letters separate words.

## Notes

2026-10-08, from the human: adopt both changes; a slug must output ASCII.

## Resolution

Completed in [[0.0.20]].
