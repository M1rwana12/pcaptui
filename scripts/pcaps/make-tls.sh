#!/usr/bin/env bash
#
# Build tls.pcap, the fixture behind one measurement this program's behaviour
# rests on.
#
# Two packets on port 443: a Client Hello, and an Application Data record going
# the other way.
#
#   1  10.0.0.5 -> 10.0.0.44  TLS Client Hello
#   2  10.0.0.44 -> 10.0.0.5  TLS Application Data
#
# The records are written by hand - enough of a Client Hello for the dissector
# to recognise the stream and number it, and an Application Data record whose
# contents are not really encrypted anything. That is all this fixture is for:
# it is NOT a decryptable session, and it cannot be, because a real one needs a
# key log and a real handshake.
#
# What it pins down is the thing that decides how :streams tls behaves:
# `-z follow,tls,raw,0` over a stream tshark plainly knows about answers with a
# complete banner, `Node 0: :0`, no payload and exit status 0 - byte for byte
# the shape of "there is no such stream" - while `-z follow,tcp,raw,0` over the
# same stream returns the bytes. That is why TLS is a family the user asks for
# by name and never one chosen for them.
set -euo pipefail

cd "$(dirname "$0")"

for tool in text2pcap mergecap tshark; do
  if ! command -v $tool >/dev/null 2>&1; then
    echo "$tool is not on PATH; it ships with Wireshark" >&2
    exit 1
  fi
done

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Handshake record: type 22, TLS 1.0 record version, one Client Hello with a
# 32-byte random of As, no session id, one cipher suite, no compression, no
# extensions.
cat > "$work/c2s.txt" <<'PKT'
2026-01-01 12:00:00
000000  16 03 01 00 2a 01 00 00 26 03 03 41 41 41 41 41
000010  41 41 41 41 41 41 41 41 41 41 41 41 41 41 41 41
000020  41 41 41 41 41 41 41 41 41 41 00 00 02 13 01 01
000030  00 00 00
PKT

# Application data record: type 23, sixteen bytes of it.
cat > "$work/s2c.txt" <<'PKT'
2026-01-01 12:00:01
000000  17 03 03 00 10 53 45 43 52 45 54 50 41 59 4c 4f
000010  41 44 31 32 33
PKT

fmt='%Y-%m-%d %H:%M:%S'

text2pcap -q -t "$fmt" -4 10.0.0.5,10.0.0.44 -T 51000,443 "$work/c2s.txt" "$work/c2s.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.44,10.0.0.5 -T 443,51000 "$work/s2c.txt" "$work/s2c.pcap"

mergecap -w tls.pcap "$work/c2s.pcap" "$work/s2c.pcap"

echo "wrote $(pwd)/tls.pcap"
tshark -r tls.pcap
echo
echo "--- tls.stream"
tshark -r tls.pcap -T fields -e frame.number -e tcp.stream -e tls.stream
echo "--- follow,tls,raw,0 (no key log: expect an empty banner and exit 0)"
tshark -r tls.pcap -q -z follow,tls,raw,0
echo "--- follow,tcp,raw,0 (the same stream, with the bytes)"
tshark -r tls.pcap -q -z follow,tcp,raw,0
