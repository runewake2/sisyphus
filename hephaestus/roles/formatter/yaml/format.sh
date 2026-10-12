#!/bin/sh
# The command of the role formatter/yaml: prettier on the YAML files that the work changes.
LANGUAGE=YAML
PATTERNS="*.yaml *.yml"
. "$HEPHAESTUS_ROLE_DIR/../prettier.sh"
