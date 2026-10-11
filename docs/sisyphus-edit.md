# sisyphus edit

Edit an issue in your editor, or replace the text of one section of its body, or append to it. Use
the section form to complete the `Summary` of a new issue, to add a dated line to `Notes`, or to
complete the `Resolution` when an issue closes, also from a script or an agent.

## Usage

```bash
sisyphus edit <name>                             # open the file in your editor
sisyphus edit <name> <section> <text> [--append]
sisyphus edit <name> <section> - [--append]      # read the text from standard input
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]). `<section>` is the
heading of the section, matched case-insensitively, for example `resolution` or
`"acceptance criteria"`. Give either `<name>` alone, or `<name>`, `<section>`, and `<text>`.

## Flags

| Flag | Meaning |
|---|---|
| `-a`, `--append` | Add the text at the end of the section, after a blank line, instead of replacing the section. Only with `<section>` and `<text>`. |
| `--json` | Print the changed issue as JSON, as [[sisyphus-show]] `--json` does, with its path and the warnings. Only with `<section>` and `<text>`. |

## Behavior

### In your editor

- The editor is `$VISUAL`, or else `$EDITOR`. The value can have arguments, for example
  `code --wait` or `subl -w`. A graphical editor must wait until you close the file, or sisyphus
  checks the file before you edit it. Without `$VISUAL` and `$EDITOR`, the command fails.
- Without a terminal on standard input, the command fails at once, also for a graphical editor:
  `No terminal to run your editor in, ...`. An agent's shell tool and sisyphus-mcp have no terminal,
  and an editor there would wait for ever. Use the section form instead.
- When the editor exits, sisyphus checks the file as [[sisyphus-lint]] does for one issue, and
  warns about problems, for example a frontmatter line that it cannot read. It does not change the
  file.

### One section

- The text replaces everything between the section's heading and the next heading at the same or a
  higher level. The heading line, the frontmatter, and every other section do not change.
- With `--append`, the text goes after the section's current text, separated by a blank line. If
  the section holds only the template placeholder (`<...>`), the text replaces the placeholder.
- The text can have more than one line. sisyphus removes blank lines at its start and its end.
- Empty text is an error.
- Text with a heading at the level of the section or above, for example `## Notes` in a `##`
  section, is an error: it would split the section. A deeper heading (`###`) and a heading in a code
  block are allowed.
- A section that the issue does not have is an error that lists the sections it has. The title
  (`# ...`) is not a section: change it with `sisyphus update <name> --title` (see
  [[sisyphus-update]]).
- `edit` works on an issue in any state, including a closed issue.
- Text that starts with `-`, for example a Markdown list, needs `--` before the positional
  arguments, or `-` to read it from standard input. Without it, the text is read as a flag.

`sisyphus update <name> closed` warns when the `Resolution` section still holds the template
placeholder (see [[sisyphus-update]]).

## Output

In your editor: the output of the editor. Warnings go to stderr. A failed editor is an error.

With `--json`, the changed issue as [[sisyphus-show]] prints it with `--json`, plus `path` (the file)
and `warnings` (a list; the warnings do not go to stderr then). See [[0.0.72]].

The path of the issue's file.

## Examples

```bash
sisyphus edit fix-login-bug                # in $VISUAL or $EDITOR
VISUAL="code --wait" sisyphus edit web/auth/fix-login-bug
sisyphus edit fix-login-bug summary "Fix the login loop after a password reset."
sisyphus edit fix-login-bug notes "2026-10-09: the bug is in the session cache." --append
sisyphus update fix-login-bug closed --resolution abandoned
sisyphus edit fix-login-bug resolution "Abandoned. Login moved to SSO in [[sso-login]]."
git log -1 --format=%B | sisyphus edit fix-login-bug resolution -
sisyphus edit fix-login-bug "acceptance criteria" -- "- [ ] The login works after a reset."
```
