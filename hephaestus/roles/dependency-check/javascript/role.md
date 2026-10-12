---
role: dependency-check/javascript
kind: tool
command: sh "$HEPHAESTUS_ROLE_DIR/check.sh"
trigger: [request, schedule]
on-failure: security-researcher
listens: [bookmark-ready]
can-call-referendum: false
---

# Dependency check: JavaScript

## Charter

Check the dependencies of the npm, yarn, and pnpm projects of the repo for known vulnerabilities and for versions that
can change under the repo.

It fails on the vulnerabilities of high or critical severity, and each `package.json` that has no lockfile, so that it cannot be audited. It is a tool role: it uses no model. It follows the rules of every dependency
check ([[roles/dependency-check/README|the dependency check roles]]).

## Inputs

- A task message or a schedule: it checks the whole repo.
- A `bookmark-ready` notice: it checks only when the change touches a JavaScript dependency file.
- A referendum ballot request: it votes with its exit code.

## Outputs

- Exit code 1, with each problem and its file, for what it fails on. Newer versions and
  information are in the output, and do not fail it.
- The issue of a task, closed or stopped from the exit code; for a notice, the result in the Notes
  of its issue.

## Works with

- `implementer`: closes the issue whose `bookmark-ready` notice, from `hephaestus watch`, starts it, and updates the dependencies.
- `security-researcher`: gets each check that fails, as its `on-failure` role.
- `economist/compute`: reads how long each check runs.

## Tools

`check.sh` in the directory of this role file: the `audit` of the package manager of each lockfile (`npm audit`, `yarn audit` or `yarn npm audit`, `pnpm audit`), and its `outdated`. It needs network access to the advisory service of the npm registry.

## Ask a human

- When a vulnerability has no fixed version.
- When an update of a dependency changes its major version.
- When the check cannot run: the tool is not installed, or the network blocks its database.
