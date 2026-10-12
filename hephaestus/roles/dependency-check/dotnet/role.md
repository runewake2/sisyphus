---
role: dependency-check/dotnet
kind: tool
command: sh "$HEPHAESTUS_ROLE_DIR/check.sh"
trigger: [request, schedule]
on-failure: security-researcher
listens: [bookmark-ready]
can-call-referendum: false
---

# Dependency check: .NET

## Charter

Check the dependencies of the .NET solutions and projects (in any .NET language) of the repo for known vulnerabilities and for versions that
can change under the repo.

It fails on the vulnerable NuGet packages, also transitive ones. It is a tool role: it uses no model. It follows the rules of every dependency
check ([[roles/dependency-check/README|the dependency check roles]]).

## Inputs

- A task message or a schedule: it checks the whole repo.
- A `bookmark-ready` notice: it checks only when the change touches a .NET dependency file.
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

`check.sh` in the directory of this role file: `dotnet restore`, `dotnet list package --vulnerable --include-transitive`, and `--outdated`. It needs network access to the NuGet sources of the projects.

## Ask a human

- When a vulnerability has no fixed version.
- When an update of a dependency changes its major version.
- When the check cannot run: the tool is not installed, or the network blocks its database.
