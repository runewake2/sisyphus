# sisyphus remote

Pin an issue to the external ticket that tracks it, for example a GitHub issue or a Jira ticket,
or remove that pin.

## Usage

```bash
sisyphus remote <name> <url>
sisyphus remote <name> --clear
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]). Give exactly one of
`<url>` and `--clear`.

## Flags

| Flag | Meaning |
|---|---|
| `--clear` | Remove the remote. Not with `<url>`. |
| `--json` | Print the changed issue as JSON, as [[sisyphus-show]] `--json` does, with its path and the warnings. |

## Behavior

- An issue has at most one remote. Setting one replaces the old one.
- The value must be a URL with a scheme and a host, for example
  `https://github.com/acme/widgets/issues/42`. sisyphus does not check that the ticket exists.
- The URL is written quoted: `remote: "https://github.com/acme/widgets/issues/42"`.

sisyphus itself never talks to the tracker. The planned integration plugins will use `remote` to
decide which plugin handles an issue.

`sisyphus new --remote <url>` sets the remote when the issue is created.

## Output

With `--json`, the changed issue as [[sisyphus-show]] prints it with `--json`, plus `path` (the file)
and `warnings` (a list; the warnings do not go to stderr then). See [[0.0.72]].

The path of the issue's file.

## Examples

```bash
sisyphus remote fix-login-bug https://github.com/acme/widgets/issues/42
sisyphus remote fix-login-bug --clear
```
