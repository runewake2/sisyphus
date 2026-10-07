---
title: "Write the GitHub plugin"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [plugins, github] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[integration-plugins]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[plugin-protocol]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Write the GitHub plugin

## Summary

Write the first plugin: a GitHub plugin, in this repo but outside the sisyphus core, that speaks
[[plugin-protocol]] and translates between a sisyphus issue and a GitHub issue.

## Context

Part of [[integration-plugins]]. It replaces the GitHub logic that lived in the removed workflows
([[github-action-issue-to-pr]], [[github-action-sync-issues-to-github]]), so port their behavior:

- Owns `remote` URLs of the form `https://github.com/<owner>/<repo>/issues/<number>`.
- **push**: title from the issue title; body from the issue body, with a note at the top that the
  issue is generated from sisyphus and edits are overwritten; open or closed from the state, with
  the close reason `completed` or `not planned` from the resolution.
- **pull**: title and body into a new sisyphus issue (the body goes in the Context section, as
  `sisyphus new --context` does).
- **list**: open issues of the repo, for intake.
- **Auth**: a token from the plugin's environment (for example `GITHUB_TOKEN`), never from the
  configuration file.

Keep it a separate Go module (or at least a separate binary that the core never imports), so the
core gains no GitHub dependency. Prefer the GitHub REST API over the standard library's HTTP
client to shelling out to `gh`.

## Acceptance criteria

- [ ] A `sisyphus-github` plugin binary speaks the protocol, including the handshake.
- [ ] `push`, `pull`, and `list` behave as described above.
- [ ] The core module does not import it or any GitHub library.
- [ ] Tested against a fake GitHub API server, with no network access.
- [ ] The README says how to configure it in a repo.

## Out of scope

- The workflows that run it ([[github-intake-workflow]], [[github-sync-workflow]]).
- Jira and GitLab plugins: separate issues later.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
