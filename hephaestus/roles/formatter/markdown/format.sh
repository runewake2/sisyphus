#!/bin/sh
# The command of the role formatter/markdown: prettier on the Markdown files that the work changes.
# The files that tools own are left out: sisyphus formats its issues, hephaestus generates the
# subagents, and a referendum record does not change after its result.
LANGUAGE=Markdown
PATTERNS="*.md"
OPTIONS="--prose-wrap preserve"
EXCLUDE="issues,.claude,referendums"
. "$HEPHAESTUS_ROLE_DIR/../prettier.sh"
