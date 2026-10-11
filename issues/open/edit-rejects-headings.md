---
title: "Stop edit from adding headings that split a section"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Stop edit from adding headings that split a section

## Summary

`sisyphus edit` accepts text that contains a heading at the level of the section or above, for example `## Notes`. The text then makes a new section, and a later `edit` of that name changes the wrong section.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new edit-test
sisyphus edit edit-test Summary -- $'text\n\n## Notes\n\nfake notes'
grep -c '^## Notes' issues/open/edit-test.md     # 2
sisyphus edit edit-test Notes "real"             # changes the first Notes: the one inside Summary
```

`editSection` in `cmd/sisyphus/issues.go` replaces the lines of the section with the text and does not check the
text. A Markdown heading in the text, for example from a pasted commit message or an agent's answer, splits the
section. The real `Notes` section then cannot be reached by name.

## Acceptance criteria

- [ ] `edit` fails when the text has a heading (outside code blocks) at the level of the section or a higher level, and the error names the heading.
- [ ] A deeper heading (for example `###` in a `##` section) stays allowed.
- [ ] Tests cover both cases, for replace and for `--append`.

## Out of scope

- Changing headings in the text automatically.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
