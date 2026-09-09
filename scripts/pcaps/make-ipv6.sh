#!/usr/bin/env bash
#
# Build ipv6.pcap, the fixture the IPv6 half of the Endpoints view is tested
# against.
#
# Every other capture here is IPv4, and this machine has no capture driver, so
# the packets are written out by hand and assembled by text2pcap. Keeping the
# script means the fixture can be extended rather than being a binary nobody
# can change.
#
# Two packets, one in each direction, with different payload sizes so that the
# sent and received columns cannot be confused with each other:
#
#   1  2001:db8::1 -> 2001:db8::2   "hello"      (5 bytes of payload)
#   2  2001:db8::2 -> 2001:db8::1   "hi"         (2 bytes of payload)
set -euo pipefail

cd "$(dirname "$0")"

for tool in text2pcap mergecap; do
  if ! command -v $tool >/dev/null 2>&1; then
    echo "$tool is not on PATH; it ships with Wireshark" >&2
    exit 1
  fi
done

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

printf 'hello' | od -A x -t x1 -v | sed '$d' > "$work/a.body"
printf 'hi'    | od -A x -t x1 -v | sed '$d' > "$work/b.body"

{ echo "2026-01-01 12:00:00"; cat "$work/a.body"; } > "$work/a.txt"
{ echo "2026-01-01 12:00:01"; cat "$work/b.body"; } > "$work/b.txt"

fmt='%Y-%m-%d %H:%M:%S'

text2pcap -q -t "$fmt" -6 2001:db8::1,2001:db8::2 -T 50000,80 "$work/a.txt" "$work/a.pcap"
text2pcap -q -t "$fmt" -6 2001:db8::2,2001:db8::1 -T 80,50000 "$work/b.txt" "$work/b.pcap"

mergecap -w ipv6.pcap "$work/a.pcap" "$work/b.pcap"

echo "wrote $(pwd)/ipv6.pcap"
tshark -r ipv6.pcap
