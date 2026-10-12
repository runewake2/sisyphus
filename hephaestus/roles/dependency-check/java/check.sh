#!/bin/sh
# The command of the role dependency-check/java: OSV-Scanner on the Maven and Gradle dependencies.
# Gradle needs a lockfile (gradle.lockfile) to give the exact versions.
NAME=Java
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
only_if_changed pom.xml build.gradle build.gradle.kts gradle.lockfile libs.versions.toml
poms=$(hephaestus files pom.xml)
locks=$(hephaestus files gradle.lockfile)
builds=$(hephaestus files build.gradle build.gradle.kts)
if [ -z "$poms$locks$builds" ]; then
	echo "No Maven or Gradle project: nothing to check."
	exit 0
fi
for build in $builds; do
	dir=$(dirname "$build")
	if [ ! -f "$dir/gradle.lockfile" ]; then
		echo "$dir has no gradle.lockfile, so its versions cannot be checked. Make one with: ./gradlew dependencies --write-locks"
		fail
	fi
done
if command -v osv-scanner >/dev/null 2>&1; then
	scanner="osv-scanner"
else
	need go "Install OSV-Scanner (https://google.github.io/osv-scanner/) or Go, to run it."
	scanner="go run github.com/google/osv-scanner/cmd/osv-scanner@v1.9.2"
fi
set --
for file in $poms $locks; do
	set -- "$@" --lockfile "$file"
done
if [ $# -gt 0 ]; then
	# shellcheck disable=SC2086 # scanner is a command line.
	$scanner "$@"
	case $? in
	0 | 128) ;; # 128: no package to scan
	*) fail ;;
	esac
fi
if [ -n "$poms" ] && command -v mvn >/dev/null 2>&1; then
	echo "Newer versions (information only):"
	for pom in $poms; do (cd "$(dirname "$pom")" && mvn -q versions:display-dependency-updates) || true; done
fi
finish
