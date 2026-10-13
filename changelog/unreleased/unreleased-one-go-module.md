# Unreleased: One Go module, with the commands side by side

- Date: 2026-10-13
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus, build
- Changelog: hephaestus moves into the root Go module: its command is `cmd/hephaestus`, beside `cmd/sisyphus` and `cmd/sisyphus-mcp`, and `go test ./...` at the root tests both tools.

## Summary

A human asked for the commands of the two tools to live beside one another
([[0003-one-go-module-with-the-commands-side-by-side]]).

## Changes

- `hephaestus/cmd/hephaestus` is now `cmd/hephaestus`, and `hephaestus/internal/<package>` is now
  `internal/hephaestus/<package>`. `hephaestus/go.mod` and `hephaestus/go.sum` are gone.
  `hephaestus/` keeps the README and the role files.
- `go install github.com/runewake2/sisyphus/cmd/hephaestus@latest` installs hephaestus.
- `gofmt -l .`, `go vet ./...`, and `go test ./...` at the root check both tools. CI now tests
  hephaestus too, and `repo-check` and the merge queue run one set of checks.
- A test fails when hephaestus imports sisyphus code, or sisyphus imports hephaestus code.
- The README of the repo, the README of hephaestus, CONTRIBUTING, the design, and the role files
  describe the new layout.

## Decisions

- [[0003-one-go-module-with-the-commands-side-by-side]]

## Issues

- Closed [[one-go-module]].
