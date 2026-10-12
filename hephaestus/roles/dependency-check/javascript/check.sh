#!/bin/sh
# The command of the role dependency-check/javascript: the audit of the package manager of each
# JavaScript project, by its lockfile. A vulnerability of high or critical severity fails the check.
NAME=JavaScript
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
only_if_changed package.json package-lock.json npm-shrinkwrap.json yarn.lock pnpm-lock.yaml
projects=$(hephaestus files package.json)
if [ -z "$projects" ]; then
	echo "No JavaScript project (package.json): nothing to check."
	exit 0
fi
for project in $projects; do
	dir=$(dirname "$project")
	echo "== $dir"
	if [ -f "$dir/pnpm-lock.yaml" ]; then
		need pnpm "Install it: npm install --global pnpm"
		(cd "$dir" && pnpm audit --audit-level high) || fail
		(cd "$dir" && pnpm outdated) || true
	elif [ -f "$dir/yarn.lock" ]; then
		need yarn "Install it: npm install --global yarn"
		case "$(cd "$dir" && yarn --version)" in
		1.*)
			# yarn 1 exits with a bit mask of the severities that it found: 8 is high, 16 is critical.
			(cd "$dir" && yarn audit --level high)
			[ $? -ge 8 ] && fail
			;;
		*) (cd "$dir" && yarn npm audit --all --recursive --severity high) || fail ;;
		esac
	elif [ -f "$dir/package-lock.json" ] || [ -f "$dir/npm-shrinkwrap.json" ]; then
		need npm "Install Node.js: https://nodejs.org/"
		(cd "$dir" && npm audit --audit-level=high) || fail
		echo "Newer versions (information only):"
		(cd "$dir" && npm outdated) || true
	else
		echo "$dir has no lockfile, so its dependencies cannot be audited. Make one with: npm install --package-lock-only"
		fail
	fi
done
finish
