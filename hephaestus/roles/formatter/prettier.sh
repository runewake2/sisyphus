#!/bin/sh
# Shared by formatter/yaml, formatter/json, and formatter/markdown: prettier on the files that the
# work changes. The script of a role sets LANGUAGE, PATTERNS (for example "*.yaml *.yml"), and
# optionally OPTIONS (options of prettier) and EXCLUDE (comma-separated directories or patterns),
# and then sources this file.
set -euf
cd "$(hephaestus files --root)"
set -- --changed
if [ -n "${EXCLUDE:-}" ]; then
	set -- "$@" --exclude "$EXCLUDE"
fi
# shellcheck disable=SC2086 # PATTERNS is a list; set -f stops the shell from expanding its globs.
set -- "$@" $PATTERNS
if [ -z "$(hephaestus files "$@")" ]; then
	echo "No $LANGUAGE file changed: nothing to format."
	exit 0
fi
# An installed prettier first; else a fixed version through npx, so that every run formats the same.
if command -v prettier >/dev/null 2>&1; then
	prettier="prettier"
elif command -v npx >/dev/null 2>&1; then
	prettier="npx --yes prettier@3.3.3"
else
	echo "prettier is not installed: npm install --global prettier@3.3.3"
	exit 1
fi
# shellcheck disable=SC2086
formatted=$(hephaestus files -0 "$@" | xargs -0 $prettier --write --list-different ${OPTIONS:-})
echo "prettier formatted: ${formatted:-nothing}"
