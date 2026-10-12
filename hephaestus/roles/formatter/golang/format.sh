#!/bin/sh
# The command of the role formatter/golang: gofmt on the Go files that the work changes.
set -euf
cd "$(hephaestus files --root)"
if [ -z "$(hephaestus files --changed '*.go')" ]; then
	echo "No Go file changed: nothing to format."
	exit 0
fi
command -v gofmt >/dev/null 2>&1 || { echo "gofmt is not installed. Install Go: https://go.dev/dl/"; exit 1; }
formatted=$(hephaestus files --changed -0 '*.go' | xargs -0 gofmt -l -w)
echo "gofmt formatted: ${formatted:-nothing}"
