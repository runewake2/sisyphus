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
| `--owner` | The person or agent working on the issue. Default: the value of `SISYPHUS_OWNER`. |
| `-b`, `--bookmark` | The jj bookmark or git branch of the work. |
| `--workspace` | The jj workspace or git worktree where the work happens. Added to the issue's workspace history. |
| `--take` | Take the issue from its owner when another owner has it in progress. A warning names the old owner. |
| `--json` | Print the changed issue as JSON, with its path and the warnings. |

## Behavior

The same as [[sisyphus-update]] with the state `in-progress`, with the same warnings: for example
when an issue that this issue depends on is not closed. A closed issue cannot start; see
`update --reopen` for a close that was a mistake.

An issue that is in progress with a different owner is refused, so that two agents cannot both
work on it:

```console
$ sisyphus start fix-login-bug --owner a2
Error: Issue 'fix-login-bug' is in progress, owned by a1. Select another issue, or to take it from a1, add --take.
```

Another workspace of the repo that has started the issue also refuses it, before its work merges:
see [[sisyphus-claims]].

A start by the same owner changes nothing and succeeds, so an agent can repeat it. `--take` takes
the issue; use it only when the old owner has stopped the work. See
[[0022-refuse-to-take-a-claimed-issue]].

Without `--owner`, the owner is the value of the `SISYPHUS_OWNER` environment variable. When neither
gives an owner, `start` warns: nobody can tell who works on the issue.

## Output

The path of the issue's file, or with `--json` the changed issue.

## Examples

```bash
sisyphus start fix-login-bug
sisyphus start fix-login-bug --owner alice --bookmark ai/fix-login-bug
sisyphus start fix-login-bug --owner bob --take   # alice has stopped the work
```
