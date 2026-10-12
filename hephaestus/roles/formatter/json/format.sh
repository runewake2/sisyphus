#!/bin/sh
# The command of the role formatter/json: prettier on the JSON files that the work changes.
# The raw results of the local model benchmark are left out: the harness writes them, and a
# formatted copy would only add whitespace changes to large files.
LANGUAGE=JSON
PATTERNS="*.json"
EXCLUDE="design/ollama-local-models/results"
. "$HEPHAESTUS_ROLE_DIR/../prettier.sh"
