# sisyphus commands

sisyphus is a command-line issue manager. Every issue is one Markdown file with YAML frontmatter in
`issues/open/`, `issues/in-progress/`, or `issues/closed/`; the directory always matches the issue's
`state`. This page describes what every command has in common, then lists the commands. For the
fields, states, and sections of an issue file, and what a repo can change in its template, see
[[issue-format]].

## Commands

The groups are the same as in `sisyphus --help`.

### Create and change issues

| Command | What it does |
|---|---|
| [[sisyphus-new]] | Create an issue. |
| [[sisyphus-start]] | Start work on an issue (`update <name> in-progress`). |
| [[sisyphus-update]] | Change an issue's state, or edit its fields. |
| [[sisyphus-close]] | Close an issue as completed, with an optional Resolution text. |
| [[sisyphus-abandon]] | Close an issue as abandoned, with an optional Resolution text. |
| [[sisyphus-edit]] | Edit an issue in your editor, or replace or append to one section. |
| [[sisyphus-rename]] | Rename an issue, or move it to a directory, and update every link to it. |

### Read and find issues

| Command | What it does |
|---|---|
| [[sisyphus-show]] | Print one issue. |
| [[sisyphus-list]] | List and filter issues. |
| [[sisyphus-search]] | Search issues by content and filters. |
| [[sisyphus-graph]] | Draw an issue, everything below it, and the path above it. |
| [[sisyphus-claims]] | List the issues that each workspace has in progress. |

### Relate issues

| Command | What it does |
|---|---|
| [[sisyphus-parent]] | Set or remove an issue's parent. |
| [[sisyphus-depends-on]] | Add or remove an issue that must close first. |
| [[sisyphus-remote]] | Pin an issue to an external ticket URL. |

### Check names and links

| Command | What it does |
|---|---|
| [[sisyphus-slug]] | Turn text into a unique, valid issue name. |
| [[sisyphus-resolve]] | Find the file a wikilink points to. |
| [[sisyphus-links]] | Check every wikilink in a document. |
| [[sisyphus-lint]] | Check every issue and every wikilink in the repo. |

### Set up

| Command | What it does |
|---|---|
| [[sisyphus-init]] | Set up `issues/` in a repo. |
| [[sisyphus-completion]] | Print the shell completion script. |

[[sisyphus-mcp]] gives an agent all of these as tools over the Model Context Protocol. [[agent-guide]] tells
how to let AI agents use sisyphus.

## How sisyphus finds the repo

Every command except `init` needs the repo root: the directory that contains `issues/TEMPLATE.md`.
sisyphus looks for it in the current directory and each parent directory in turn. If that finds
nothing, it tries the same from the directory of the `sisyphus` binary. So you can run sisyphus
from anywhere inside a repo, and each jj workspace uses its own files.

`-C <directory>` (`--directory`), before or after the command, makes the search start in that
directory instead of the current one, as `git -C` does: `sisyphus -C ../other-repo list`. A relative
path is taken from the current directory. `init` writes into `-C` too.

When there is no repo root, a command fails with:

```
Error: Cannot find the repo root. Run sisyphus in a directory below the one that contains issues/TEMPLATE.md. To set up issue tracking here, run: sisyphus init
```

## Environment variables

| Variable | Use |
|---|---|
| `SISYPHUS_OWNER` | The owner of an issue that starts, when `--owner` is not given: [[sisyphus-start]], `update <name> in-progress`, and `new --state in-progress`. `list --mine` and `search --mine` filter on it. An agent harness can set it once for each agent, so that every claim has an owner. |
| `VISUAL`, `EDITOR` | The editor of `sisyphus edit <name>`. |

## Several commands at the same time

Several people and agents can run sisyphus in one repo at the same time:

- A command writes an issue file to a temporary file and renames it, so a command that reads at the
  same time sees the old file or the new file, never a part of one.
- A command that changes issues holds a lock from the time it reads them to the time it writes
  them, so two changes cannot overwrite each other. Of two starts of one issue at the same time,
  one succeeds and the other is refused ([[sisyphus-start]]). The lock is a file outside the working
  copy: `sisyphus/lock` in the jj repo directory (`.jj/repo`) or in the git directory, which every
  workspace and worktree shares. In a directory without jj or git, it is `issues/.lock`.
- A command waits up to 10 seconds for the lock, then fails with the path of the lock. A lock older
  than 30 seconds is left from a command that stopped, and is removed.
- `sisyphus edit <name>` in your editor does not hold the lock while the editor is open.

See [[0023-lock-and-write-issues-atomically]].

## Invalid values

A `state`, `priority`, `effort`, or `resolution` that is not valid is an error that lists the valid
values. When one valid value is near, the error suggests it: a common word (`done` → `closed`,
`start` → `in-progress`), the only value that starts with it (`hi` → `high`), or the only value
within two typing errors (`hihg` → `high`). sisyphus never uses the suggestion by itself.

## Naming an issue

An issue's file name, without `.md`, is 2-6 lowercase words in kebab-case, for example
`fix-login-bug`. The file can sit directly in its state directory, or in kebab-case subdirectories
below it, for example `issues/open/web/auth/fix-login-bug.md`. The issue's **full name** is its
path below the state directory, without `.md`: `web/auth/fix-login-bug`. The subdirectories stay
with the issue when its state changes. To move an issue to another directory, or to change its
name, use [[sisyphus-rename]]: it changes every link to the issue in the repo. Prefer a move to a
directory over a new file name: commit messages, other branches, and other repos refer to the file
name, and [[sisyphus-rename]] cannot change them.

Every command that takes an existing issue accepts any of these forms:

| Form | Example |
|---|---|
| The full name | `web/auth/fix-login-bug` |
| The end of the full name | `fix-login-bug`, `auth/fix-login-bug` |
| A wikilink | `[[web/auth/fix-login-bug]]`, `[[fix-login-bug]]` |
| A hashtag | `#web/auth/fix-login-bug` |
| A path | `issues/open/web/auth/fix-login-bug.md` |

Every form matches without regard to case, as wikilinks do in Obsidian: `[[Fix-Login-Bug]]` gives
`fix-login-bug`. Output always uses the real, lowercase name.

A full name is also the fastest form: sisyphus finds it with one look at each state directory,
while a shorter form makes it read every state directory first. In a repo with thousands of issues,
scripts and agents should use full names.

A full name always gives its own issue. A shorter form must give exactly one issue: if two issues
end with it, for example `web/fix-login-bug` and `mobile/fix-login-bug`, the command fails, lists
the full names, and changes nothing. Use the full name instead.

You never need to know which state directory holds an issue. If an issue's file is in more than
one state directory, the command fails and lists the copies, so you can remove the duplicate.

sisyphus writes links between issues (`parent`, `depends-on`, `deferred-from`) with the full name,
so that a later issue with the same file name does not make them ambiguous. Obsidian resolves a
link such as `[[web/auth/fix-login-bug]]` by the end of the path in the same way.

A name that gives no issue is an error. When some issues share words with the name, the error
suggests up to 3 of them, with their titles. Issues that are not closed come first:

```
Error: No issue named 'docs-site-theme' in issues/open, issues/in-progress, or issues/closed. Did you mean: update-docs-site (Update the docs site theme)?
```

A near word counts too, for example `convention` for `conventions`. sisyphus never uses a
suggestion by itself.

## Output and exit codes

- A command that writes an issue prints the path of the file it wrote, from the repo root.
- Warnings go to stderr, start with `Warning:`, and do not fail the command.
- Errors go to stderr, start with `Error:`, and make the command exit with status 1.
- `show`, `list`, `search`, `graph`, `links`, `lint`, `claims`, and every command that changes an
  issue accept `--json` for output a program can read.
- `lint` also exits with status 1 when it finds problems.

[[output]] has the rules for stdout, stderr, and exit codes, and the shape of each JSON output.
- `sisyphus --version` prints the version the binary was built from, in any directory.
- `sisyphus <command> --help` prints every flag of a command.

## What sisyphus does not do

- It does not run `jj` or create, move, or delete bookmarks. `bookmark` is only a field it records.
- It runs `git mv` only to move an issue file between state directories, and only when the repo
  root contains `.git`.
- It skips `.git`, `.jj`, `bin`, `obj`, and `node_modules` when it searches the repo.
