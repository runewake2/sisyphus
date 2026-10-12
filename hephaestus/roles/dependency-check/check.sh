#!/bin/sh
# The command of the role dependency-check: govulncheck and go mod verify on each Go module of the
# working copy, and the direct dependencies that have a newer version.
set -u
cd "$(hephaestus files --root)" || exit 1
if command -v govulncheck >/dev/null 2>&1; then
	vulncheck="govulncheck"
else
	vulncheck="go run golang.org/x/vuln/cmd/govulncheck@v1.1.4"
fi
status=0
for mod in $(hephaestus files 'go.mod'); do
	dir=$(dirname "$mod")
	echo "== $dir"
	# shellcheck disable=SC2086 # vulncheck is a command line.
	(cd "$dir" && $vulncheck ./...) || status=1
	(cd "$dir" && go mod verify) || status=1
	updates=$(cd "$dir" && go list -m -u -f '{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}' all 2>/dev/null)
	if [ -n "$updates" ]; then
		echo "newer versions of direct dependencies (information only):"
		echo "$updates"
	fi
done
exit $status
