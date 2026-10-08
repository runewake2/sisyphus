# sisyphus new

Create an issue from `issues/TEMPLATE.md`, in the directory of its state.

## Usage

```bash
sisyphus new <name> [flags]
```

`<name>` is the new issue's name: 2-6 lowercase words in kebab-case, for example `fix-login-bug`.
[[sisyphus-slug]] makes a valid, unique name from any text.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-t`, `--title` | made from the name | The title. `fix-login-bug` becomes "Fix login bug". |
| `-s`, `--state` | `open` | `open`, `in-progress`, or `closed`. |
| `-r`, `--resolution` | `completed` when closed | `completed` or `abandoned`. Only for a closed issue. |
| `-p`, `--priority` | `medium` | `critical`, `high`, `medium`, or `low`. |
| `-e`, `--effort` | `medium` | `small`, `medium`, `large`, or `x-large`. |
| `--tags` | none | Comma-separated tags. Built-in: `bug`, `feature`; any other tag works too. |
| `-b`, `--bookmark` | none | The jj bookmark of the work. |
| `--owner` | none | The person or agent working on the issue. |
| `--approver` | none | The person or agent who accepts the issue when it closes. |
| `--workspace` | none | A jj workspace where work on the issue happens. |
| `--parent` | none | The parent issue, if this is a sub-issue. See [[sisyphus-parent]]. |
| `--depends-on` | none | An issue that must close first. See [[sisyphus-depends-on]]. |
| `-d`, `--deferred-from` | none | The issue or bookmark that deferred this work. |
| `--remote` | none | The URL of an external ticket. See [[sisyphus-remote]]. |
| `--context` | none | Text that replaces the Context section's placeholder. |
| `--metadata` | none | `key=value` notes. Repeat the flag, or separate pairs with commas. |

## What it does

1. Checks the name: its shape, and that no Markdown file anywhere in the repo already has that
   name, ignoring case. A clash lists the files that have it.
2. Checks the values: state, priority, effort, and resolution must be valid; `--resolution` is
   only allowed for a closed issue; `--parent` and `--depends-on` must name existing issues and
   cannot create a cycle; `--remote` must be a URL.
3. Copies `issues/TEMPLATE.md` from the repo, drops its instruction comments (the `#` lines at the
   top of the frontmatter), and fills in the fields. Inline comments on field lines stay.
4. Sets `created` to today and, for a closed issue, `closed` to today.
5. Replaces `# <Title>` in the body with the title, and the Context placeholder with `--context`.
6. Writes the file to `issues/<state>/<name>.md`.

Value details:

- `--tags "bug, scheduler ,,ui"` is written as `[bug, scheduler, ui]`: spaces trimmed, empties dropped.
- `--parent` and `--depends-on` are written as quoted wikilinks, for example `"[[some-epic]]"`.
- `--deferred-from` takes an issue (name, `[[name]]`, or `#name`), written as a quoted wikilink, or
  a bookmark (any value with `/`), written as it is.
- `--metadata` values are written quoted, with keys in sorted order:
  `{note: "first pass", session-id: "abc123"}`. A value cannot contain a comma.

## Output

The path of the new file, for example `issues/open/fix-login-bug.md`.

Warnings:

- The issue is in-progress but has no bookmark.
- The parent issue, or an issue in `--depends-on`, is closed.

## Examples

```bash
sisyphus new fix-login-bug --tags bug --priority high
sisyphus new add-dark-mode --title "Add a dark mode" --tags feature,ui --effort large
sisyphus new split-login-form --parent add-dark-mode --depends-on fix-login-bug
sisyphus new retry-backoff --state in-progress --bookmark ai/retry-backoff --owner alice
sisyphus new "$(sisyphus slug "Crash when the cache is empty")" --tags bug \
  --remote https://github.com/acme/widgets/issues/42 --context "Steps to reproduce: ..."
```
