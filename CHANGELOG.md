# Changelog

## [1.0.0] - 2026-09-07

First release.

### Analysis

- **Expert Information** — what Wireshark's dissectors found wrong with a
  capture: retransmissions, malformed packets, checksum failures, protocol
  violations, grouped by severity with the worst first. `:expert`, or the
  Analysis menu.

- **Protocol Hierarchy** — every protocol present, by packet and byte count, as
  a tree. `:hierarchy`, or the Analysis menu.

  Both narrow to the display filter in force, the way Wireshark does, and both
  say so plainly when the filter matches nothing instead of showing an empty
  dialog.

- **TLS decryption** via `--tls-keylog <file>`, or `tls-keylog` in the config
  file — the key log written by browsers and other tools when `SSLKEYLOGFILE`
  is set.

  `pcaptui` checks the file before starting and refuses to run if it is missing
  or unreadable, in the UI and when its output is piped alike. `tshark` accepts
  a key log path that does not exist, starts normally, exits zero, decrypts
  nothing and says nothing — so a mistyped path would otherwise be
  indistinguishable from traffic whose keys you never had.

### Correctness

- Column names read from `tshark -G column-formats` were mangled against
  Wireshark 4.x, which added a third tab-separated field to that output. Every
  column title offered in the preferences carried its padding and the display
  filter field name along with it. Now split on the tab that actually separates
  the fields, accepting either two fields or three so older `tshark` still
  works, and stripping the carriage return `tshark` emits on Windows.

- A cancellable context was built from the loader's context and then abandoned
  without being cancelled when a capture file could not be stat'd, staying
  attached to the loader for the life of the session. Filter searches against a
  file that has rotated away hit this routinely.

- Seven calls passed message text to `Fprintf` or `Warnf` as the format string.
  Any percent sign in an error — in a display filter, a filename, a `tshark`
  diagnostic — was read as a verb and destroyed the output, so the message a
  user most needs to read was the one most likely to be mangled.

### Security

- Eleven vulnerabilities reachable from this code are gone. `golang.org/x/net`,
  `github.com/antchfx/xpath` — an infinite loop reachable through
  `xmlquery.Parse`, which runs for every packet decoded — and
  `github.com/sirupsen/logrus` are upgraded; the remainder are in the standard
  library and are fixed by building with a current Go toolchain.

  Capture files are untrusted input by definition, so this is not housekeeping.

### Build and test

- Tests run on Linux, macOS and Windows, with `gofmt`, `go vet` and
  `govulncheck` as gates.

- Several tests drive a real `tshark` over a real capture, because the parts of
  this program that talk to `tshark` are exactly the parts that break when
  Wireshark changes, and a unit test cannot see that coming.

- Releases are built by goreleaser from a tag, which the workflow refuses to
  build unless it matches `version.go`.
