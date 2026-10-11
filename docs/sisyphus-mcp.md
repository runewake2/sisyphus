# sisyphus-mcp

An MCP server that gives an agent every sisyphus command as a tool, over stdio.

## Install

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest
```

`sisyphus-mcp` is a thin wrapper: each tool runs the `sisyphus` binary and returns its output. So
`sisyphus` must be installed and on the `PATH` of the MCP client too, and the tools always behave
exactly like the commands. `sisyphus-mcp` reports the same version as `sisyphus --version`.

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

A guide to running sisyphus with a local Ollama model is planned: [[ollama-task-manager-guide]].

### The repo

The tools act on the repo of the directory that the client starts `sisyphus-mcp` in, as `sisyphus`
does in a shell. Claude Code starts it in the project directory. When a client starts it somewhere
else, or the agent works in more than one repo, give each tool call a `dir` argument: a directory
in the repo.

## Tools

| Tool | Command | Default output |
|---|---|---|
| `sisyphus_init` | [[sisyphus-init]] | text |
| `sisyphus_new` | [[sisyphus-new]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_update` | [[sisyphus-update]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_edit` | [[sisyphus-edit]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_show` | [[sisyphus-show]] | JSON (`text: true` for text) |
| `sisyphus_list` | [[sisyphus-list]] | JSON (`text: true` for the table) |
| `sisyphus_search` | [[sisyphus-search]] | JSON (`text: true` for text) |
| `sisyphus_graph` | [[sisyphus-graph]] | the drawing (`full: true` for every linked issue, `json: true` for JSON, `mermaid: true` for Mermaid source) |
| `sisyphus_parent` | [[sisyphus-parent]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_depends_on` | [[sisyphus-depends-on]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_remote` | [[sisyphus-remote]] | the path (`json: true` for the issue as JSON) |
| `sisyphus_slug` | [[sisyphus-slug]] | text |
| `sisyphus_resolve` | [[sisyphus-resolve]] | text |
| `sisyphus_links` | [[sisyphus-links]] | JSON |
| `sisyphus_lint` | [[sisyphus-lint]] | JSON (problems are a result, not an error) |

[[sisyphus-start]], [[sisyphus-close]], and [[sisyphus-abandon]] are short forms of `update`, so
they have no tools of their own: use `sisyphus_update` with `state` `in-progress` or `closed`,
`resolution`, and `message`. [[sisyphus-open]] and [[sisyphus-completion]] are for people at a
terminal, and have no tools.

Each tool's arguments mirror the command's arguments and flags, in camelCase (for example
`dependsOn`, `deferredFrom`, `blockingIssue`). A list, such as `tags` or `dependsOn`, is one
comma-separated string, as on the command line. `metadata` is a JSON object. Every tool also takes
the optional `dir` (see [[#The repo]]). The client shows the description of each argument.

## Results

- **Output**: the stdout of the command, as text.
- **Warnings**: the warnings that the command writes to stderr come after the output, in the same
  result, each on a line that starts with `Warning:`. The call succeeds, so the agent must read
  them.
- **Errors**: a command that fails gives a result marked as an error (`isError: true`), with the
  message of the command, for example `Error: Invalid priority 'urgent'. ...`. The message tells
  what to change, so the agent can correct the call.
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

sisyphus_update {"name": "fix-login-bug", "state": "in-progress", "owner": "agent"}
→ issues/in-progress/fix-login-bug.md

sisyphus_update {"name": "fix-login-bug", "state": "closed", "message": "Fixed: the session cache now expires."}
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
| `Error: Cannot find the repo root.` | `sisyphus-mcp` runs outside the repo. Give a `dir` in the repo, start the client in the repo, or set up the repo with `sisyphus_init`. |
| The issues of another repo | The server started in another repo. Give a `dir`. |
| `validating "arguments": ...` | An argument has the wrong type. Lists are comma-separated strings. |
| Old behavior after an update | The client still runs the old `sisyphus-mcp`. Restart the client, or the server in the client. |
