# sisyphus update

Change an issue's state and move its file to the directory of the new state, or edit its fields.

## Usage

```bash
sisyphus update <name> <state> [flags]
sisyphus update <name> [flags]          # change fields, keep the state
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]). `<state>` is `open`,
`in-progress`, or `closed`. Without `<state>`, the state does not change, and at least one flag must
be given. When the state does not change and `--resolution` is not given, a closed issue keeps its
resolution and its closed date. The warnings about closing an issue come only when the issue closes.

## Flags

| Flag | Meaning |
|---|---|
| `-r`, `--resolution` | `completed` or `abandoned`. Only when the new state is closed; defaults to `completed`. |
| `-b`, `--bookmark` | The jj bookmark of the work. Set it when work starts. |
| `--owner` | The person or agent working on the issue. |
| `--approver` | The person or agent who accepts the issue when it closes. |
| `--workspace` | A jj workspace where work on the issue happens. Added to `workspaces`. |
| `--priority` | A new priority: `critical`, `high`, `medium`, or `low`. |
| `--effort` | A new effort: `small`, `medium`, `large`, or `x-large`. |
| `--tags` | New tags, comma-separated. They replace the old tags. A tag cannot contain a space. |
| `--metadata` | `key=value` notes to add or change. Repeat the flag, or separate pairs with commas. |

A flag you do not give leaves its field as it is, except where the rules below say otherwise.

## Rules

**State**

- A closed issue cannot reopen: `update` to `open` or `in-progress` fails. Create a new issue that
  links to the old one instead.
- Closing sets `resolution` (default `completed`) and sets `closed` to today. Any other state
  clears both.
- `--resolution` with a new state other than `closed` is an error.
- If the issue's `state` field does not match its directory, `update` warns and corrects both.

**Fields that follow the work**

| Field | On `open` | On `in-progress` or `closed` |
|---|---|---|
| `bookmark`, `owner` | Cleared: nobody is working on the issue. A value in `--owner` or `--bookmark` is an error. | Set if the flag is given; otherwise kept. |
| `workspaces` | Kept. | Kept. |
| `approver` | Set if the flag is given; otherwise kept. | Same. |
| `metadata` | Merged: given keys are added or replaced. | Same. |

`workspaces` is a history: `--workspace` adds a name once, and nothing ever removes one. `metadata`
is never cleared automatically.

**Moving the file**

The file moves to `issues/<state>/<name>.md`, where `<name>` is the full name, so the issue keeps
its subdirectories. If the repo root contains `.git` and git tracks the
file, sisyphus moves it with `git mv`, so git records a rename. Otherwise it renames the file; jj
detects the rename itself. In a colocated jj repo, git does not track an issue until jj commits it.

## Output

The path of the issue's file after the update.

Warnings:

- The issue is in-progress but has no bookmark.
- The issue is in-progress, but some of its `depends-on` issues are not closed. The warning names
  them. The issue still starts.
- The issue was in one state directory but its `state` field said another; both are corrected.
- Closing an issue that has sub-issues that are not closed.
- Closing an issue that open or in-progress issues still depend on.
- Closing an issue whose `Resolution` section still holds the template placeholder. Complete it
  with [[sisyphus-edit]].

## Examples

```bash
# Start work
sisyphus update fix-login-bug in-progress --bookmark ai/fix-login-bug --owner alice \
  --workspace fix-login-bug --metadata session-id=abc123

# Re-prioritize without changing state
sisyphus update fix-login-bug in-progress --priority critical --tags bug,auth
sisyphus update fix-login-bug --effort small            # the state does not change

# Finish, or give up
sisyphus update fix-login-bug closed
sisyphus update add-dark-mode closed --resolution abandoned

# Stop work without closing
sisyphus update retry-backoff open
```
