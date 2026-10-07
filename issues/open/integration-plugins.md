---
title: "Talk to GitHub, Jira, and GitLab only through configured plugin processes"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: x-large        # small | medium | large | x-large
tags: [plugins, integration] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
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

# Talk to GitHub, Jira, and GitLab only through configured plugin processes

## Summary

sisyphus talks to an external tracker (GitHub, Jira, GitLab, ...) only through a configured plugin:
a separate executable that sisyphus starts as a child process, one per integration. The plugin owns
everything platform-specific, including translating between a sisyphus issue and the platform's
issue. sisyphus itself knows only the plugin protocol, never a platform.

## Context

From the human, 2026-10-07: "When sisyphus is used to interact with a github/jira/gitlab ticket
this would go through a configured plugin. It would use a separate child process to enable each
integration and translating between a sisyphus issue and platform specific issue."

This follows the rule that sisyphus itself must not know about any particular repo or issue: a
tracker's API, auth, field mapping, and URL format are exactly that kind of knowledge. Today the
GitHub integration breaks the rule from outside the core: `.github/workflows/issue-to-pr.yml` and
`.github/workflows/sync-to-github.yml` ([[github-action-issue-to-pr]],
[[github-action-sync-issues-to-github]]) call `gh` directly, so the translation lives in shell
scripts. Under this design those workflows shrink to calling sisyphus, and a GitHub plugin does
the work.

Proposed design, to confirm before it is split into sub-issues:

- **Configuration**: a per-repo file (for example `issues/plugins.toml`) names each plugin and the
  command that starts it. sisyphus finds plugins only through this file: no built-in list.
- **Process model**: sisyphus starts the plugin as a child process for the length of one command
  and speaks a request/response protocol over its stdin and stdout. A crash or a hang in a plugin
  is an error for that command, never for sisyphus as a whole.
- **Protocol**: a handshake (name, protocol version, which `remote` URLs the plugin owns, which
  operations it supports), then operations on a platform-neutral issue document: `pull` (remote
  issue -> sisyphus fields), `push` (sisyphus issue -> create or update the remote, returns its
  `remote` URL), and `list` (remote issues, for intake). Credentials stay inside the plugin.
- **Routing**: an issue's `remote` URL picks the plugin, by the URL patterns each plugin claims in
  its handshake. `remote` stays a plain URL, as it is today ([[pin-issue-to-remote-reference]]).
- **CLI**: for example `sisyphus pull <remote-url>` (create or refresh a sisyphus issue from a
  remote one), `sisyphus push <name>` (mirror one issue), and `sisyphus sync` (push every issue).
  The source-of-truth rule from [[github-source-of-truth]] stays: a push overwrites the remote,
  and nothing is read back into a sisyphus issue except through an explicit `pull`.
- **Where plugins live**: outside the sisyphus core module (for example `sisyphus-github`,
  `sisyphus-jira`, `sisyphus-gitlab` binaries), so the core never imports a platform SDK.

Open questions:

- **P1**: Reuse MCP as the plugin protocol (each plugin is an MCP server exposing `pull`/`push`/
  `list` tools, and sisyphus is its client), or define a smaller JSON-RPC protocol of our own? MCP
  gives a handshake, schemas, and an SDK sisyphus already depends on ([[sisyphus-mcp-server]]);
  a custom protocol is smaller and fully under our control.
- **P2**: One config file per repo, a user-level file, or both (per-repo plugins plus per-user
  credentials)?
- **P3**: Do plugins live in this repo (as separate modules) or in their own repos?
- **P4**: What happens to the two GitHub workflows: rewrite them on top of the GitHub plugin, or
  remove them and leave CI wiring to the project-scaffolding template?

## Acceptance criteria

- [ ] The open questions above are answered, and the answers recorded.
- [ ] The design is split into sub-issues: protocol, configuration and process runner, CLI
      commands, and one issue per plugin (GitHub first, porting the two workflows).
- [ ] sisyphus core contains no platform-specific code, names, or dependencies.

## Out of scope

- Implementing any plugin before the protocol is settled.
- Two-way sync. sisyphus stays the source of truth ([[github-source-of-truth]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
