# sisyphus

> **DO NOT USE.** This is a personal project and it's very vibes based atm. That may change later but for now this is primarily a project built just for me.

sisyphus is an issue tracker that keeps the issues of a project in the repo, next to the code. Each
issue is one Markdown file in `issues/open/`, `issues/in-progress/`, or `issues/closed/`, so issues
change in the same commits as the code, and you can read and edit them with any editor. The
`sisyphus` command creates issues, moves them through their states, finds and draws them, and checks
the links between issues and documents. An MCP server gives the same commands to AI agents.

## Install

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
```

Then, in your repo:

```bash
sisyphus init
```

`init` writes `issues/TEMPLATE.md` and the three state directories, and nothing else.

## A first session

```console
$ sisyphus new --title "Fix the login bug" --tags bug --priority high
issues/open/fix-login-bug.md
$ sisyphus new --title "Add a dark mode" --tags feature
issues/open/add-dark-mode.md
$ sisyphus start fix-login-bug --owner alice
issues/in-progress/fix-login-bug.md
$ sisyphus list
name           title              state        priority  owner  tags
-------------  -----------------  -----------  --------  -----  -------
add-dark-mode  Add a dark mode    open         medium           feature
fix-login-bug  Fix the login bug  in-progress  high      alice  bug
$ sisyphus close fix-login-bug -m "The session cache now expires after a password reset."
issues/closed/fix-login-bug.md
$ sisyphus show fix-login-bug
title:      Fix the login bug
state:      closed
resolution: completed
priority:   high
effort:     medium
tags:       [bug]
created:    2026-10-11
closed:     2026-10-11
owner:      alice

# Fix the login bug

## Resolution

The session cache now expires after a password reset.
```

The name of each issue comes from its title. Each command that changes an issue prints the path of
the file, and moves the file when the state changes.

## The main commands

| To | Run |
|---|---|
| Create an issue | `sisyphus new --title "<title>"` |
| Start, close, or abandon it | `sisyphus start <name>`, `sisyphus close <name>`, `sisyphus abandon <name>` |
| Change its fields | `sisyphus update <name> --priority high --add-tags auth` |
| Write in a section | `sisyphus edit <name> notes "<text>" --append`, or `sisyphus open <name>` for your editor |
| See one issue | `sisyphus show <name>` |
| Find issues | `sisyphus list --ready`, `sisyphus search timeout` |
| Relate issues | `sisyphus parent <name> <epic>`, `sisyphus depends-on <name> <other>` |
| Draw an issue and its relatives | `sisyphus graph <name>` |
| Check the repo | `sisyphus lint` |

Every command has `--help` with examples. The commands that change an issue take `--json`, and the
commands that read issues print JSON with `--json`, for scripts. [docs/commands.md](docs/commands.md)
describes every command, and [docs/issue-format.md](docs/issue-format.md) describes the issue file:
its fields, states, and sections, and what you can change in the template.

## Shell completion

sisyphus completes commands, flags, issue names, states, priorities, and tags in bash, zsh, fish,
PowerShell, and Nushell. For bash, add this line to `~/.bashrc`:

```bash
source <(sisyphus completion bash)
```

See [docs/sisyphus-completion.md](docs/sisyphus-completion.md) for the other shells.

## Obsidian

sisyphus issues are plain Markdown with YAML frontmatter, so a repo that uses sisyphus also works
as an [Obsidian](https://obsidian.md/) vault: open the repo's root as a vault to browse, link, and
edit issues there. This means that issues and docs can dynamically link between one another
seamlessly, the intent there is better documentation through stronger issue/code/doc ties but
it's still an experiment on if that works out.

- sisyphus resolves wikilinks the way Obsidian does: `[[name]]`, `[[name#Heading]]`, and
  `[[name|text]]` match a file by name in any directory, ignoring case. A link to an issue keeps
  working when the issue moves between `open/`, `in-progress/`, and `closed/`.
  `sisyphus links <file>` reports any link that does not resolve, and `sisyphus lint` checks every
  file in the repo.
- Like Obsidian, sisyphus ignores wikilinks inside code blocks and inline code.
- In frontmatter, a wikilink is quoted (`parent: "[[some-epic]]"`) so the YAML stays valid.
  sisyphus writes them that way.
- Obsidian keeps its own settings in `.obsidian/`. Those are per-user, so add `.obsidian/` to your
  `.gitignore`.

## MCP server

`sisyphus-mcp` gives an AI agent every `sisyphus` command as an MCP tool, over stdio. It runs the
`sisyphus` binary, so install both:

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest
claude mcp add sisyphus -- sisyphus-mcp      # Claude Code
```

[docs/sisyphus-mcp.md](docs/sisyphus-mcp.md) tells how to set up other clients, lists the tools,
and shows how warnings and errors come back, with an example session and troubleshooting.

## Contributing

To work on sisyphus itself, read [CONTRIBUTING.md](CONTRIBUTING.md).
