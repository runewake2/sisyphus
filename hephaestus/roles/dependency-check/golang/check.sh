#!/bin/sh
# The command of the role dependency-check/golang: govulncheck and go mod verify on each Go module,
# and the direct dependencies that have a newer version.
NAME=Go
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
only_if_changed '*.go' go.mod go.sum
modules=$(hephaestus files go.mod)
if [ -z "$modules" ]; then
	echo "No Go module (go.mod): nothing to check."
	exit 0
fi
need go "Install Go: https://go.dev/dl/"
if command -v govulncheck >/dev/null 2>&1; then
	vulncheck="govulncheck"
else
	vulncheck="go run golang.org/x/vuln/cmd/govulncheck@v1.1.4"
fi
for mod in $modules; do
	dir=$(dirname "$mod")
	echo "== $dir"
	# shellcheck disable=SC2086 # vulncheck is a command line.
	(cd "$dir" && $vulncheck ./...) || fail
	(cd "$dir" && go mod verify) || fail
	updates=$(cd "$dir" && go list -m -u -f '{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}' all 2>/dev/null)
	if [ -n "$updates" ]; then
		echo "Newer versions of direct dependencies (information only):"
		echo "$updates"
	fi
done
finish
