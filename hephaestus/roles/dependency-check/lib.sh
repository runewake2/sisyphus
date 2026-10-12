#!/bin/sh
# Shared by the dependency-check roles. The script of a role sets NAME (the ecosystem, for a
# message) and sources this file from the directory of its role file.
set -u
cd "$(hephaestus files --root)" || exit 1
status=0

# only_if_changed <pattern>...: on a bookmark-ready notice, end with exit code 0 when the change
# touches no file that matches. A request or a schedule checks the whole repo.
only_if_changed() {
	if [ "${HEPHAESTUS_MESSAGE_KIND:-}" = "bookmark-ready" ] && [ -z "$(hephaestus files --changed "$@")" ]; then
		echo "The change touches no $NAME dependency file: nothing to check."
		exit 0
	fi
}

# need <command> <how to install>: end with exit code 1 when the command is not installed.
need() {
	command -v "$1" >/dev/null 2>&1 || {
		echo "$1 is not installed, so the $NAME dependencies cannot be checked. $2"
		exit 1
	}
}

# fail records a failed check, and the checks go on.
fail() { status=1; }

# finish ends with the status of the checks.
finish() {
	if [ "$status" -eq 0 ]; then
		echo "The $NAME dependency checks passed."
	else
		echo "A $NAME dependency check failed: see the output above."
	fi
	exit "$status"
}
