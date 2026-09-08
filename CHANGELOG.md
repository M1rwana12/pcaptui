# Changelog

## [Unreleased]

### Fixed

- **The hex pane showed bytes that were not in the packet.** `tshark -x` prints
  each line twice — once as hex, once as text — and the reader scanned the
  whole line for "two hex digits and a space". Text in the second column
  matched too. A line whose text begins `00 OK` yielded a seventeenth byte on a
  sixteen-byte line, and everything after it in that packet was shifted by one,
  including the layer highlighting derived from the protocol tree.

  It struck exactly the protocols people open a hex pane to read: three of the
  seven packets in this project's own demo capture were affected, and ten of
  the ninety-two in the sample telnet capture.

  The hex column is now read by position rather than by pattern.

- **The conversations table showed one direction's figures under the totals.**
  tshark prints `<-`, `->` and then the total; the row was emitted in that
  order against headers reading Pkts, Bytes, Pkts A→B, …, Pkts B→A. A
  conversation carrying seven packets was reported as carrying zero, and
  sorting by packet count sorted by reverse-direction traffic.

- **A column too narrow for its value showed a different value, not a cut one.**
  The packet list simply stopped drawing at the column edge, so in an
  eighty-column terminal `192.0.2.10` appeared as `192.0.2.1` and
  `198.51.100.20` as `198.51.10` — both well-formed addresses, with nothing to
  say a digit was missing. A timestamp lost its last decimals the same way.
  Anyone reading the list at a size the columns did not fit was being shown
  plausible wrong answers.

  Values that do not fit now end in `…`. The marker costs one cell, which is
  the cheapest thing in the row.

- `f` was bound twice in the Misc menu, so Feature Request could never be
  reached. It moves to `r`.

- **Hiding every column stopped the program from starting.** With no visible
  column, gowid renders each child of the packet list as Max, finds no maximum
  height among them and panics — before the interface appears, so there was no
  way to undo it from inside the program:

  ```
  panic: All columns widgets were rendered Max, so there is no max height to use.
  ```

  Nothing in the interface prevents hiding the last column, so a user could
  write a configuration that made their own program refuse to start, with a
  message that says nothing about columns. It now falls back to the default
  columns and says why.

  This was found by driving the program with a deliberately broken
  configuration, using the same screenshot machinery that renders the images in
  the README.

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
