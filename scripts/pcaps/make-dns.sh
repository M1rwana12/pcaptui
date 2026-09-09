#!/usr/bin/env bash
#
# Build dns.pcap, the fixture the DNS statistic is tested against.
#
# There is no DNS in either of the other captures here, and this machine has no
# capture driver to make one with, so the packets are written out by hand and
# assembled by text2pcap. Keeping the script means the fixture is readable and
# can be extended - a second rcode, a AAAA query - rather than being a binary
# nobody can change.
#
# Four packets: a lookup that succeeds and one that does not.
#
#   1  10.0.0.5  -> 10.0.0.53  A? example.test
#   2  10.0.0.53 -> 10.0.0.5   A  example.test = 93.184.216.34
#   3  10.0.0.5  -> 10.0.0.53  A? missing.test
#   4  10.0.0.53 -> 10.0.0.5   No such name
#
# A second apart, so the request-response time in `tshark -z dns,tree` is a
# round 1000 ms and a test can say so.
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

# Transaction 0x1234: query and answer for example.test.
cat > "$work/q1.txt" <<'PKT'
2026-01-01 12:00:00
000000  12 34 01 00 00 01 00 00 00 00 00 00 07 65 78 61
000010  6d 70 6c 65 04 74 65 73 74 00 00 01 00 01
PKT

cat > "$work/r1.txt" <<'PKT'
2026-01-01 12:00:01
000000  12 34 81 80 00 01 00 01 00 00 00 00 07 65 78 61
000010  6d 70 6c 65 04 74 65 73 74 00 00 01 00 01 c0 0c
000020  00 01 00 01 00 00 00 3c 00 04 5d b8 d8 22
PKT

# Transaction 0x5678: query and NXDOMAIN for missing.test.
cat > "$work/q2.txt" <<'PKT'
2026-01-01 12:00:02
000000  56 78 01 00 00 01 00 00 00 00 00 00 07 6d 69 73
000010  73 69 6e 67 04 74 65 73 74 00 00 01 00 01
PKT

cat > "$work/r2.txt" <<'PKT'
2026-01-01 12:00:03
000000  56 78 81 83 00 01 00 00 00 00 00 00 07 6d 69 73
000010  73 69 6e 67 04 74 65 73 74 00 00 01 00 01
PKT

fmt='%Y-%m-%d %H:%M:%S'

text2pcap -q -t "$fmt" -4 10.0.0.5,10.0.0.53 -u 40000,53 "$work/q1.txt" "$work/q1.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.53,10.0.0.5 -u 53,40000 "$work/r1.txt" "$work/r1.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.5,10.0.0.53 -u 40001,53 "$work/q2.txt" "$work/q2.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.53,10.0.0.5 -u 53,40001 "$work/r2.txt" "$work/r2.pcap"

mergecap -w dns.pcap \
  "$work/q1.pcap" "$work/r1.pcap" "$work/q2.pcap" "$work/r2.pcap"

echo "wrote $(pwd)/dns.pcap"
tshark -r dns.pcap
