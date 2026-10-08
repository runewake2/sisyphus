# sisyphus init

Set up issue tracking in a repo.

## Usage

```bash
sisyphus init [--dir <directory>] [--force]
```

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--dir` | `.` | The root directory of the repo. |
| `--force` | off | Replace files that already exist. |

## What it does

`init` writes these files into the repo, and nothing else:

```
issues/TEMPLATE.md
issues/open/.gitkeep
issues/in-progress/.gitkeep
issues/closed/.gitkeep
```

`issues/TEMPLATE.md` defines every issue's frontmatter fields and section headings. After `init`,
every other command works in the repo, because the repo root is the directory that contains
`issues/TEMPLATE.md` (see [[commands#How sisyphus finds the repo]]).

- A file that already exists stays as it is, with a warning. `--force` replaces it.
- You can change `issues/TEMPLATE.md` to fit your repo. [[sisyphus-new]] reads it from the repo,
  not from the binary.
- `init` does not write a contributing guide, agent rules, a changelog, a version file, or CI or
  GitHub workflows. Those belong to a project-scaffolding template, not to sisyphus.

## Output

The path of each file written, one per line. For each file skipped:

```
Warning: issues/TEMPLATE.md exists, so init did not change it. Use --force to replace it.
```

## Examples

```bash
sisyphus init
sisyphus init --dir ../other-repo
sisyphus init --force    # restore the default issues/TEMPLATE.md
```
