# sisyphus commands

sisyphus is a command-line issue manager. Every issue is one Markdown file with YAML frontmatter in
`issues/open/`, `issues/in-progress/`, or `issues/closed/`; the directory always matches the issue's
`state`. This page describes what every command has in common, then lists the commands.

## Commands

| Command | What it does |
|---|---|
| [[sisyphus-init]] | Set up `issues/` in a repo. |
| [[sisyphus-new]] | Create an issue. |
| [[sisyphus-update]] | Change an issue's state, or edit its fields. |
| [[sisyphus-edit]] | Replace or append to one section of an issue's body. |
| [[sisyphus-show]] | Print one issue. |
| [[sisyphus-list]] | List and filter issues. |
| [[sisyphus-search]] | Search issues by content and filters. |
| [[sisyphus-graph]] | Draw an issue, everything below it, and the path above it. |
| [[sisyphus-parent]] | Set or remove an issue's parent. |
| [[sisyphus-depends-on]] | Add or remove an issue that must close first. |
| [[sisyphus-remote]] | Pin an issue to an external ticket URL. |
| [[sisyphus-slug]] | Turn text into a unique, valid issue name. |
| [[sisyphus-resolve]] | Find the file a wikilink points to. |
| [[sisyphus-links]] | Check every wikilink in a document. |

[[sisyphus-mcp]] gives an agent all of these as tools over the Model Context Protocol.

## How sisyphus finds the repo

Every command except `init` needs the repo root: the directory that contains `issues/TEMPLATE.md`.
sisyphus looks for it in the current directory and each parent directory in turn. If that finds
nothing, it tries the same from the directory of the `sisyphus` binary. So you can run sisyphus
from anywhere inside a repo, and each jj workspace uses its own files.

When there is no repo root, a command fails with:

```
Error: Cannot find the repo root. Run sisyphus in a directory below the one that contains issues/TEMPLATE.md.
```

## Naming an issue

An issue's file name, without `.md`, is 2-6 lowercase words in kebab-case, for example
`fix-login-bug`. The file can sit directly in its state directory, or in kebab-case subdirectories
below it, for example `issues/open/web/auth/fix-login-bug.md`. The issue's **full name** is its
path below the state directory, without `.md`: `web/auth/fix-login-bug`. The subdirectories stay
with the issue when its state changes. A full name is permanent: never rename or move an issue.

Every command that takes an existing issue accepts any of these forms:

| Form | Example |
|---|---|
| The full name | `web/auth/fix-login-bug` |
| The end of the full name | `fix-login-bug`, `auth/fix-login-bug` |
| A wikilink | `[[web/auth/fix-login-bug]]`, `[[fix-login-bug]]` |
| A hashtag | `#web/auth/fix-login-bug` |
| A path | `issues/open/web/auth/fix-login-bug.md` |

A full name always gives its own issue. A shorter form must give exactly one issue: if two issues
end with it, for example `web/fix-login-bug` and `mobile/fix-login-bug`, the command fails, lists
the full names, and changes nothing. Use the full name instead.

You never need to know which state directory holds an issue. If an issue's file is in more than
one state directory, the command fails and lists the copies, so you can remove the duplicate.

sisyphus writes links between issues (`parent`, `depends-on`, `deferred-from`) with the full name,
so that a later issue with the same file name does not make them ambiguous. Obsidian resolves a
link such as `[[web/auth/fix-login-bug]]` by the end of the path in the same way.

## Output and exit codes

- A command that writes an issue prints the path of the file it wrote, from the repo root.
- Warnings go to stderr, start with `Warning:`, and do not fail the command.
- Errors go to stderr, start with `Error:`, and make the command exit with status 1.
- `show`, `list`, `search`, `graph`, and `links` accept `--json` for output a program can read.
- `sisyphus --version` prints the version the binary was built from, in any directory.
- `sisyphus <command> --help` prints every flag of a command.

## What sisyphus does not do

- It does not run `jj` or create, move, or delete bookmarks. `bookmark` is only a field it records.
- It runs `git mv` only to move an issue file between state directories, and only when the repo
  root contains `.git`.
- It skips `.git`, `.jj`, `bin`, `obj`, and `node_modules` when it searches the repo.
