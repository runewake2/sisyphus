# 0029: sisyphus does not manage agents

- Status: Accepted. Its layout of hephaestus (an issues tracker and design documents of its own) is superseded by [[0002-one-tracker-and-one-design-log-for-both-projects]].
- Date: 2026-10-11
- Version: none (a decision about scope, with no code change)

## Context

A human asked for agents that each have one role, for example a merge queue, and that take
delegated issues and resolve them. A first plan put this work in sisyphus. The human decided that
sisyphus will never manage agents: it is only a tool to interact with issues. A separate worker
manager, hephaestus, depends on sisyphus and uses its issues to provision and manage agents
(`hephaestus/README.md`).

## Decision

- sisyphus creates, changes, finds, links, and checks issues. It does not know roles, agents,
  runners, or merge queues, and it never starts an agent.
- hephaestus does all agent-specific work and all worker management. It lives in `hephaestus/`,
  beside sisyphus, with its own issues tracker and design documents.
- hephaestus uses sisyphus only through its CLI or `sisyphus-mcp`, as any other client does.
- sisyphus adds work for hephaestus only when it is generic issue data that any client can use,
  for example metadata fields or a filter on `metadata` ([[filter-by-metadata]]). An agent role is
  recorded in the generic `metadata` field, not in a field for agents.
- The sisyphus features that help any agent use issues stay in sisyphus: claims, `SISYPHUS_OWNER`,
  the MCP server, and [[agent-guide]]. They are about issues, not about running agents.

## Alternatives considered

- **Agent roles and a runner in sisyphus.** One tool, but it would mix two jobs, and every user of
  the issue tracker would get the worker manager too.
- **A `delegate` field in sisyphus.** Simple to filter on, but it is a field for agents. The generic
  `metadata` field and a generic filter give hephaestus the same, with no meaning in sisyphus.

## Consequences

- The design and the issues of agent roles are in `hephaestus/`, not in the issues of sisyphus.
- A request for sisyphus that comes from hephaestus is checked against this decision: generic issue
  data only.
