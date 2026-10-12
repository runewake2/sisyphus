# Formatter roles

The formatter roles format the files of a change, one role for each language. Each is a tool role:
a program does its work, with no model ([[0003-a-tool-agent-runs-its-command-for-each-message]]).

| Role | Files | Tool |
|---|---|---|
| `formatter/golang` | `*.go` | `gofmt` |
| `formatter/csharp` | `*.cs` | `dotnet format whitespace`, in folder mode |
| `formatter/yaml` | `*.yaml`, `*.yml` | prettier |
| `formatter/json` | `*.json`, but not the benchmark results in `design/ollama-local-models/results/` | prettier |
| `formatter/markdown` | `*.md`, but not `issues/`, `.claude/`, or `referendums/` | prettier |

The role files: [[roles/formatter/golang/role]], [[roles/formatter/csharp/role]],
[[roles/formatter/yaml/role]], [[roles/formatter/json/role]], and [[roles/formatter/markdown/role]].

## Layout

Each language has a directory named after it, with two files:

- `role.md`: the role file ([[role-file]]). Its command is `sh "$HEPHAESTUS_ROLE_DIR/format.sh"`.
- `format.sh`: the implementation. hephaestus gives the directory of the role file in
  `HEPHAESTUS_ROLE_DIR`, so the script comes from the roles of the caller, also when it runs in the
  scratch worktree of a referendum.

`prettier.sh` is shared by the roles that use prettier. It uses an installed `prettier`, or else
prettier 3.3.3 through `npx`.

## Rules for every formatter

- Format only the files that the work changes, from `hephaestus files --changed`: the commits from
  the fork point of the base, and the changes that are not committed. The base is
  `HEPHAESTUS_BASE` in a referendum, else `main`. A formatter never changes a file that the work did
  not touch.
- When no file of the language changed, end with exit code 0, and do not need the tool.
- When the tool is not installed, end with exit code 1, and tell how to install it.
- Change only the layout, never the meaning. `formatter/markdown` does not format the code blocks
  of a document, because their text is an example of exact input or output.

To add a language, add a directory with a `role.md` and a `format.sh`, and a row to the table.
