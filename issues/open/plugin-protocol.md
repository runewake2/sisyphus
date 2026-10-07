---
title: "Define the sisyphus plugin protocol"
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
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Define the sisyphus plugin protocol

## Summary

Define the protocol sisyphus uses to talk to an integration plugin: a small, versioned,
request/response protocol over the plugin's stdin and stdout. It is our own protocol, not MCP:
sisyphus must not depend on MCP.

## Context

Part of [[integration-plugins]]. Proposed shape, to settle in this issue:

- **Framing**: one JSON object per line. A request is `{"id", "method", "params"}`; a response is
  `{"id", "result"}` or `{"id", "error": {"code", "message"}}`. One response per request.
  Anything the plugin writes to stderr is diagnostics, passed through to the user.
- **`handshake`**: the first request. The plugin returns its name, the protocol version it speaks,
  the URL patterns of the `remote` values it owns, and the operations it supports. sisyphus
  refuses a plugin with an incompatible version, with a clear message.
- **`pull`**: a `remote` URL in, a platform-neutral issue document out.
- **`push`**: an issue document (and its current `remote`, if any) in, the `remote` URL out. It
  creates the remote issue or updates it.
- **`list`**: filters in, a list of remote issues (`remote` URL plus summary fields) out, for intake.
- **Issue document**: the same platform-neutral fields `sisyphus show --json` prints (name,
  title, state, resolution, priority, tags, body, remote, metadata, ...). The plugin translates it
  to and from its platform; sisyphus never sees a platform's own fields.

Write the spec down in this repo, and add the message types to a small Go package with no
dependencies, so a plugin written in Go can import them. A plugin in any other language follows
the spec.

## Acceptance criteria

- [ ] The protocol is specified in a document in this repo.
- [ ] Go types for every message, in a package with no dependencies beyond the standard library.
- [ ] The handshake carries a protocol version, and the version rules are written down.
- [ ] Errors have a code and a message, and the codes are listed.

## Out of scope

- Starting and talking to a plugin process ([[plugin-runner]]).
- The commands that use plugins ([[plugin-sync-commands]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
