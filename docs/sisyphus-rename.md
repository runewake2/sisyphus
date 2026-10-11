# sisyphus rename

Rename an issue, or move it to another directory, and change every link to it in the repo.

## Usage

```bash
sisyphus rename <name> <new-name> [flags]
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]). `<new-name>` is a full
name: 2-6 lowercase words in kebab-case, optionally below kebab-case directories, for example
`web/auth/fix-login-bug`.

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the move and each link that would change, and change nothing. |
| `--json` | Print the renamed issue as JSON, as [[sisyphus-show]] `--json` does, with its path and the warnings. With `--dry-run`, print the plan as JSON. |

## Behavior

- The file moves to `issues/<state>/<new-name>.md`. The state, the title, and the body do not
  change. Directories that the move leaves empty are removed. If git tracks the file, sisyphus moves
  it with `git mv`, as [[sisyphus-update]] does.
- Every wikilink in a Markdown file of the repo that points to the issue changes, also in the issue
  itself, in the frontmatter of other issues (`parent`, `depends-on`, `deferred-from`), and in
  changelogs and design documents. The heading (`#...`) and the text after `|` stay. Links in code
  blocks and inline code do not change, because they are not links.
- A link that has only the file name, for example `[[fix-login-bug]]`, stays so while no other
  file has the new file name. So a move to a directory changes few links. Every other link, and
  every link in the frontmatter, gets the full name, for example `[[web/fix-login-bug]]`.
- If another file has the new file name, `rename` warns, and every link gets the full name.
- The claim of the issue ([[sisyphus-claims]]) moves to the new name. An issue that another
  workspace has in progress is refused: its work still uses the old name.
- A name that is not valid, or that another issue has, is an error, and nothing changes.
- `rename` cannot change commit messages (`Fixes #fix-login-bug`), other branches and workspaces,
  or other repos. A move to a directory keeps the file name, so `#fix-login-bug` and
  `[[fix-login-bug]]` still find the issue. Prefer it over a new file name.

## Output

The path of the renamed issue's file, or with `--json` the renamed issue.

With `--dry-run`, the move and one line for each link that would change:

```
issues/open/fix-login-bug.md -> issues/open/web/fix-login-bug.md
issues/open/split-login-form.md:17: [[fix-login-bug]] -> [[web/fix-login-bug]]
```

With `--dry-run --json`, an object with `from`, `to`, and `links`, an array of objects with
`file`, `line`, `from`, and `to`.

## Examples

```bash
sisyphus rename fix-login-bug web/auth/fix-login-bug   # move it to a directory
sisyphus rename fix-login-bug fix-login-loop
sisyphus rename fix-login-bug web/fix-login-bug --dry-run
```
