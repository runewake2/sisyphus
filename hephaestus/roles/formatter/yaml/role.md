---
role: formatter/yaml
kind: tool
command: sh "$HEPHAESTUS_ROLE_DIR/format.sh"
trigger: request
on-failure: needs-human
listens: [bookmark-ready]
can-call-referendum: false
---

# Formatter: YAML

## Charter

Format the YAML files that the work changes, with prettier.

It changes only the layout of the files, never their meaning, and only the files that the work
since the base changes: `hephaestus files --changed`. It is a tool role: it uses no model. As a
reviewer in a referendum, it votes `yes` when the files are formatted, and `modify` with the
formatted files when they are not ([[roles/formatter/README|the formatter roles]]).

## Inputs

- A task message, with the issue and, in its data, the workspace to format. The base is `main`.
- A broadcast of the kind `bookmark-ready`.
- A referendum ballot request: hephaestus runs it in a scratch worktree of the version, with the
  base in `HEPHAESTUS_BASE`.

## Outputs

- The changed YAML files of the working copy, formatted, and the list of the files that changed.
- The issue of the task, closed with that list, or stopped with the output in its `Notes` and
  tagged `needs-human`.
- A ballot in a referendum: `yes`, or `modify` with the formatted files.

## Works with

- `implementer`: sends the `bookmark-ready` broadcast that starts it, and squashes the formatting into its change.
- `repo-check`: checks the format of the Go code. The formatter fixes what repo-check would report.

## Tools

`format.sh` in the directory of this role file: prettier, on the files of `hephaestus files --changed`.

## Ask a human

- When prettier is not installed: the output tells how to install it.
- When prettier fails: a file does not parse. The output tells the file.
