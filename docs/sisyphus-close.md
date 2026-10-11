# sisyphus close

Close an issue as completed. It is `sisyphus update <name> closed` in fewer words.

## Usage

```bash
sisyphus close <name> [-m "<text>"]
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]).

## Flags

| Flag | Meaning |
|---|---|
| `-m`, `--message` | The text of the Resolution section: what was done. |
| `--json` | Print the changed issue as JSON, with its path and the warnings. |

## Behavior

The same as [[sisyphus-update]] with the state `closed` and the resolution `completed`, with the same
warnings: open sub-issues, issues that still depend on this one, and an empty Resolution section
when `-m` is not given. `-m` replaces the Resolution section, with the rules of
[[sisyphus-edit]].

## Output

The path of the issue's file, or with `--json` the changed issue.

## Examples

```bash
sisyphus close fix-login-bug -m "Fixed in [[0.0.9]]: the session cache now expires."
```
