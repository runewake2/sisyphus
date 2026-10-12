#!/bin/sh
# The command of the role dependency-check/github-actions: each action that a workflow or a
# composite action uses must be pinned to a full commit SHA, because a tag or a branch can move to
# other code. With zizmor installed, it also runs its security checks.
NAME="GitHub Actions"
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
# The list of images or actions of one file, outside the working copy: a file in it is a change.
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
only_if_changed '.github/workflows/*.yml' '.github/workflows/*.yaml' action.yml action.yaml
files=$(hephaestus files '.github/workflows/*.yml' '.github/workflows/*.yaml' action.yml action.yaml)
if [ -z "$files" ]; then
	echo "No GitHub Actions workflow or action: nothing to check."
	exit 0
fi
for file in $files; do
	awk '/^[[:space:]-]*uses:[[:space:]]*/ { sub(/^[[:space:]-]*uses:[[:space:]]*/, ""); gsub(/["'"'"']/, ""); sub(/[[:space:]]+#.*$/, ""); print NR, $1 }' "$file" > "$tmp"
	while read -r line uses; do
		case "$uses" in
		./* | docker://*@sha256:*) continue ;; # an action of this repo, or an image with a digest
		docker://*)
			echo "$file:$line: $uses is not pinned to a digest (@sha256:...)."
			fail
			continue
			;;
		esac
		ref=${uses##*@}
		if [ "$ref" = "$uses" ] || ! printf '%s' "$ref" | grep -Eqx '[0-9a-f]{40}'; then
			echo "$file:$line: $uses is not pinned to a full commit SHA. Pin it, and keep the version in a comment: uses: ${uses%@*}@<sha> # ${ref}"
			fail
		fi
	done < "$tmp"
	done
if command -v zizmor >/dev/null 2>&1; then
	# shellcheck disable=SC2086 # files is a list of paths.
	zizmor --min-severity high $files || fail
else
	echo "zizmor is not installed: the other security checks of the workflows are not run (information only)."
fi
finish
