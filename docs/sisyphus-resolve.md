# sisyphus resolve

Find the file that a wikilink points to, anywhere in the repo.

## Usage

```bash
sisyphus resolve <link> [--all] [--absolute]
```

`<link>` is a wikilink (`[[name]]`, `[[name#Heading]]`, `[[name|text]]`), a `#name`, or a bare name
or path. Only the target matters: the heading and the display text are ignored.

## Flags

| Flag | Meaning |
|---|---|
| `-a`, `--all` | Print every match. Without it, more than one match is an error. |
| `--absolute` | Print absolute paths. The default is paths from the repo root. |

## How a link matches

The same way Obsidian matches it:

- By file name, with or without `.md`, in any directory, ignoring case.
- A link that contains `/` matches the end of a path, for example `[[example/README]]`.
- An exact path from the repo root wins over every other match.
- `.git`, `.jj`, `bin`, `obj`, and `node_modules` are skipped, and so is every file that
  `.gitignore` ignores. In a git repo, git decides, with every `.gitignore` and
  `.git/info/exclude`; without git, sisyphus applies the simple rules of the root `.gitignore`
  (names, globs, paths from the root, a trailing `/` for a directory; `!` is not supported). The
  same files count for [[sisyphus-links]], [[sisyphus-lint]], `new`, and `slug`. See
  [[0019-skip-ignored-files-in-the-repo-walk]].

## Output

The path of the match, or with `--all` every match, one per line.

Errors:

- No file matches the link.
- More than one file matches: the error lists them and suggests a longer link, for example
  `[[design/example/README]]`, or `--all`.

## Examples

```bash
sisyphus resolve fix-login-bug
sisyphus resolve "[[CONTRIBUTING#Issues]]"
sisyphus resolve README --all
```
