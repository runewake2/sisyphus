#!/usr/bin/env bash
# Pull an Ollama model through the sandbox HTTP proxy.
# Ollama's own pull fails here: it DNS-checks the blob CDN redirect locally,
# and the sandbox has no direct DNS. This writes the same on-disk layout.
# Usage: opull.sh <name>[:tag] ...
set -euo pipefail
MODELS="${OLLAMA_MODELS:?set OLLAMA_MODELS to the Ollama models directory}"
REG=https://registry.ollama.ai
for ref in "$@"; do
  name="${ref%%:*}"; tag=latest; [[ "$ref" == *:* ]] && tag="${ref#*:}"
  ns=library; repo="$name"; [[ "$name" == */* ]] && { ns="${name%%/*}"; repo="${name#*/}"; }
  mdir="$MODELS/manifests/registry.ollama.ai/$ns/$repo"
  mkdir -p "$mdir" "$MODELS/blobs"
  manifest="$(curl -fsSL --retry 3 -H 'Accept: application/vnd.docker.distribution.manifest.v2+json' "$REG/v2/$ns/$repo/manifests/$tag")"
  for d in $(jq -r '.config.digest, .layers[].digest' <<<"$manifest"); do
    f="$MODELS/blobs/${d/:/-}"
    if [[ -f "$f" ]] && [[ "$(sha256sum "$f" | cut -d' ' -f1)" == "${d#sha256:}" ]]; then continue; fi
    curl -fsSL --retry 5 -C - -o "$f.partial" "$REG/v2/$ns/$repo/blobs/$d" || curl -fsSL --retry 5 -o "$f.partial" "$REG/v2/$ns/$repo/blobs/$d"
    [[ "$(sha256sum "$f.partial" | cut -d' ' -f1)" == "${d#sha256:}" ]] || { echo "digest mismatch $d" >&2; exit 1; }
    mv "$f.partial" "$f"
  done
  printf '%s' "$manifest" > "$mdir/$tag"
  echo "pulled $ns/$repo:$tag"
done
