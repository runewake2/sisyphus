# sisyphus show

Print one issue, without first finding which state directory holds it.

## Usage

```bash
sisyphus show <name> [--all] [--json]
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]).

## Output

By default, the fields that have a value, aligned, then the body without the sections that hold only
the template text (the `<...>` placeholders). `title`, `state`, `priority`, `effort`, and `created`
always show. `resolution` and `closed` have a value only when the issue is closed, so they show
whether a close or an abandon worked. `--all` prints every field and every section:

```
title:      Fix login bug
state:      in-progress
priority:   high
effort:     medium
tags:       [bug]
created:    2026-10-08
owner:      alice
approver:   bob
bookmark:   ai/fix-login-bug
parent:     [[auth-overhaul]]
remote:     https://github.com/acme/widgets/issues/42
metadata:   {session-id: "abc123"}

# Fix login bug

## Summary
...
```

With `--json`, one object with every frontmatter field plus the body. Lists and maps are real JSON
arrays and objects, `depends-on` holds plain issue names, and empty optional fields are left out:

```json
{
  "name": "fix-login-bug",
  "title": "Fix login bug",
  "state": "in-progress",
  "priority": "high",
  "effort": "medium",
  "tags": ["bug"],
  "created": "2026-10-08",
  "owner": "alice",
  "bookmark": "ai/fix-login-bug",
  "parent": "[[auth-overhaul]]",
  "remote": "https://github.com/acme/widgets/issues/42",
  "metadata": {"session-id": "abc123"},
  "body": "\n# Fix login bug\n..."
}
```

A missing issue, or an issue in more than one state directory, is an error.

Warnings, on stderr: one for each issue file that sisyphus cannot read, or whose frontmatter has a
line that is not a `key: value` field, a value that opens `[`, `{`, or a quote and does not close
it, a `state` that does not match its directory, or a `priority`, `effort`, or `resolution` that is
not valid. `show` checks the issue that it prints. The warnings do not change stdout, so `--json` output stays valid. See
[[0012-warn-about-invalid-issues-on-read]].

## Examples

```bash
sisyphus show fix-login-bug
sisyphus show "[[fix-login-bug]]" --json | jq -r .state
```
