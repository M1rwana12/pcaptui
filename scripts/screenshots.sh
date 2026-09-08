#!/usr/bin/env bash
#
# Regenerate the screenshots in .github/assets from a fixed capture.
#
# Both the .txt and the .svg come from the same run: the text is what CI
# compares against, and the image is what the README shows. A picture that is
# produced by the command a test asserts on cannot quietly drift away from what
# the program does.
#
#   scripts/screenshots.sh          regenerate, overwriting the committed files
#   scripts/screenshots.sh --check  compare only; never touches them
#
set -euo pipefail

cd "$(dirname "$0")/.."

FIXTURE=scripts/pcaps/demo.pcap
OUT=.github/assets
SIZE=130x34

if ! command -v tshark >/dev/null 2>&1; then
  echo "tshark is not on PATH; screenshots need it to dissect the capture" >&2
  exit 1
fi

check=0
[ "${1:-}" = "--check" ] && check=1

# The committed screenshots are the Linux rendering. gowid picks its frame
# characters from runtime.GOOS - light box drawing on Windows, heavy elsewhere -
# so regenerating on another platform produces a file that will never match in
# CI. Download the screenshots-linux artifact from a CI run instead.
if [ "$(uname -s)" != "Linux" ]; then
  echo "warning: the committed screenshots are the Linux rendering." >&2
  echo "         gowid draws different frame characters on $(uname -s), so what" >&2
  echo "         this produces will not match CI. Use the screenshots-linux" >&2
  echo "         artifact from a CI run to regenerate." >&2
fi

# Checking writes somewhere else entirely. A check that overwrites the files it
# is checking leaves the working tree dirty on any machine whose rendering
# differs - which is every machine that is not Linux.
work=$(mktemp -d)
dest=$OUT
[ "$check" = 1 ] && dest=$work

bin=$work/pcaptui
go build -o "$bin" ./cmd/pcaptui

shoot() { # name, extra args...
  local name=$1; shift
  "$bin" --screenshot "$dest/$name" --screenshot-size "$SIZE" -r "$FIXTURE" "$@" >/dev/null
}

# The comparison ignores everything from the Info column rightwards.
#
# That text comes from tshark's dissectors, and differs between Wireshark
# releases - Ubuntu's tshark says "[TCP segment of a reassembled PDU]" where a
# newer one does not. Asserting on it would mean the gate breaks whenever a
# runner updates Wireshark, which is noise, not a regression in this program.
# What is checked is the part pcaptui draws: the title bar, the filter box, the
# panes, the column layout, and the values in every column up to Info.
normalise() {
  cut -c1-65 "$1" | sed 's/[[:space:]]*$//'
}

shoot screenshot-packets
shoot screenshot-expert --screenshot-keys ':expert
'

if [ "$check" = 0 ]; then
  echo "Regenerated:"
  ls -1 "$OUT"/screenshot-*
  exit 0
fi

status=0
for new in "$work"/screenshot-*.txt; do
  base=$(basename "$new")
  old=$OUT/$base

  if [ ! -e "$old" ]; then
    echo "Screenshot $base is new and has not been committed." >&2
    status=1
    continue
  fi

  if ! diff -u <(normalise "$old") <(normalise "$new") > "$work/diff"; then
    echo "Screenshot $base changed:" >&2
    head -40 "$work/diff" >&2
    echo >&2
    echo "If the change is intended, run scripts/screenshots.sh and commit the result." >&2
    status=1
  fi
done

[ "$status" = 0 ] && echo "Screenshots are up to date."
exit "$status"
