#!/usr/bin/env bash
#
# Build http2.pcap, the fixture the HTTP/2 stream family is tested against.
#
# Cleartext HTTP/2 - h2c - on TCP port 8080. Six packets, three each way: one
# TCP connection carrying one HTTP/2 stream, numbered 1.
#
#   1  10.0.0.5 -> 10.0.0.88  connection preface + SETTINGS
#   2  10.0.0.88 -> 10.0.0.5  SETTINGS
#   3  10.0.0.5 -> 10.0.0.88  HEADERS  (stream 1, POST /)
#   4  10.0.0.88 -> 10.0.0.5  HEADERS  (stream 1, 200)
#   5  10.0.0.5 -> 10.0.0.88  DATA     (stream 1, "hello from http2")
#   6  10.0.0.88 -> 10.0.0.5  DATA     (stream 1, "hi from server h2")
#
# No decode-as is needed: the dissector recognises the 24-byte connection
# preface, so it finds HTTP/2 on port 8080 and on port 80 alike - measured both
# ways, with identical dissection.
#
# What this fixture is for. An HTTP/2 connection multiplexes streams over one
# TCP connection, so following the TCP stream gives interleaved frame headers
# for every stream at once. Following the HTTP/2 stream gives one exchange, and
# tshark hands back its decoded HPACK headers as text before the DATA payload -
# `:method: POST` and then `hello from http2`, which is why the pane is worth
# offering.
#
# Each direction is written as ONE text2pcap input file, because text2pcap
# restarts its sequence numbers per file and two packets of one direction in
# two files come out as a retransmission rather than a stream.
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

# --- client to server
{
  echo "2026-01-01 12:00:00"
  # The 24-byte connection preface, then an empty SETTINGS frame: length
  # 000000, type 04, flags 00, stream 00000000.
  printf 'PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n\x00\x00\x00\x04\x00\x00\x00\x00\x00' | tohex
  echo
  echo "2026-01-01 12:00:02"
  # HEADERS on stream 1, END_HEADERS. The HPACK payload is 17 bytes:
  #   83         :method POST   (static table index 3)
  #   84         :path /        (index 4)
  #   86         :scheme http   (index 6)
  #   41 0c ...  :authority     (index 1, literal "example.test")
  printf '\x00\x00\x11\x01\x04\x00\x00\x00\x01\x83\x84\x86\x41\x0cexample.test' | tohex
  echo
  echo "2026-01-01 12:00:04"
  # DATA on stream 1, END_STREAM, sixteen bytes of it.
  printf '\x00\x00\x10\x00\x01\x00\x00\x00\x01hello from http2' | tohex
} > "$work/c2s.txt"

# --- server to client
{
  echo "2026-01-01 12:00:01"
  printf '\x00\x00\x00\x04\x00\x00\x00\x00\x00' | tohex
  echo
  echo "2026-01-01 12:00:03"
  # HEADERS on stream 1, END_HEADERS. Thirteen bytes of HPACK:
  #   88         :status 200    (index 8)
  #   5f 0a ...  content-type   (index 31, literal "text/plain")
  printf '\x00\x00\x0d\x01\x04\x00\x00\x00\x01\x88\x5f\x0atext/plain' | tohex
  echo
  echo "2026-01-01 12:00:05"
  # DATA on stream 1, END_STREAM, seventeen bytes.
  printf '\x00\x00\x11\x00\x01\x00\x00\x00\x01hi from server h2' | tohex
} > "$work/s2c.txt"

fmt='%Y-%m-%d %H:%M:%S'

text2pcap -q -t "$fmt" -4 10.0.0.5,10.0.0.88 -T 51000,8080 "$work/c2s.txt" "$work/c2s.pcap"
text2pcap -q -t "$fmt" -4 10.0.0.88,10.0.0.5 -T 8080,51000 "$work/s2c.txt" "$work/s2c.pcap"

mergecap -w http2.pcap "$work/c2s.pcap" "$work/s2c.pcap"

echo "wrote $(pwd)/http2.pcap"
tshark -r http2.pcap
echo
echo "--- tcp.stream and http2.streamid"
tshark -r http2.pcap -T fields -e frame.number -e tcp.stream -e http2.streamid
echo "--- follow,http2,raw,0,1"
tshark -r http2.pcap -q -z follow,http2,raw,0,1
