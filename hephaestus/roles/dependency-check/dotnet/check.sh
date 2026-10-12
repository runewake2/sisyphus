#!/bin/sh
# The command of the role dependency-check/dotnet: the vulnerable NuGet packages, also transitive
# ones, of each .NET solution, or of each project when the repo has no solution.
NAME=.NET
. "$HEPHAESTUS_ROLE_DIR/../lib.sh"
only_if_changed '*.sln' '*.csproj' '*.fsproj' '*.vbproj' packages.lock.json Directory.Packages.props Directory.Build.props nuget.config NuGet.Config global.json
targets=$(hephaestus files '*.sln')
[ -n "$targets" ] || targets=$(hephaestus files '*.csproj' '*.fsproj' '*.vbproj')
if [ -z "$targets" ]; then
	echo "No .NET solution or project: nothing to check."
	exit 0
fi
need dotnet "Install the .NET SDK 8 or later: https://dot.net"
for target in $targets; do
	echo "== $target"
	dotnet restore "$target" --verbosity quiet || { fail; continue; }
	report=$(dotnet list "$target" package --vulnerable --include-transitive 2>&1)
	echo "$report"
	case "$report" in *"has the following vulnerable packages"*) fail ;; esac
	echo "Newer versions (information only):"
	dotnet list "$target" package --outdated 2>&1 || true
done
finish
