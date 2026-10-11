# Changelog

## [[0.0.106]] - 2026-10-11
- Draw only the issues in selected states with sisyphus graph --state.

## [[0.0.105]] - 2026-10-11
- Add sisyphus rename, which renames an issue or moves it to a directory and updates every link to it.

## [[0.0.104]] - 2026-10-11
- List all among the valid values of --state of list and search.

## [[0.0.103]] - 2026-10-11
- Document stdout, stderr, exit codes, and every JSON shape.

## [[0.0.102]] - 2026-10-11
- Give the MCP tools titles and annotations, and return JSON results as structured content.

## [[0.0.101]] - 2026-10-11
- Add the bookmark to the JSON rows of list and search.

## [[0.0.100]] - 2026-10-11
- Remove the nested name example from the MCP tools, and say that slug only names new issues.

## [[0.0.99]] - 2026-10-11
- Hide the dir argument of the MCP tools by default, and add --dir and --allow-dir to sisyphus-mcp.

## [[0.0.98]] - 2026-10-11
- Match all words of a search query in any order, and exact phrases in double quotes.

## [[0.0.97]] - 2026-10-11
- Suggest up to 3 existing issues when an issue name does not exist.

## [[0.0.96]] - 2026-10-11
- Stop the sisyphus run of an MCP call that is cancelled or takes longer than 60 seconds.

## [[0.0.95]] - 2026-10-11
- Refuse to start an editor without a terminal, and refuse sisyphus_edit without a section.

## [[0.0.94]] - 2026-10-11
- Return matched lines from search, leave template text out of show and search, and add --limit.

## [[0.0.93]] - 2026-10-11
- Add a guide with instructions to paste for repos that let AI agents use sisyphus.

## [[0.0.92]] - 2026-10-11
- Give MCP clients instructions that tell agents the sisyphus workflow.

## [[0.0.91]] - 2026-10-11
- Tell the agent in MCP errors when sisyphus and sisyphus-mcp have different versions.

## [[0.0.90]] - 2026-10-11
- Add the MCP tools sisyphus_start, sisyphus_close, and sisyphus_abandon.

## [[0.0.89]] - 2026-10-11
- Share claims between jj workspaces and git worktrees, and add sisyphus claims.

## [[0.0.88]] - 2026-10-11
- Write issue files atomically, and lock the repo while a command changes issues.

## [[0.0.87]] - 2026-10-11
- Refuse to start an issue that another owner has in progress, unless --take is given.

## [[0.0.86]] - 2026-10-11
- Take the owner of an issue that starts from SISYPHUS_OWNER, and add list --mine and search --mine.

## [[0.0.85]] - 2026-10-11
- Name git next to jj in the bookmark and workspace arguments of the MCP tools.

## [[0.0.84]] - 2026-10-11
- Replace sisyphus open with sisyphus edit <name>, which opens the issue in your editor.

## [[0.0.83]] - 2026-10-11
- Rewrite the README for users, with a first session and links that work on GitHub.

## [[0.0.82]] - 2026-10-11
- Document MCP client setup, results, an example session, and troubleshooting for sisyphus-mcp.

## [[0.0.81]] - 2026-10-11
- Add a page that describes the issue file and the template, and fail new --context without a Context section.

## [[0.0.80]] - 2026-10-11
- Give every command page the sections Usage, Flags, Behavior, Output, and Examples, and test them.

## [[0.0.79]] - 2026-10-11
- Add examples to the help of every command, and group the commands in the root help.

## [[0.0.78]] - 2026-10-11
- Add sisyphus completion for bash, zsh, fish, PowerShell, and Nushell, with issue names and flag values.

## [[0.0.77]] - 2026-10-11
- Stop assuming jj: warn about a missing bookmark only in a jj repo, and describe the template fields for git too.

## [[0.0.76]] - 2026-10-11
- Add `sisyphus open` to open an issue in your editor and check it after.

## [[0.0.75]] - 2026-10-11
- Shorten the title column of `sisyphus list` to fit the terminal; `--wide` keeps full titles.

## [[0.0.74]] - 2026-10-11
- Hide empty fields and template-only sections in `sisyphus show`; `--all` shows everything.

## [[0.0.73]] - 2026-10-11
- Add `start`, `close`, and `abandon`, and `update -m` for the Resolution section.

## [[0.0.72]] - 2026-10-11
- Add `--json` to the commands that change an issue.

## [[0.0.71]] - 2026-10-11
- Let `sisyphus new --title` make the issue name from the title.

## [[0.0.70]] - 2026-10-11
- Accept `--state all`, and suggest the nearest valid value for an invalid state, priority, effort, or resolution.

## [[0.0.69]] - 2026-10-11
- Tell a new user to run `sisyphus init` outside a repo, and the next step after `init`.

## [[0.0.68]] - 2026-10-11
- Add the global flag `-C <directory>` to run sisyphus on another repo.

## [[0.0.67]] - 2026-10-11
- Read each issue file once in `sisyphus lint`.

## [[0.0.66]] - 2026-10-11
- Skip the files that `.gitignore` ignores in the repo walk of `resolve`, `links`, `lint`, `new`, and `slug`.

## [[0.0.65]] - 2026-10-11
- Build the link maps only when a command looks up many links, so that `resolve` and `links` are as fast as before 0.0.58.

## [[0.0.64]] - 2026-10-11
- Find an issue by its full name without reading every state directory.

## [[0.0.63]] - 2026-10-11
- Read the files of many issues on a bounded pool of goroutines.

## [[0.0.62]] - 2026-10-11
- Read only the frontmatter of an issue when a command does not use its body.

## [[0.0.61]] - 2026-10-11
- Split frontmatter lines without a regular expression, so that `list`, `search`, and `lint` are about twice as fast in a large repo.

## [[0.0.60]] - 2026-10-11
- Refuse to draw a graph of more than 200 issues in `sisyphus graph`, and point to `--json` or `--mermaid`.

## [[0.0.59]] - 2026-10-11
- Find cycles in `sisyphus lint` with one pass over the graph, and report each cycle once.

## [[0.0.58]] - 2026-10-11
- Resolve wikilinks with maps instead of a scan of every file, so that `lint` at 20,000 issues takes 2.3 s instead of 39 s.

## [[0.0.57]] - 2026-10-11
- Add `sisyphus update --reopen` to reopen an issue whose close was a mistake.

## [[0.0.56]] - 2026-10-11
- Remove the empty subdirectories that a move of an issue leaves behind.

## [[0.0.55]] - 2026-10-11
- Store tags in lowercase, and compare tags without regard to case.

## [[0.0.54]] - 2026-10-11
- Find an issue by name without regard to case in every command.

## [[0.0.53]] - 2026-10-11
- Warn when `sisyphus depends-on --clear` removes nothing.

## [[0.0.52]] - 2026-10-11
- Fail `sisyphus list --parent` and `search --parent` for a name that gives no issue.

## [[0.0.51]] - 2026-10-11
- Warn when `sisyphus new --deferred-from` names no issue and is not a bookmark.

## [[0.0.50]] - 2026-10-11
- Refuse `sisyphus init` below a directory that already has a tracker, unless `--nested` is given.

## [[0.0.49]] - 2026-10-11
- Refuse text with a heading that would split the section in `sisyphus edit`.

## [[0.0.48]] - 2026-10-11
- Keep the line endings of an issue file when sisyphus writes it.

## [[0.0.47]] - 2026-10-11
- Read issue files that start with a UTF-8 byte order mark, and keep the mark.

## [[0.0.46]] - 2026-10-11
- Read metadata values with commas correctly, so that a later metadata change does not corrupt them.

## [[0.0.45]] - 2026-10-11
- Add `sisyphus lint` to check every issue and every wikilink in the repo, and run it in CI.

## [[0.0.44]] - 2026-10-11
- Add `--ready` to `sisyphus list` and `sisyphus search` for the issues that can start now.

## [[0.0.43]] - 2026-10-11
- Let `sisyphus new --depends-on` take more than one issue.

## [[0.0.42]] - 2026-10-11
- Add `sisyphus update --add-tags` and `--remove-tags` to change single tags.

## [[0.0.41]] - 2026-10-11
- Add `sisyphus update --title` to change the title in the field and in the first heading of the body.

## [[0.0.40]] - 2026-10-11
- Let `sisyphus update` change fields without a state, and keep the resolution and closed date of a closed issue when its state does not change.

## [[0.0.39]] - 2026-10-11
- Warn in `list`, `search`, and `show` about issue files that are not valid.

## [[0.0.38]] - 2026-10-11
- Warn when an issue is in progress while some of its `depends-on` issues are not closed.

## [[0.0.37]] - 2026-10-11
- Fail `sisyphus new` and `sisyphus update` when an open issue would get an owner or a bookmark.

## [[0.0.36]] - 2026-10-11
- Align the columns of the `list` and `links` tables for wide characters, for example CJK text.

## [[0.0.35]] - 2026-10-11
- Reject a tag that contains a space in `sisyphus new` and `sisyphus update --tags`.

## [[0.0.34]] - 2026-10-11
- Reject an empty title, or a title with a line break or another control character, in `sisyphus new`.

## [[0.0.33]] - 2026-10-11
- Fail `sisyphus search --section` for a section that no issue and no template has.

## [[0.0.32]] - 2026-10-11
- Pass free text that starts with `-` from sisyphus-mcp to sisyphus as text, not as a flag.

## [[0.0.31]] - 2026-10-09
- Add `sisyphus edit` and the `sisyphus_edit` MCP tool to replace or append to one section of an
  issue's body, and warn when an issue closes with no `Resolution` text.

## [[0.0.30]] - 2026-10-09
- Print `resolution`, `created`, and `closed` in the text output of `sisyphus show`.

## [[0.0.29]] - 2026-10-09
- Add [[ollama-local-models]], the report of the spike that tested local Ollama models as sisyphus
  task managers, with its harness and results, and the follow-up issues for `sisyphus-mcp`.

## [[0.0.28]] - 2026-10-08
- Draw the work path of `<name>` in `sisyphus graph` with heavy lines, from the issues to start
  now, and name those issues in the legend.

## [[0.0.27]] - 2026-10-08
- Color each issue linked directly to `<name>` in a colored `sisyphus graph` by its state.

## [[0.0.26]] - 2026-10-08
- Color `<name>` and its arrows bold in `sisyphus graph` on a terminal, and the rest gray, as an
  experiment; `--color auto|always|never` decides.

## [[0.0.25]] - 2026-10-08
- Show the issue state on a second line below the name in `sisyphus graph` boxes, so graphs are
  narrower.

## [[0.0.24]] - 2026-10-08
- Draw each issue that only `sisyphus graph --full` adds in a dashed box, and mark it
  `indirect` in `--json` and `--mermaid`.

## [[0.0.23]] - 2026-10-08
- Call an issue that is ready to start "available" instead of "leaf" in `sisyphus graph`, its
  `--json` field, and its `--mermaid` class.

## [[0.0.22]] - 2026-10-08
- Place each issue in `sisyphus graph` one column after the farthest issue that points to it, so
  arrows no longer turn back or run side by side.

## [[0.0.21]] - 2026-10-08
- Draw each issue subdirectory in `sisyphus graph` as a box around its issues, show only file
  names in issue boxes, and mark the focus and leaves with box line styles instead of emoji.

## [[0.0.20]] - 2026-10-08
- `sisyphus slug` spells letters in ASCII (é becomes e), drops filler words before it keeps 6 words,
  and always prints ASCII.

## [[0.0.19]] - 2026-10-08
- Read the state directories once per command and parse each issue at most once; link-following
  commands run hundreds of times faster on a repo with thousands of issues.

## [[0.0.18]] - 2026-10-08
- Let issues sit in subdirectories below the state directories, found by their full name or by their
  file name; an ambiguous file name fails and lists the full names.

## [[0.0.17]] - 2026-10-08
- Fix `sisyphus update` in a colocated jj repo: move an issue that git does not track with a plain
  rename instead of `git mv`.

## [[0.0.16]] - 2026-10-08
- Draw `sisyphus graph` as a real dependency graph: everything below an issue and the path above
  it, with leaves marked, every linked issue with `--full`, plus `--mermaid` output and a new JSON format.

## [[0.0.15]] - 2026-10-07
- Suggest `bug` and `feature` as built-in tags in the template and in `--tags` help.

## [[0.0.14]] - 2026-10-07
- Make sisyphus itself independent of any particular repo or issue: generic template, error
  examples, and tests.

## [[0.0.13]] - 2026-10-07
- `sisyphus --version` prints the version the binary was compiled from, in any directory.

## [[0.0.12]] - 2026-10-07
- Add `sisyphus-mcp`: an MCP server giving an agent every `sisyphus` command as a tool.

## [[0.0.11]] - 2026-10-07
- Add `sisyphus graph` to view a large task's family tree in the terminal.

## [[0.0.10]] - 2026-10-07
- Narrow `sisyphus init` to set up `issues/` only; replace `agent-session` with a generic
  `metadata` field.

## [[0.0.9]] - 2026-10-07
- Add the two GitHub Actions that finish github-source-of-truth: issue-to-pr and sync-to-github.
  Add `sisyphus slug` and `sisyphus new --context`; fix `sisyphus show` to include `remote` and
  `depends-on`.

## [[0.0.8]] - 2026-10-07
- Default a closed issue's resolution to `completed` instead of requiring `--resolution`.

## [[0.0.7]] - 2026-10-07
- Add `depends-on` and `sisyphus depends-on`, with cycle detection and a close-time warning, plus
  `--blocked` on `list`/`search`.

## [[0.0.6]] - 2026-10-07
- Add `sisyphus search` to find issues by frontmatter filters and content, with section-scoped
  results.

## [[0.0.5]] - 2026-10-07
- Add `sisyphus list` to list and filter issues; update AGENTS.md's claimed-issue check to use it.

## [[0.0.4]] - 2026-10-07
- Add `remote` to pin an issue to a GitHub issue or Jira ticket, via `sisyphus new --remote` or
  `sisyphus remote`.

## [[0.0.3]] - 2026-10-07
- Let `sisyphus update` change `priority`, `effort`, and `tags`.

## [[0.0.2]] - 2026-10-07
- Move issue files through `git mv` when the repo has a `.git` directory.

## [[0.0.1]] - 2026-10-07
- Add `sisyphus show` to print one issue's key fields and body.

## [[0.0.0]] - 2026-10-07
- Start the workflow: contributing guide, agent rules, issues, and the changelog.
