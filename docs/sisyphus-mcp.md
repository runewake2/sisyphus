# sisyphus-mcp

An MCP server that gives an agent every sisyphus command as a tool, over stdio.

## Install

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest
```

`sisyphus-mcp` is a thin wrapper: each tool runs the `sisyphus` binary and returns its output. So
`sisyphus` must be installed and on the `PATH` of the MCP client too, and the tools always behave
exactly like the commands. `sisyphus-mcp` reports the same version as `sisyphus --version`. When the `sisyphus` that it finds has
another version, every error result tells it (see [[#Troubleshooting]]).

## Set up a client

### Claude Code

```bash
claude mcp add sisyphus -- sisyphus-mcp                    # for you, in this directory
claude mcp add --scope project sisyphus -- sisyphus-mcp    # for everyone, in .mcp.json
```

### Other clients

Most MCP clients take a JSON configuration with a command for each stdio server:

```json
{
  "mcpServers": {
    "sisyphus": {
      "command": "sisyphus-mcp",
      "args": []
    }
  }
}
```

A client that starts servers from a desktop launcher often has a shorter `PATH` than your shell, so
it can find neither `sisyphus-mcp` nor `sisyphus`. Give the full path of `sisyphus-mcp` (for
example `/home/alice/go/bin/sisyphus-mcp`; `go env GOPATH` tells the directory), and add the
directory of `sisyphus` to the `PATH` of the server:

```json
{
  "mcpServers": {
    "sisyphus": {
      "command": "/home/alice/go/bin/sisyphus-mcp",
      "args": [],
      "env": { "PATH": "/home/alice/go/bin:/usr/local/bin:/usr/bin:/bin" }
    }
  }
}
```

Set `SISYPHUS_OWNER` in the `env` of the server to give each agent an owner name: an issue that the
agent starts gets it as its owner when the call has no `owner`, and `mine` of `sisyphus_list` and
`sisyphus_search` filters on it.

[[agents]] has instructions to paste into the `AGENTS.md` or `CLAUDE.md` of your repo.

A guide to running sisyphus with a local Ollama model is planned: [[ollama-task-manager-guide]].

### The repo

The tools act on the repo of the directory that the client starts `sisyphus-mcp` in, as `sisyphus`
does in a shell. Claude Code starts it in the project directory. Two flags of `sisyphus-mcp` change
this:

| Flag | Use |
|---|---|
| `--dir <path>` | The repo of every tool call: a directory in it. For a client that starts the server somewhere else. |
| `--allow-dir` | Give each tool a `dir` argument, so that an agent that works in more than one repo can name the repo in each call. |

```json
{ "command": "sisyphus-mcp", "args": ["--dir", "/home/alice/src/widgets"] }
```

Without `--allow-dir`, the tools have no `dir` argument. Small models filled the optional `dir`
with words from its description, for example `"current"` or `"docs"`, and the call failed. With
the argument gone, they cannot.

## Instructions

When a client connects, the server gives it these instructions, which the client shows to the
model:

```
sisyphus tracks the issues of this repo as Markdown files in issues/. Work on one issue at a time:

1. Find work: sisyphus_list with ready true and state open. sisyphus_claims shows the issues that other workspaces have started.
2. Read the issue with sisyphus_show.
3. Claim it with sisyphus_start and your owner name. If the start fails because another owner or workspace has the issue, select another issue. Do not use take unless a human tells you to.
4. Record progress with sisyphus_edit, section notes, and append true.
5. For work that you find but that is out of scope, create an issue with sisyphus_new and deferredFrom.
6. When the work is complete, use sisyphus_close with a message that tells what was done, or sisyphus_abandon with a message that tells why the work stopped. Then run sisyphus_lint.

Read the warnings at the end of each result. Lists, for example tags and dependsOn, are comma-separated strings.
```

With `--allow-dir`, they end with "Leave dir out unless the user names another repo." When the
`sisyphus` binary has another version than `sisyphus-mcp`, a warning line follows them.

## Tools

| Tool | Command | Default output |
|---|---|---|
| `sisyphus_init` | [[sisyphus-init]] | text |
| `sisyphus_new` | [[sisyphus-new]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_start` | [[sisyphus-start]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_close` | [[sisyphus-close]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_abandon` | [[sisyphus-abandon]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_update` | [[sisyphus-update]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_edit` | [[sisyphus-edit]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_show` | [[sisyphus-show]] | JSON (`text: true` for text) |
| `sisyphus_list` | [[sisyphus-list]] | JSON (`text: true` for the table) |
| `sisyphus_search` | [[sisyphus-search]] | JSON (`text: true` for text) |
| `sisyphus_graph` | [[sisyphus-graph]] | the drawing (`full: true` for every linked issue, `json: true` for JSON, `mermaid: true` for Mermaid source) |
| `sisyphus_claims` | [[sisyphus-claims]] | JSON |
| `sisyphus_parent` | [[sisyphus-parent]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_depends_on` | [[sisyphus-depends-on]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_remote` | [[sisyphus-remote]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_slug` | [[sisyphus-slug]] | text |
| `sisyphus_resolve` | [[sisyphus-resolve]] | text |
| `sisyphus_links` | [[sisyphus-links]] | JSON |
| `sisyphus_lint` | [[sisyphus-lint]] | JSON (problems are a result, not an error) |

`sisyphus_close` and `sisyphus_abandon` need `message`, the text of the Resolution section, so that
the issue tells what happened. `sisyphus_edit` always needs `section` and `text`, and an empty `section`
is an error: `edit` with only a name opens an editor, which is for people at a terminal, as is
[[sisyphus-completion]].

Each tool's arguments mirror the command's arguments and flags, in camelCase (for example
`dependsOn`, `deferredFrom`, `blockingIssue`). A list, such as `tags` or `dependsOn`, is one
comma-separated string, as on the command line. `metadata` is a JSON object. With `--allow-dir`, every tool
also takes the optional `dir` (see [[#The repo]]). The client shows the description of each argument.

## Results

- **Output**: the stdout of the command, as text.
- **Warnings**: the warnings that the command writes to stderr come after the output, in the same
  result, each on a line that starts with `Warning:`. The call succeeds, so the agent must read
  them.
- **Errors**: a command that fails gives a result marked as an error (`isError: true`), with the
  message of the command, for example `Error: Invalid priority 'urgent'. ...`. The message tells
  what to change, so the agent can correct the call.
- **Time limit**: each run of `sisyphus` stops after 60 seconds, with an error that tells the limit.
  A command takes much less; a longer run waits for something, for example a lock. Set
  `SISYPHUS_MCP_TIMEOUT` in the `env` of the server to change it, for example to `2m`.
- **Cancel**: when the client cancels a call, sisyphus-mcp stops the `sisyphus` run of the call. A
  change that the run made before it stopped stays.
- **Invalid arguments**: an argument of the wrong type, for example a JSON array for `dependsOn`, is
  an error from the MCP library, before sisyphus runs.

## An example session

The tool calls of an agent in a new repo, with each result:

```
sisyphus_init {}
→ issues/TEMPLATE.md
  issues/closed/.gitkeep
  issues/in-progress/.gitkeep
  issues/open/.gitkeep
  Next: create an issue with: sisyphus new <name> --title "<title>"

sisyphus_new {"title": "Fix the login bug", "tags": "bug", "priority": "high"}
→ issues/open/fix-login-bug.md

sisyphus_new {"title": "Split the login form", "dependsOn": "fix-login-bug"}
→ issues/open/split-login-form.md

sisyphus_list {"ready": true}
→ [
    {
      "name": "fix-login-bug",
      "title": "Fix the login bug",
      "state": "open",
      "priority": "high",
      "tags": [
        "bug"
      ]
    }
  ]

sisyphus_start {"name": "fix-login-bug", "owner": "agent"}
→ issues/in-progress/fix-login-bug.md

sisyphus_start {"name": "fix-login-bug", "owner": "other-agent"}
→ (isError) Error: Issue 'fix-login-bug' is in progress, owned by agent. Select another issue, or to take it from agent, add --take.

sisyphus_close {"name": "fix-login-bug", "message": "Fixed: the session cache now expires."}
→ issues/closed/fix-login-bug.md
  Warning: Issue 'fix-login-bug' is closing, but these issues still depend on it: split-login-form.

sisyphus_update {"name": "fix-login-bug", "priority": "urgent"}
→ (isError) Error: Invalid priority 'urgent'. Use one of: critical, high, medium, low. Did you mean 'critical'?
```

`split-login-form` is not in the `ready` list, because it depends on `fix-login-bug`, which is not
closed yet.

## Troubleshooting

| Result | Cause and fix |
|---|---|
| `sisyphus is not installed or not on PATH.` on every tool | The client finds `sisyphus-mcp` but not `sisyphus`. Install `sisyphus`, and add its directory to the `PATH` of the server (see [[#Other clients]]). |
| The client cannot start the server | The client does not find `sisyphus-mcp`. Give its full path as the `command`. |
| `Error: Cannot find the repo root.` | `sisyphus-mcp` runs outside the repo. Start it with `--dir` in the repo, start the client in the repo, or set up the repo with `sisyphus_init`. |
| The issues of another repo | The server started in another repo. Start it with `--dir`, or with `--allow-dir` and give a `dir` in each call. |
| `validating "arguments": ... dir` | The server has no `dir` argument. Start it with `--allow-dir`. |
| `validating "arguments": ...` | An argument has the wrong type. Lists are comma-separated strings. |
| `sisyphus-mcp X runs sisyphus Y from <path>` after an error | The two binaries have different versions, so a tool can pass a flag that the binary does not know. Install the same version of both, and restart the server. The server also writes this line to its log. |
| `sisyphus ... did not finish in 1m0s` | The command waited, for example for a lock of the git index that another process holds. Find that process, or raise `SISYPHUS_MCP_TIMEOUT`. |
| Old behavior after an update | The client still runs the old `sisyphus-mcp`. Restart the client, or the server in the client. |
