---
title: "Warn about invalid issue files in list, show, and search"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
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

# Warn about invalid issue files in list, show, and search

## Summary

`list`, `show`, and `search` read issue files that are not valid without any warning, and then show wrong data. The frontmatter reader in `cmd/sisyphus/document.go` is a line-based key/value reader, not a YAML parser, so it accepts frontmatter that YAML tools (and Obsidian) reject.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. Each case below exits 0 with no warning:

- **Invalid YAML:** a file with `title: [broken` is listed with the title `[broken`, and `show` prints empty priority and effort.
- **State does not match the directory:** a file in `issues/in-progress/` with `state: open` appears in `list --state in-progress` with the state `open`. Only `update` notices and corrects it.
- **Invalid enum value:** a hand-edited `priority: urgent` appears in `list`, but `list --priority high,critical` and every other priority filter hide it, so the issue can drop out of view.

[[lint-issues-command]] will find these problems when someone runs it. This issue is about the commands that people and agents run every day: they must not show bad data as if it were good.

## Acceptance criteria

- [ ] `list`, `show`, and `search` print one warning to stderr for each issue file that has frontmatter that does not parse, a state that does not match its directory, or an enum value that is not valid.
- [ ] The warnings do not change stdout, so `--json` output stays parseable.
- [ ] Invalid YAML is detected, for example by a real YAML parse of the frontmatter, or by the same checks that `lint` will use.
- [ ] Tests cover each of the three cases.

## Out of scope

- Automatic repair (see [[lint-issues-command]] and [[migrate-command]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
