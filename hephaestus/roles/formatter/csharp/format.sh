#!/bin/sh
# The command of the role formatter/csharp: dotnet format whitespace, in folder mode, on the C#
# files that the work changes.
set -euf
cd "$(hephaestus files --root)"
files=$(hephaestus files --changed '*.cs')
if [ -z "$files" ]; then
	echo "No C# file changed: nothing to format."
	exit 0
fi
command -v dotnet >/dev/null 2>&1 || { echo "dotnet is not installed. Install the .NET SDK 6 or later: https://dot.net"; exit 1; }
hephaestus files --changed -0 '*.cs' | xargs -0 dotnet format whitespace . --folder --include
echo "dotnet format whitespace checked: $(echo "$files" | wc -l | tr -d ' ') files"
