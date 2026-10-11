# sisyphus start

Start work on an issue. It is `sisyphus update <name> in-progress` in fewer words.

## Usage

```bash
sisyphus start <name> [flags]
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]).

## Flags

| Flag | Meaning |
|---|---|
| `--owner` | The person or agent working on the issue. |
| `-b`, `--bookmark` | The jj bookmark of the work. |
| `--workspace` | The jj workspace where the work happens. Added to the issue's workspace history. |
| `--json` | Print the changed issue as JSON, with its path and the warnings. |

## Behavior

The same as [[sisyphus-update]] with the state `in-progress`, with the same warnings: for example
when an issue that this issue depends on is not closed. A closed issue cannot start; see
`update --reopen` for a close that was a mistake.

## Output

The path of the issue's file, or with `--json` the changed issue.

## Examples

```bash
sisyphus start fix-login-bug --owner alice --bookmark ai/fix-login-bug
```
