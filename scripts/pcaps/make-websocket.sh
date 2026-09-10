#!/usr/bin/env bash
#
# Build websocket.pcap, the fixture the WebSocket stream family is tested
# against.
#
# There is no WebSocket in any of the other captures here, and this machine has
# no capture driver to make one with, so the packets are written out by hand and
# assembled by text2pcap. Keeping the script means the fixture is readable and
# can be extended - a binary frame, a fragmented message - rather than being a
# binary nobody can change.
#
# Four packets: the HTTP Upgrade handshake, then one message each way.
#
#   1  10.0.0.5 -> 10.0.0.80  GET /chat with Upgrade: websocket
#   2  10.0.0.80 -> 10.0.0.5  HTTP/1.1 101 Switching Protocols
#   3  10.0.0.5 -> 10.0.0.80  WebSocket text "Hello", masked
#   4  10.0.0.80 -> 10.0.0.5  WebSocket text "Hi there", unmasked
#
# Packet 3 is masked because a client frame must be, and the point of following
# a WebSocket stream rather than the TCP stream underneath is that tshark hands
# back the message - "Hello" - while the TCP stream carries the mask, the
# handshake and the framing bytes. Measured: `-z follow,websocket,raw,0` on this
# capture returns 48656c6c6f, and `-z follow,tcp,raw,0` returns the whole
# handshake and the masked bytes.
#
# Each direction is written as ONE text2pcap input file. text2pcap starts its
# sequence numbers at zero per file, so two packets of the same direction in
# two files become a retransmission rather than a stream.
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

# hexdump in the offset-and-bytes form text2pcap reads.
tohex() {
  xxd -p -c 16 | sed 's/../& /g' | awk '{printf "%06x  %s\n", (NR-1)*16, $0}'
}

# --- client to server: the request, then the masked "Hello" frame.
{
  echo "2026-01-01 12:00:00"
  printf 'GET /chat HTTP/1.1\r\nHost: example.test\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n' | tohex
  echo
  echo "2026-01-01 12:00:02"
  # 81 = FIN + text, 85 = masked + length 5, then the 4-byte mask and the
  # masked payload. This is the example from RFC 6455 section 5.7.
  printf '\x81\x85\x37\xfa\x21\x3d\x7f\x9f\x4d\x51\x58' | tohex
} > "$work/c2s.txt"

# --- server to client: the 101, then the unmasked "Hi there" frame.
{
  echo "2026-01-01 12:00:01"
  printf 'HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=\r\n\r\n' | tohex
  echo
  echo "2026-01-01 12:00:03"
  printf '\x81\x08Hi there' | tohex
} > "$work/s2c.txt"

fmt='%Y-%m-%d %H:%M:%S'

text2pcap -q -t "$fmt" -4 10.0.0.5,10.0.0.80 -T 51000,80 "$work/c2s.txt" "$work/c2s.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.80,10.0.0.5 -T 80,51000 "$work/s2c.txt" "$work/s2c.pcap"

mergecap -w websocket.pcap "$work/c2s.pcap" "$work/s2c.pcap"

echo "wrote $(pwd)/websocket.pcap"
tshark -r websocket.pcap
echo
echo "--- follow,websocket,raw,0"
tshark -r websocket.pcap -q -z follow,websocket,raw,0
