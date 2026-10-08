# sisyphus update

Change an issue's state and move its file to the directory of the new state, or edit its fields.

## Usage

```bash
sisyphus update <name> <state> [flags]
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]). `<state>` is `open`,
`in-progress`, or `closed`. To edit fields without changing state, give the issue's current state.

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
| `--tags` | New tags, comma-separated. They replace the old tags. |
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
| `bookmark`, `owner` | Cleared: nobody is working on the issue. | Set if the flag is given; otherwise kept. |
| `workspaces` | Kept. | Kept. |
| `approver` | Set if the flag is given; otherwise kept. | Same. |
| `metadata` | Merged: given keys are added or replaced. | Same. |

`workspaces` is a history: `--workspace` adds a name once, and nothing ever removes one. `metadata`
is never cleared automatically.

**Moving the file**

The file moves to `issues/<state>/<name>.md`. If the repo root contains `.git`, sisyphus moves it
with `git mv`, so git records a rename. Otherwise it renames the file; jj detects the rename itself.

## Output

The path of the issue's file after the update.

Warnings:

- The issue is in-progress but has no bookmark.
- The issue was in one state directory but its `state` field said another; both are corrected.
- Closing an issue that has sub-issues that are not closed.
- Closing an issue that open or in-progress issues still depend on.

## Examples

```bash
# Start work
sisyphus update fix-login-bug in-progress --bookmark ai/fix-login-bug --owner alice \
  --workspace fix-login-bug --metadata session-id=abc123

# Re-prioritize without changing state
sisyphus update fix-login-bug in-progress --priority critical --tags bug,auth

# Finish, or give up
sisyphus update fix-login-bug closed
sisyphus update add-dark-mode closed --resolution abandoned

# Stop work without closing
sisyphus update retry-backoff open
```
