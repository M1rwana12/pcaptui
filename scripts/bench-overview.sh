#!/usr/bin/env bash
#
# Measure what the Overview costs on this machine, and print the rate.
#
# The Overview opens by itself when a capture finishes loading, and it is one
# `tshark -z` pass over the whole file. That is fine for a capture of a few
# megabytes and it is not fine for a large one, so the question is where the
# line is - and the answer is a property of the machine, not of the program.
# This script is how the line gets chosen, and how it can be checked again on
# a platform nobody here has: the maintainer's box is Windows, where starting
# processes is dear, and the CI runners are Linux and macOS.
#
# It reports two numbers per size:
#
#   capinfos  - what the file's own properties cost. The Overview asks for
#               these first because they answer "when was this captured" for a
#               fraction of the price.
#   -z pass   - the three statistics plus the traffic-over-time histogram,
#               exactly the arguments pcaptui uses.
#
# Usage: scripts/bench-overview.sh [megabytes ...]      (default: 1 5 20 50)
set -euo pipefail

cd "$(dirname "$0")/.."

for tool in tshark mergecap capinfos; do
  if ! command -v $tool >/dev/null 2>&1; then
    echo "$tool is not on PATH; it ships with Wireshark" >&2
    exit 1
  fi
done

sizes=("$@")
if [ ${#sizes[@]} -eq 0 ]; then
  sizes=(1 5 20 50)
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

seed=scripts/pcaps/telnet-cooked.pcap
if [ ! -f "$seed" ]; then
  echo "$seed is missing" >&2
  exit 1
fi

# Grow a capture by merging it with itself, which doubles it each time. The
# packets repeat and their timestamps repeat with them; that costs nothing here,
# because what is being measured is how fast tshark reads and dissects bytes.
grow_to() { # megabytes -> path
  local want_mb=$1
  local cur="$work/cur.pcap"
  cp "$seed" "$cur"

  while true; do
    local bytes
    bytes=$(wc -c < "$cur")
    if [ "$bytes" -ge $((want_mb * 1024 * 1024)) ]; then
      break
    fi
    mergecap -w "$work/next.pcap" "$cur" "$cur"
    mv "$work/next.pcap" "$cur"
  done

  echo "$cur"
}

# seconds, to two places, around a command whose output is discarded.
timed() {
  local start end
  start=$(date +%s%N)
  "$@" >/dev/null 2>&1 || true
  end=$(date +%s%N)
  awk -v ns=$((end - start)) 'BEGIN { printf "%.2f", ns / 1000000000 }'
}

echo "tshark:   $(tshark -v | head -1)"
echo "platform: $(uname -s) $(uname -m)"
echo

printf '%8s  %10s  %10s  %10s  %12s\n' "size" "packets" "capinfos" "-z pass" "rate"
printf '%8s  %10s  %10s  %10s  %12s\n' "----" "-------" "--------" "-------" "----"

for mb in "${sizes[@]}"; do
  file=$(grow_to "$mb")
  bytes=$(wc -c < "$file")
  actual_mb=$(awk -v b="$bytes" 'BEGIN { printf "%.0f", b / 1048576 }')
  packets=$(capinfos -c -M "$file" | awk '/Number of packets/ { print $NF }')

  info=$(timed capinfos -M "$file")

  # The arguments pcaptui's Overview uses, from pkg/stats: the protocol
  # hierarchy, the expert information, both endpoint families, and the
  # histogram whose interval the program picks from the capture's duration.
  pass=$(timed tshark -r "$file" -q \
    -z io,phs -z expert -z endpoints,ip -z endpoints,ipv6 -z io,stat,1)

  rate=$(awk -v mb="$actual_mb" -v s="$pass" 'BEGIN { if (s > 0) printf "%.1f MB/s", mb / s; else print "-" }')

  printf '%7sM  %10s  %9ss  %9ss  %12s\n' "$actual_mb" "$packets" "$info" "$pass" "$rate"
done
