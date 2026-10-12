#!/bin/sh
# The command of the role dependency-check/docker: the base images of each Dockerfile and the images
# of each compose file must name a version, not latest. With trivy installed, it also checks the
# configuration, and, when HEPHAESTUS_SCAN_IMAGES=1, the vulnerabilities of each image.
NAME=Docker
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
# The list of images or actions of one file, outside the working copy: a file in it is a change.
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
only_if_changed Dockerfile '*.Dockerfile' Containerfile 'docker-compose*.yml' 'docker-compose*.yaml' 'compose*.yml' 'compose*.yaml'
dockerfiles=$(hephaestus files Dockerfile '*.Dockerfile' Containerfile)
composes=$(hephaestus files 'docker-compose*.yml' 'docker-compose*.yaml' 'compose*.yml' 'compose*.yaml')
if [ -z "$dockerfiles$composes" ]; then
	echo "No Dockerfile or compose file: nothing to check."
	exit 0
fi
images=""
# check <file> <line> <image>: an image must name a version; a digest is best.
check() {
	case "$3" in
	scratch | *'$'*) return ;; # no image, or one that a build argument gives
	esac
	name=${3%@*}
	case "$3" in
	*@sha256:*) ;;
	*)
		case "${name##*/}" in
		*:latest | *:*) ;;
		*) name="$name:latest" ;;
		esac
		case "$name" in
		*:latest)
			echo "$1:$2: the image $3 names no version (latest). Name a version, or pin a digest (@sha256:...)."
			fail
			return
			;;
		esac
		echo "$1:$2: the image $3 has no digest, so its content can change (information only)."
		;;
	esac
	images="$images $3"
}
for file in $dockerfiles; do
	# The names of the build stages (FROM ... AS <name>) are not images.
	stages=$(awk 'toupper($1) == "FROM" { for (i = 2; i < NF; i++) if (toupper($i) == "AS") print tolower($(i + 1)) }' "$file")
	awk 'toupper($1) == "FROM" { i = 2; while ($i ~ /^--/) i++; print NR, $i }' "$file" > "$tmp"
	while read -r line image; do
		if printf '%s\n' "$stages" | grep -qx "$(printf '%s' "$image" | tr 'A-Z' 'a-z')"; then
			continue
		fi
		check "$file" "$line" "$image"
	done < "$tmp"
	done
for file in $composes; do
	awk '/^[[:space:]]*image:[[:space:]]*/ { sub(/^[[:space:]]*image:[[:space:]]*/, ""); gsub(/["'"'"']/, ""); print NR, $1 }' "$file" > "$tmp"
	while read -r line image; do check "$file" "$line" "$image"; done < "$tmp"
	done
if command -v trivy >/dev/null 2>&1; then
	trivy config --exit-code 1 --severity HIGH,CRITICAL --quiet . || fail
	if [ "${HEPHAESTUS_SCAN_IMAGES:-}" = "1" ]; then
		for image in $images; do trivy image --exit-code 1 --severity HIGH,CRITICAL --quiet "$image" || fail; done
	fi
else
	echo "trivy is not installed: the configuration and the image vulnerabilities are not checked (information only)."
fi
finish
