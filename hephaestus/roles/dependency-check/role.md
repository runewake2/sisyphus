---
role: dependency-check
kind: tool
command: sh "$HEPHAESTUS_ROLE_DIR/check.sh"
trigger: [request, schedule]
on-failure: security-researcher
listens: [bookmark-ready]
can-call-referendum: false
---

# Dependency check

## Charter

Check the Go dependencies of the repo: known vulnerabilities that the code reaches, and modules
whose files do not match `go.sum`.

It runs `govulncheck` and `go mod verify` on each Go module of the repo, and lists the direct
dependencies that have a newer version. Only a vulnerability that the code calls, or a module that
fails its check, fails it. It is a tool role: it uses no model. It gives the security researcher a
cheap first pass, and the economists and the security researcher read its results.

## Inputs

- A task message, a schedule, or a `bookmark-ready` notice, with the workspace to check.
- The `go.mod` and `go.sum` of each module of the repo.

## Outputs

- The vulnerabilities that the code reaches, with their ids and fixed versions, and the modules that
  fail `go mod verify`: exit code 1.
- The direct dependencies that have a newer version: information only.
- The issue of a task, closed or stopped from the exit code, or for a notice, the result in its
  Notes.

## Works with

- `implementer`: sends the `bookmark-ready` broadcast that starts it.
- `security-researcher`: gets each check that fails, as its `on-failure` role.

## Tools

`check.sh` in the directory of this role file: `govulncheck` (an installed one, or v1.1.4 through
`go run`), `go mod verify`, and `go list -m -u`. They need network access to the Go vulnerability
database and the module proxy.

## Ask a human

- When a vulnerability has no fixed version.
- When an update of a dependency changes its major version.
