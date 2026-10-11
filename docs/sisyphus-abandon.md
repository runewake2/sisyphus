# sisyphus abandon

Close an issue as abandoned: the work will not happen. It is `sisyphus update <name> closed --resolution abandoned` in fewer words.

## Usage

```bash
sisyphus abandon <name> [-m "<text>"]
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]).

## Flags

| Flag | Meaning |
|---|---|
| `-m`, `--message` | The text of the Resolution section: why the work stopped, and what replaces it. |
| `--json` | Print the changed issue as JSON, with its path and the warnings. |

## Behavior

The same as [[sisyphus-update]] with the state `closed` and the resolution `abandoned`, with the same
warnings: open sub-issues, issues that still depend on this one, and an empty Resolution section
when `-m` is not given. `-m` replaces the Resolution section, with the rules of
[[sisyphus-edit]].

## Output

The path of the issue's file, or with `--json` the changed issue.

## Examples

```bash
sisyphus abandon old-login-form
sisyphus abandon old-login-form -m "Login moved to SSO in [[sso-login]]."
```
