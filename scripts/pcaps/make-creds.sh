#!/usr/bin/env bash
#
# Build creds.pcap, the fixture the Credentials view is tested against.
#
# One HTTP request carrying Basic authentication, written out by hand and
# assembled by text2pcap - no capture here contains a login, and this machine
# has no capture driver to record one with.
#
# The password is "password" and the user is "user", base64 as
# dXNlcjpwYXNzd29yZA==. It is a fixture, not a secret: the point is that
# pcaptui can tell you a capture is carrying one in the clear.
set -euo pipefail

cd "$(dirname "$0")"

for tool in text2pcap tshark; do
  if ! command -v $tool >/dev/null 2>&1; then
    echo "$tool is not on PATH; it ships with Wireshark" >&2
    exit 1
  fi
done

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

printf 'GET /secret HTTP/1.1\r\nHost: intranet.test\r\nAuthorization: Basic dXNlcjpwYXNzd29yZA==\r\n\r\n' \
  | od -A x -t x1 -v | sed '$d' > "$work/req.body"

{ echo "2026-01-01 12:00:00"; cat "$work/req.body"; } > "$work/req.txt"

text2pcap -q -t '%Y-%m-%d %H:%M:%S' -4 10.0.0.5,10.0.0.80 -T 51000,80 \
  "$work/req.txt" creds.pcap

echo "wrote $(pwd)/creds.pcap"
tshark -q -r creds.pcap -z credentials
