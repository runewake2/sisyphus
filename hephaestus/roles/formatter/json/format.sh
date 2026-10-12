#!/bin/sh
# The command of the role formatter/json: prettier on the JSON files that the work changes.
LANGUAGE=JSON
PATTERNS="*.json"
. "$HEPHAESTUS_ROLE_DIR/../prettier.sh"
