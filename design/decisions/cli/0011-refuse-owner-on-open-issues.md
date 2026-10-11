# 0011: Refuse an owner or a bookmark for an open issue

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.37]]

## Context

`owner` and `bookmark` describe active work, so sisyphus clears them when an issue is open (see
[[CONTRIBUTING#Frontmatter]]). But `sisyphus update <name> open --owner carol` exited 0 and dropped
the owner without a message, and `sisyphus new --owner carol` wrote an owner on an open issue. The
user asked for a value and got something else ([[update-open-drops-owner]]).

## Decision

`sisyphus new` and `sisyphus update` fail when the state is `open` and `--owner` or `--bookmark`
has a value that is not empty. The error says that an open issue has no owner and no bookmark, and
tells the user to leave out the flags or to use the state `in-progress`. An empty value
(`--owner ""`) is still accepted, because it asks for what the rule does anyway.

## Alternatives considered

- **Warn and drop the value.** The command would succeed, so a script or an agent that does not
  read stderr would still lose the value without noticing.
- **Keep the value on an open issue.** This would let an owner mean "planned owner" on an open
  issue. It changes the meaning of the field and the rule in CONTRIBUTING, and `update` to open
  would then have to keep an old owner too.

## Consequences

- A command that silently did something different from the request now fails, with no change to
  the file.
- Anyone who used `new --owner` to note a planned owner must use `--metadata` or the body instead.
