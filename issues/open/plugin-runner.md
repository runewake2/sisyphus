---
title: "Configure plugins per repo and run each as a child process"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [plugins]        # The components that the work touches, for example [widget-scheduler, plan]
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

# Configure plugins per repo and run each as a child process

## Summary

Let a repo configure its plugins, and run each one as a child process that sisyphus talks to over
[[plugin-protocol]].

## Context

Part of [[integration-plugins]]. Plugins are configured per repo only.

- **Configuration**: a file in the repo names each plugin and the command that starts it (a path,
  or a name on `PATH`, plus arguments). A plugin can come from anywhere; the file only says how to
  start it. Pick a format the standard library can read (for example JSON), so the core gains no
  dependency. Optionally name a default plugin for issues that have no `remote` yet.
- **Process**: one child process per sisyphus command. Handshake first, then requests. Close
  stdin to end the session. A timeout, a crash, or a protocol error is an error for that command
  only, and names the plugin and includes its stderr.
- **Environment**: the plugin inherits sisyphus's environment, so credentials (for example a
  token) reach the plugin without ever being written to the configuration file.
- **Routing**: an issue's `remote` URL selects the plugin, by the URL patterns from each plugin's
  handshake. A URL that no plugin owns, or that more than one plugin owns, is a clear error.

## Acceptance criteria

- [ ] The configuration file and its format are documented.
- [ ] The runner starts a plugin, runs the handshake, sends requests, and stops it cleanly.
- [ ] A timeout, a crash, and a malformed response each give a clear error naming the plugin.
- [ ] Routing by `remote` URL works, and an unowned or ambiguous URL is an error.
- [ ] A command lists the configured plugins and whether each one completes its handshake.
- [ ] Tested against a small fake plugin built by the tests.

## Out of scope

- The user-facing pull/push/sync commands ([[plugin-sync-commands]]).
- User-level configuration: plugins are configured per repo only.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
