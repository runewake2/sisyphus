# sisyphus

> **DO NOT USE.** This is a personal project and it's very vibes based atm. That may change later but for now this is primarily a project built just for me.

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

Draw an issue, everything below it, and the path above it in the terminal. On an epic, that is the
whole epic:

```bash
sisyphus graph <issue-name>
sisyphus graph <issue-name> --full
```

Check wikilinks:

```bash
sisyphus resolve "[[<issue-name>]]"
sisyphus links <file>
```

Run `sisyphus <command> --help` for all options on any command, and see [[commands]] (`docs/`)
for every command in detail. See [[CONTRIBUTING]] for the full
workflow: the design log, decisions, versioning, the changelog, wikilinks, issues, and jj.

## Obsidian

sisyphus issues are plain Markdown with YAML frontmatter, so a repo that uses sisyphus also works
as an [Obsidian](https://obsidian.md/) vault: open the repo's root as a vault to browse, link, and
edit issues there. This means that issues and docs can dynamically link between one another
seamlessly, the intent there is better documentation through stronger issue/code/doc ties but
it's still an experiment on if that works out.

- sisyphus resolves wikilinks the way Obsidian does: `[[name]]`, `[[name#Heading]]`, and
  `[[name|text]]` match a file by name in any directory, ignoring case. A link to an issue keeps
  working when the issue moves between `open/`, `in-progress/`, and `closed/`.
  `sisyphus links <file>` reports any link that does not resolve.
- Like Obsidian, sisyphus ignores wikilinks inside code blocks and inline code.
- In frontmatter, a wikilink is quoted (`parent: "[[some-epic]]"`) so the YAML stays valid.
  sisyphus writes them that way.
- Obsidian keeps its own settings in `.obsidian/`. Those are per-user, so keep them out of version
  control; this repo ignores them.

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
command it wraps. See [[sisyphus-mcp]] for the full list of tools.
