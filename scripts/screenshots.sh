#!/usr/bin/env bash
#
# Regenerate the screenshots in .github/assets from a fixed capture.
#
# Both the .txt and the .svg come from the same run: the text is what CI
# compares against, and the image is what the README shows. A picture that is
# produced by the command a test asserts on cannot quietly drift away from what
# the program does.
#
#   scripts/screenshots.sh          regenerate
#   scripts/screenshots.sh --check  fail if anything changed
#
set -euo pipefail

cd "$(dirname "$0")/.."

FIXTURE=scripts/pcaps/demo.pcap
OUT=.github/assets
SIZE=110x32

if ! command -v tshark >/dev/null 2>&1; then
  echo "tshark is not on PATH; screenshots need it to dissect the capture" >&2
  exit 1
fi

check=0
[ "${1:-}" = "--check" ] && check=1

bin=$(mktemp -d)/pcaptui
go build -o "$bin" ./cmd/pcaptui

shoot() { # name, extra args...
  local name=$1; shift
  "$bin" --screenshot "$OUT/$name" --screenshot-size "$SIZE" -r "$FIXTURE" "$@" >/dev/null
}

if [ "$check" = 1 ]; then
  tmp=$(mktemp -d)
  cp "$OUT"/screenshot-*.txt "$tmp/" 2>/dev/null || true
fi

shoot screenshot-packets

if [ "$check" = 1 ]; then
  for f in "$OUT"/screenshot-*.txt; do
    base=$(basename "$f")
    if ! diff -u "$tmp/$base" "$f" > /tmp/shotdiff 2>/dev/null; then
      echo "Screenshot $base changed:" >&2
      head -40 /tmp/shotdiff >&2
      echo >&2
      echo "If the change is intended, run scripts/screenshots.sh and commit the result." >&2
      exit 1
    fi
  done
  echo "Screenshots are up to date."
else
  echo "Regenerated:"
  ls -1 "$OUT"/screenshot-*
fi
