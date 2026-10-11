# The issue file

Each sisyphus issue is one Markdown file: YAML frontmatter with the fields of the issue, then a body
with a title and sections. This page tells what each part means, what sisyphus needs, and what a
repo can change in its template. For the commands, see [[commands]].

## An example

`issues/in-progress/web/fix-login-bug.md`:

```markdown
---
title: "Fix the login loop after a password reset"
state: in-progress
resolution:
priority: high
effort: small
tags: [bug, auth]
created: 2026-10-08
closed:
owner: alice
approver: bob
bookmark: ai/fix-login-bug
workspaces: [fix-login-bug]
metadata: {session-id: "abc123"}
deferred-from:
parent: "[[auth-overhaul]]"
depends-on: ["[[session-cache-expiry]]"]
remote: "https://github.com/acme/widgets/issues/42"
---

# Fix the login loop after a password reset

## Summary

After a password reset, the login page sends the user back to itself. Fix the redirect.

## Context

The session cache keeps the old token. See [[session-cache-expiry]].

## Acceptance criteria

- [ ] A user who resets the password can log in once, with no loop.

## Out of scope

Single sign-on: [[sso-login]].

## Notes

2026-10-09: the bug is in the session cache, not in the redirect.

## Resolution

<Complete this section when you close the issue.>
```

## Where issues live

```
issues/
  TEMPLATE.md       The template that sisyphus new copies
  open/             Work that nobody has started
  in-progress/      Work that someone has started
  closed/           Work that is complete or abandoned
```

The path of the file below `issues/<state>/`, without `.md`, is the **full name** of the issue, for
example `web/fix-login-bug`. The file name alone, `fix-login-bug`, also names the issue while no
other issue has that file name. A name is 2-6 lowercase words in kebab-case, below optional
kebab-case directories. The name is the permanent id of the issue: links use it, so do not rename an
issue. See [[commands#Naming an issue]].

## Fields

| Field | Values | Meaning |
|---|---|---|
| `title` | Text | A short imperative title. `new` sets it from `--title`, or from the name. |
| `state` | `open`, `in-progress`, `closed` | Must match the directory that holds the file. `update` moves the file and changes the field together. |
| `resolution` | `completed`, `abandoned`, or empty | Empty until the issue closes. A close without `--resolution` is `completed`. |
| `priority` | `critical`, `high`, `medium`, `low` | `critical`: blocks other work. `high`: do next. `medium`: the default. `low`: do when there is time. |
| `effort` | `small`, `medium`, `large`, `x-large` | The size of the work. `medium` is the default. Split `x-large` work into sub-issues. |
| `tags` | A list, for example `[bug, auth]` | Free-form, with no spaces in a tag. Built-in: `bug`, `feature`. |
| `created` | `YYYY-MM-DD` | The date that `new` created the issue. |
| `closed` | `YYYY-MM-DD`, or empty | The date that the issue closed. Cleared when the issue is reopened. |
| `owner` | A person or an agent | Who works on the issue. Cleared when the issue goes back to open. |
| `approver` | A person or an agent | Who accepts the issue when it closes. Never cleared. |
| `bookmark` | A jj bookmark or git branch | The work on the issue. Cleared when the issue goes back to open. |
| `workspaces` | A list | The jj workspaces or git worktrees where work on the issue happened. `--workspace` adds one; entries are never removed. |
| `metadata` | `{key: value, ...}` | Free notes, for example an agent session id to resume the work. `--metadata key=value` adds or replaces one key. |
| `deferred-from` | `"[[issue]]"` or a bookmark | The work that deferred this issue. |
| `parent` | `"[[issue]]"` | The issue that this issue is a part of. See [[sisyphus-parent]]. |
| `depends-on` | A list of `"[[issue]]"` | Issues that must close before this issue can start. See [[sisyphus-depends-on]]. |
| `remote` | A URL | The GitHub issue or Jira ticket that tracks the issue outside the repo. See [[sisyphus-remote]]. |

Links in the frontmatter are quoted wikilinks, so the YAML stays valid and Obsidian resolves them.
Each field is on one line. A comment after `#` stays on the line when sisyphus changes the value.

## States and the lifecycle

| Step | Command | What changes |
|---|---|---|
| Create | `sisyphus new` | A new file in `issues/open/`, with `created` set. |
| Start | `sisyphus start` | The file moves to `issues/in-progress/`. `--owner`, `--bookmark`, and `--workspace` record who works on it, and where. |
| Stop without closing | `sisyphus update <name> open` | The file moves back to `issues/open/`. `owner` and `bookmark` are cleared. Record what remains in Notes. |
| Complete | `sisyphus close` | The file moves to `issues/closed/`, with `resolution: completed` and `closed` set. `-m` fills Resolution. |
| Abandon | `sisyphus abandon` | The same, with `resolution: abandoned`. Tell why in Resolution. |

A closed issue stays closed. For more work, create a new issue that links to the closed one. Only a
close that was a mistake, for example of the wrong issue, can be undone, with
`sisyphus update <name> open --reopen`.

A sub-issue is part of its parent: close the parent when all of its sub-issues are closed. A
dependency is a different link: the issue cannot start until the dependency closes. An issue that is
not closed, and that waits on no dependency or sub-issue that is not closed, is **available**: work
on it can start now (`sisyphus list --ready`).

## Body sections

| Section | Use | Used by sisyphus |
|---|---|---|
| `# <Title>` | The title as a heading. | `new` writes the title in place of `<Title>`. |
| Summary | What must change, and why, in one or two sentences. | |
| Context | The background that someone needs to start the work. | `new --context` fills it. |
| Acceptance criteria | The conditions that must be true before the issue is complete. | |
| Out of scope | Related work that the issue does not include. | |
| Notes | Progress, findings, and open questions while the work continues. | |
| Resolution | Completed: what was done. Abandoned: why the work stopped. | `close -m`, `abandon -m`, and `update --message` fill it. `update ... closed` warns while it holds only template text. |

Text in `<...>` on its own is template text. [[sisyphus-show]] hides a section that holds only
template text, and [[sisyphus-edit]] `--append` replaces it. Change any section with
[[sisyphus-edit]], or open the file with [[sisyphus-open]].

## What sisyphus checks

The commands of sisyphus refuse to break these rules, and [[sisyphus-lint]] reports a break that a
hand edit makes:

- `state` matches the directory, and the full name is in one state directory only.
- `priority` and `effort` hold valid values, and `resolution` is empty or valid.
- Each frontmatter line is one `key: value` field that sisyphus can read.
- No tag has a space.
- `parent` and each `depends-on` entry name one existing issue, with no cycle.
- A closed issue is not moved back without `--reopen`. A new name is valid and not in use.
- Every wikilink in the repo resolves.

The rest is a convention that sisyphus does not check, for example: an open issue has no owner, an
approver accepts the close, and the Resolution section is complete. `update` warns about some of
them, for example a close with sub-issues that are not closed.

## Changing the template

`sisyphus init` writes `issues/TEMPLATE.md`, and `sisyphus new` copies it for each new issue. A repo
can change it:

- **Comments**: a frontmatter line that starts with `#` is a comment for the people who read the
  template. `new` leaves it out of the issue. A comment after a value stays in the issue.
- **Sections**: add, remove, rename, or reword any section and its template text. Keep a Context
  section for `new --context`, and a Resolution section for `close -m`, `abandon -m`, and
  `update --message`: without them, these flags are an error.
- **Optional fields**: `owner`, `approver`, `bookmark`, `workspaces`, `metadata`, `deferred-from`,
  `parent`, `depends-on`, and `remote` can go. A command that sets one of them adds the field to the
  issue when it needs it.
- **Fields of your own**: a field that sisyphus does not know, for example `team: core`, stays in
  each issue as it is, and lint accepts it. `show`, `list`, and their JSON do not print it.

`new` always sets `title`, `state`, `resolution`, `priority`, `effort`, `tags`, and `created`, from
its flags and their defaults, so the values of these fields in the template have no effect. Keep the
fields in the template so that they have a place and a comment. `sisyphus init --force` writes the
default template again.
