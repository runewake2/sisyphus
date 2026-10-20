---
title: "Make the on-failure role get the failed notices, or correct the role files"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: []               # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
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

# Make the on-failure role get the failed notices, or correct the role files

## Summary

The six dependency-check roles declare on-failure: security-researcher, and their README says: a failure goes to the security researcher (hephaestus/roles/dependency-check/README.md, line 38). The design says the same for the bookmark-ready path (design/hephaestus/agent-delegation.md, step 3 of The government). The code does not do it on that path: for a notice (a broadcast, for example bookmark-ready), a failed check always tags the issue needs-human (internal/hephaestus/toolagent/toolagent.go, noticeResult, lines 504-512). Only a direct task message uses on-failure (stop, lines 536-545, and sendNext, line 570). The dependency-check roles run mainly on bookmark-ready notices (their listens), so their on-failure role gets no work on the main path. Found by the operations auditor, 2026-10-20. Decide one way: send the failed notice of a role with an AI on-failure to that role (a message to the role, with the output, without a change of the issue state), or change the dependency-check README and role files to say that a failed notice goes to a human, and that security-researcher gets only failed task messages.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] A human decides: the notice path uses on-failure, or the docs say that it does not.
- [ ] If the code changes: a failed notice of a tool role with an AI on-failure sends a message to that role with the output, and a test shows it.
- [ ] If the docs change: hephaestus/roles/dependency-check/README.md and design/hephaestus/agent-delegation.md say that a failed notice tags needs-human, and that on-failure applies to task messages only.
- [ ] hephaestus roles reports no problem for the dependency-check family.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: the human decision can happen now; no dependency is set. But if the decision is the code option, the change is in toolagent.go (noticeResult), where hephaestus evict is in work: make that change after [[hephaestus/core/evict-command]] lands, and after [[hephaestus/core/share-platform-shell]] if both are open, to avoid churn in one file.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
