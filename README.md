# sisyphus

> **DO NOT USE.** This is a personal project, built for my own workflow with no
> guarantees of stability, support, or correctness. Expect breaking changes
> without notice. Use at your own risk.

A small CLI for managing the issues in `issues/` and managing work in a more useful way.

## Usage

Install it, then initialize a repo. Init sets up issue tracking only: `issues/TEMPLATE.md` and the
`open`/`in-progress`/`closed` directories. A contributing guide, agent rules, versioning, CI, and
GitHub Actions workflows are a separate concern of a project-scaffolding template, not of sisyphus.

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
sisyphus init
```

Create and move issues through `issues/open/`, `issues/in-progress/`, and `issues/closed/`:

```bash
sisyphus new <issue-name> --title "<title>" --priority high --tags "<component>"
sisyphus update <issue-name> in-progress --bookmark <bookmark> --owner <you>
sisyphus update <issue-name> closed --resolution completed   # or: abandoned
sisyphus show <issue-name>
```

Find issues:

```bash
sisyphus list --state open,in-progress --priority critical,high
sisyphus search "timeout" --section summary
```

Relate an issue to another issue, or to an external tracker:

```bash
sisyphus parent <issue-name> <parent-issue-name>
sisyphus depends-on <issue-name> <blocking-issue-name>
sisyphus remote <issue-name> <url>
```

View a large task's whole family tree in the terminal:

```bash
sisyphus graph <issue-name>
```

Check wikilinks:

```bash
sisyphus resolve "[[<issue-name>]]"
sisyphus links <file>
```

Run `sisyphus <command> --help` for all options on any command. See [[CONTRIBUTING]] for the full
workflow: the design log, decisions, versioning, the changelog, wikilinks, issues, and jj.

## MCP server

`sisyphus-mcp` gives an agent every `sisyphus` command above as an MCP tool, over stdio. It is a
thin wrapper: each tool shells out to the `sisyphus` binary, so it needs `sisyphus` installed and on
`PATH` too.

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest
```

Add it to an MCP-compatible client, for example Claude Code:

```bash
claude mcp add sisyphus -- sisyphus-mcp
```

Each tool (`sisyphus_new`, `sisyphus_update`, `sisyphus_show`, `sisyphus_list`, `sisyphus_search`,
`sisyphus_graph`, `sisyphus_parent`, `sisyphus_remote`, `sisyphus_depends_on`, `sisyphus_slug`,
`sisyphus_resolve`, `sisyphus_links`, `sisyphus_init`) takes an optional `dir` argument (the repo to
act on; defaults to `sisyphus-mcp`'s own working directory), plus the same arguments as the CLI
command it wraps.
