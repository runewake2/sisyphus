# sisyphus edit

Replace the text of one section of an issue's body, or append to it. Use it to complete the
`Summary` of a new issue, to add a dated line to `Notes`, or to complete the `Resolution` when an
issue closes.

## Usage

```bash
sisyphus edit <name> <section> <text> [--append]
sisyphus edit <name> <section> - [--append]      # read the text from standard input
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]). `<section>` is the
heading of the section, matched case-insensitively, for example `resolution` or
`"acceptance criteria"`.

| Flag | Meaning |
|---|---|
| `-a`, `--append` | Add the text at the end of the section, after a blank line, instead of replacing the section. |

## Rules

- The text replaces everything between the section's heading and the next heading at the same or a
  higher level. The heading line, the frontmatter, and every other section do not change.
- With `--append`, the text goes after the section's current text, separated by a blank line. If
  the section holds only the template placeholder (`<...>`), the text replaces the placeholder.
- The text can have more than one line. sisyphus removes blank lines at its start and its end.
- Empty text is an error.
- A section that the issue does not have is an error that lists the sections it has. The title
  (`# ...`) is not a section: change the title in the frontmatter.
- `edit` works on an issue in any state, including a closed issue.

`sisyphus update <name> closed` warns when the `Resolution` section still holds the template
placeholder (see [[sisyphus-update]]).

## Output

The path of the issue's file.

## Examples

```bash
sisyphus edit fix-login-bug summary "Fix the login loop after a password reset."
sisyphus edit fix-login-bug notes "2026-10-09: the bug is in the session cache." --append
sisyphus update fix-login-bug closed --resolution abandoned
sisyphus edit fix-login-bug resolution "Abandoned. Login moved to SSO in [[sso-login]]."
git log -1 --format=%B | sisyphus edit fix-login-bug resolution -
```
