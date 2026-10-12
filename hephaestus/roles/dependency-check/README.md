# Dependency check roles

The dependency check roles check the dependencies of the repo, one role for each ecosystem. Each is
a tool role: a program does its work, with no model.

| Role | Files | Tools |
|---|---|---|
| `dependency-check/golang` | `go.mod`, `go.sum` | `govulncheck`, `go mod verify` |
| `dependency-check/javascript` | `package.json` and its npm, yarn, or pnpm lockfile | the `audit` of the package manager |
| `dependency-check/dotnet` | `*.sln`, `*.csproj`, `*.fsproj`, `*.vbproj` | `dotnet list package --vulnerable` |
| `dependency-check/java` | `pom.xml`, `gradle.lockfile` | OSV-Scanner |
| `dependency-check/docker` | Dockerfiles, compose files | the versions of the images; trivy |
| `dependency-check/github-actions` | `.github/workflows/*.yml`, `action.yml` | the pins of the actions; zizmor |

The role files: [[roles/dependency-check/golang/role]], [[roles/dependency-check/javascript/role]],
[[roles/dependency-check/dotnet/role]], [[roles/dependency-check/java/role]],
[[roles/dependency-check/docker/role]], and [[roles/dependency-check/github-actions/role]].

## Layout

Each ecosystem has a directory named after it, with a `role.md` and a `check.sh`. `lib.sh` is
shared: it finds the root of the working copy, skips a notice that touches no dependency file of the
ecosystem, and ends with the status of the checks.

## Rules for every dependency check

- **Whole repo or the change.** A request or a schedule checks every project of the ecosystem. A
  `bookmark-ready` notice checks only when the change touches a dependency file of the ecosystem.
- **Fail closed.** A vulnerability, an unpinned version, a missing lockfile, a tool that is not
  installed, and a database that cannot be read each fail the check. A check that cannot run does
  not pass.
- **Fail on what is real.** Fail on a known vulnerability, and on a version that can change under
  the repo. A newer version is information only.
- **Nothing that is slow or heavy by default.** A check that pulls images or installs much is only on
  request, for example `HEPHAESTUS_SCAN_IMAGES=1` for the docker check.
- **No change.** A check reads the repo; it never updates a dependency. The implementer does, from
  an issue.
- **A failure goes to the security researcher** (`on-failure`), which decides what to do.

To add an ecosystem, add a directory with a `role.md` and a `check.sh`, and a row to the table.
