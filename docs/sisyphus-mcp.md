# sisyphus-mcp

An MCP server that gives an agent every sisyphus command as a tool, over stdio.

## Install

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest
claude mcp add sisyphus -- sisyphus-mcp      # or the equivalent for any MCP client
```

`sisyphus-mcp` is a thin wrapper: each tool runs the `sisyphus` binary and returns its output. So
`sisyphus` must be installed and on `PATH` too, and the tools always behave exactly like the
commands. If `sisyphus` is missing, every tool returns an error that says how to install it.

## Tools

| Tool | Command | Default output |
|---|---|---|
| `sisyphus_init` | [[sisyphus-init]] | text |
| `sisyphus_new` | [[sisyphus-new]] | text |
| `sisyphus_update` | [[sisyphus-update]] | text |
| `sisyphus_show` | [[sisyphus-show]] | JSON (`text: true` for text) |
| `sisyphus_list` | [[sisyphus-list]] | JSON (`text: true` for the table) |
| `sisyphus_search` | [[sisyphus-search]] | JSON (`text: true` for text) |
| `sisyphus_graph` | [[sisyphus-graph]] | the drawing (`full: true` for every linked issue, `json: true` for JSON, `mermaid: true` for Mermaid source) |
| `sisyphus_parent` | [[sisyphus-parent]] | text |
| `sisyphus_depends_on` | [[sisyphus-depends-on]] | text |
| `sisyphus_remote` | [[sisyphus-remote]] | text |
| `sisyphus_slug` | [[sisyphus-slug]] | text |
| `sisyphus_resolve` | [[sisyphus-resolve]] | text |
| `sisyphus_links` | [[sisyphus-links]] | JSON |

Each tool's arguments mirror the command's flags, in camelCase (for example `dependsOn`,
`deferredFrom`, `blockingIssue`), and `metadata` is a JSON object. Every tool also takes an optional
`dir`: the repo to act on, which defaults to the directory `sisyphus-mcp` was started in.

A command's error becomes a tool result marked as an error, carrying the same message the command
prints, so the agent can read it and correct itself.

`sisyphus-mcp` reports the same version as `sisyphus --version`.
