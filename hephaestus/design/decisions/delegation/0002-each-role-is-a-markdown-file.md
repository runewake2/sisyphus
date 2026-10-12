# 0002: Each role is a Markdown file

- Status: Accepted
- Date: 2026-10-12
- Version: [[0.0.110]]

## Context

Q2 of [[agent-delegation]] asked where the roles are described: in files of the repo, in the files
of one runtime such as `.claude/agents/`, or in code.

## Decision

From the human, 2026-10-11: each role is a Markdown file of its own, `hephaestus/roles/<role>.md`.
Its frontmatter has what hephaestus needs to run it (`role`, `kind`, `command`, `trigger`,
`on-failure`, `can-call-referendum`), and its sections describe it to people and, for an AI role,
to the model. [[role-file]] describes the format; `hephaestus roles` checks it.

## Alternatives considered

- **`.claude/agents/` files.** One runtime only, and no place for a tool role.
- **Roles in code.** A change of a role would need a build, and people could not read the roles in
  the repo.

## Consequences

- A role changes with a change of the repo, and goes through the same review.
- A runtime, for example Claude Code subagents, can be made from the role file.
