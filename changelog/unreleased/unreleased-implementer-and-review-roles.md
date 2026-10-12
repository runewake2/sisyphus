# Unreleased: The implementer, review, and formatter roles, and families of roles

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: New hephaestus roles: `implementer` does delegated issues, `code-reviewer` and `test-engineer` review a bookmark, and the formatter family (`formatter/golang`, `csharp`, `yaml`, `json`, `markdown`) formats the changed files of each language; roles of a family live in `roles/<family>/<variant>/`.

## Summary

A review of the government found no role that does ordinary issues, and no role that judges the code
or the tests of a change. Four roles fill these gaps.

## Changes

- `hephaestus/roles/implementer.md` (AI): does the work of one delegated issue on its own bookmark,
  with tests, docs, and an unreleased entry, and broadcasts `bookmark-ready`.
- `hephaestus/roles/code-reviewer.md` (AI): reviews the code of a bookmark for correctness, tests,
  fit, and scope. Each finding has a concrete case and a severity. It does not change the code, and
  it is normally a required reviewer in a referendum.
- `hephaestus/roles/test-engineer.md` (AI): maps each acceptance criterion to the test that shows
  it, and writes the missing tests on its own bookmark.
- Roles of a family live in a directory: `roles/<family>/<variant>/role.md` is the role
  `<family>/<variant>`, and its directory holds its implementation, which the command finds in
  `HEPHAESTUS_ROLE_DIR` ([[0005-roles-of-a-family-live-in-a-directory]]).
- The formatter family, in `hephaestus/roles/formatter/` (tool roles):
  - `formatter/golang` runs `gofmt`;
  - `formatter/csharp` runs `dotnet format whitespace` in folder mode;
  - `formatter/yaml`, `formatter/json`, and `formatter/markdown` run prettier: an installed one, or
    3.3.3 through `npx`.

  Each formats only the files that the work changes. Each ends with exit code 0 when no file of its
  language changed. In a referendum, each votes `yes`, or `modify` with the formatted files.
- `formatter/json` leaves out the raw benchmark results in `design/ollama-local-models/results/`.
- A run of the formatters on the whole repo found Markdown whose meaning prettier would change.
  These docs are corrected so that they keep their meaning:
  - The lists of failure modes in `design/ollama-local-models/notes/` used the mode ids as list
    numbers, so Markdown showed them as 1, 2, 3, and their sub-items were not nested. They are now
    bullets, "Mode 19: ...".
  - The steps of [[agent-guide]] have fenced code blocks, in an example block with four backticks:
    prettier changed the indent of the list items, and two indented code blocks became text.
  - `formatter/markdown` does not format code blocks: prettier had changed the JSON and YAML
    examples of [[sisyphus-show]] and [[issue-format]], so they no longer showed the real output.
- `hephaestus files [--changed] [<pattern>...]` prints the files of the working copy, or the files
  that the work since the base changes, with jj or git.
- A tool reviewer gets the base of a referendum in `HEPHAESTUS_BASE`. At the gate, it rejects only
  what the modification makes worse. Formerly, it rejected every modification that left work for it
  in its own scope.
- Two reviewers no longer change the git worktrees at the same time. Formerly, a reviewer could
  remove the half-made worktree of another, and that gave `no-result`.
- `repo-check` finds the root of the working copy with git first, then jj, so it also runs in the
  scratch worktree of a referendum.
- A modification of a tool reviewer names its role, not its full command.
- [[agent-delegation#The government]] has the four roles, and the path of a piece of work uses
  them. Their Claude Code subagents are in `.claude/agents/`.

## Issues

- Created [[implementer-agent]], [[code-reviewer-agent]], [[test-engineer-agent]], and
  [[formatter-roles]].

## Decisions

- [[0005-roles-of-a-family-live-in-a-directory]]
