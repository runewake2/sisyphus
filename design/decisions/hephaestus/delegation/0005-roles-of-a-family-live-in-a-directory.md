# 0005: Roles of a family live in a directory

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-implementer-and-review-roles]]

## Context

The human asked for one formatter role for each language: Go, C#, YAML, JSON, and Markdown. More
will come. The formatters are to be defined in a directory, and the implementation for each
language is to be in a subdirectory named after the language. Until now, each role was one file,
`hephaestus/roles/<role>.md` ([[0002-each-role-is-a-markdown-file]]), and a tool role had only a
command line.

## Decision

- **A family is a directory.** `roles/<family>/<variant>/role.md` is the role file of the role
  `<family>/<variant>`, for example `roles/formatter/golang/role.md` is `formatter/golang`. The
  `role` field must be that path. A name is lowercase kebab-case words, with `/` between the parts.
- **Only `role.md` is a role file below the top.** `roles/<role>.md` stays a role. Below it, only a
  file named `role.md` is a role file. The other files of a family are not role files: for example a
  `README.md` of the family, a script, or a configuration file.
- **The directory of a role holds its implementation.** hephaestus gives a tool command the
  directory of its role file in `HEPHAESTUS_ROLE_DIR`, so its command can be
  `sh "$HEPHAESTUS_ROLE_DIR/format.sh"`. The script comes from the roles of the caller. This is also
  so in the scratch worktree of a referendum, so a bookmark under review cannot change the script
  of its reviewer.
- **Flat names where `/` cannot go.** A worker, a subagent file, and `{role}` in a runner use the
  name with `-` for `/`: `formatter-golang-1`, `.claude/agents/formatter-golang.md`.
  `HEPHAESTUS_ROLE`, `delegate`, and message addresses use the full name.
- **Tools work on the files of the change.** `hephaestus files --changed` prints the files that the
  work since the base changes, with jj or git: the commits from the fork point, and the changes
  that are not committed. The base is `HEPHAESTUS_BASE`, which a referendum sets, or `main`. A
  formatter formats only those files, so it never changes a file outside the change.

## Alternatives considered

- **`roles/formatter/golang.md`.** Fewer files, but then a language has no place for its script or
  configuration. Also, every Markdown file below the roles would be a role file.
- **`roles/golang/formatter.md`, a directory for each language.** This groups by language, not by
  role. The family of formatters, with its shared rules and its shared prettier script, would be in
  five places.
- **Scripts in the repo under review.** A bookmark could change the command of its own reviewer.
- **Format every file of the repo.** The first run would change files that the work did not touch.
  In a referendum, that is outside the scope of the change.

## Consequences

- A new language is a new directory with a `role.md` and a `format.sh`.
- Wikilinks to a family role name its path: `[[roles/formatter/golang/role]]`.
- A tool reviewer can be in a referendum for each language. With R2, one modification for each
  round, a change that each of N formatters modifies needs N+1 rounds.
