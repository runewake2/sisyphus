# sisyphus show

Print one issue, without first finding which state directory holds it.

## Usage

```bash
sisyphus show <name> [--json]
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]).

## Output

By default, the key fields, aligned, then the body:

```
title:      Fix login bug
state:      in-progress
priority:   high
effort:     medium
tags:       [bug]
owner:      alice
approver:   bob
bookmark:   ai/fix-login-bug
parent:     [[auth-overhaul]]
depends-on: []
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

## Examples

```bash
sisyphus show fix-login-bug
sisyphus show "[[fix-login-bug]]" --json | jq -r .state
```
