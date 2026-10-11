---
title: "Document the issue file, its fields, the lifecycle, and the template"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [docs]           # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Document the issue file, its fields, the lifecycle, and the template

## Summary

There is no user documentation of the issue file itself: the frontmatter fields and what each means, the states and the lifecycle, the body sections, and what a repo may change in `issues/TEMPLATE.md`. Add a page for it.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. Today the field meanings are in the comments of the template and in the Frontmatter table of this
repo's CONTRIBUTING, which also has rules that only this repo uses. A user who wants to know what `approver`,
`deferred-from`, or `metadata` is for, which fields sisyphus requires, or whether they can add a section or a field to
the template, must read the source. The page can also say which rules sisyphus enforces (state matches the
directory, names, enum values, no reopen without `--reopen`) and which are conventions.

## Acceptance criteria

- [ ] A page in `docs/` shows an example issue file, and describes every frontmatter field, the states, the resolutions, the body sections, and the lifecycle.
- [ ] It says which fields and sections sisyphus needs, and which a repo can change or add in its template.
- [ ] README and [[commands]] link to it.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
