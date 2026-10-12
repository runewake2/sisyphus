# Unreleased: Dependency checks for each ecosystem

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: `dependency-check` is a family of tool roles, one for each ecosystem: `golang`, `javascript`, `dotnet`, `java`, `docker`, and `github-actions`.

## Summary

The dependency check is a family now. The Go check moved to `dependency-check/golang`, and five
ecosystems joined it. Each role fails closed, and a failure goes to `security-researcher`.

## Changes

- `dependency-check/golang`: `govulncheck` and `go mod verify`, as before.
- `dependency-check/javascript`: the `audit` of npm, yarn, or pnpm, from the lockfile of each
  project. It fails on high or critical severity, and on a `package.json` with no lockfile.
- `dependency-check/dotnet`: `dotnet list package --vulnerable --include-transitive` for each
  solution, or each project.
- `dependency-check/java`: OSV-Scanner on each `pom.xml` and `gradle.lockfile`. A Gradle build with no
  lockfile fails, because its versions are not known.
- `dependency-check/docker`: each `FROM` and `image:` must name a version, not `latest`; an image
  with no digest is information. With trivy, `trivy config`, and `trivy image` on request.
- `dependency-check/github-actions`: each `uses:` must be pinned to a full commit SHA, and each
  `docker://` image to a digest. With zizmor, its security checks too.
- `lib.sh`, shared: on a `bookmark-ready` notice, a check runs only when the change touches a
  dependency file of its ecosystem. A request or a schedule checks the whole repo.
- The first run on this repo found two actions of the CI workflow that are not pinned:
  [[pin-github-actions]].

## Issues

- Updated [[dependency-check-roles]] (formerly `dependency-check-tool`).
- Created [[pin-github-actions]].
