# 0003: One Go module, with the commands side by side

- Status: Accepted
- Date: 2026-10-13
- Version: [[unreleased-one-go-module]]

## Context

hephaestus was a Go module of its own in `hephaestus/`, with its command in `hephaestus/cmd/`, apart
from `cmd/sisyphus` and `cmd/sisyphus-mcp` (Q7 of [[agent-delegation]]). The human asked to combine
the cmd directories, so that the tools live beside one another.

A Go module cannot hold a directory outside its own tree, and Go lets only the code under
`hephaestus/` import `hephaestus/internal/`. So the commands can be side by side only in one module.

## Decision

- The repo has one Go module, `github.com/runewake2/sisyphus`. `hephaestus/go.mod` is gone.
- The commands are side by side: `cmd/sisyphus`, `cmd/sisyphus-mcp`, and `cmd/hephaestus`.
- The packages of hephaestus are in `internal/hephaestus/<package>`. Its docs and role files stay in
  `hephaestus/`.
- The boundary stays: hephaestus uses sisyphus only through its command, and sisyphus knows nothing
  of agents ([[0029-sisyphus-does-not-manage-agents]]). The module no longer keeps it, so a test
  does: `cmd/hephaestus` may import only `internal/hephaestus/...` of this module, and the commands
  of sisyphus may not import them.
- The checks are one: `gofmt -l .`, `go vet ./...`, and `go test ./...` at the root, also in CI,
  `repo-check`, and the merge queue.

## Alternatives considered

- **Two modules, with a go.work file.** The commands would still be in two cmd directories, and a
  workspace file is for local work, not for users who install a command.
- **Keep the packages in `hephaestus/internal/`.** Only code under `hephaestus/` could import them, so
  `cmd/hephaestus` could not.

## Consequences

- `go install github.com/runewake2/sisyphus/cmd/hephaestus@latest` installs hephaestus.
- CI builds and tests hephaestus too. Formerly it ran only the sisyphus module.
- The two tools share one `go.mod`: a dependency of one is in the module of both.
