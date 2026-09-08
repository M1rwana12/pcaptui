# Changelog

## [Unreleased]

### Added

- **Expert Information and Protocol Hierarchy are tables you can act on.**
  Both were text in a copy dialog. The Expert table would tell you a capture
  contained a retransmission and leave you to find it yourself, which is the
  opposite of what the README claims it is for.

  Enter on a row now narrows the packet list to the packets that row is about,
  and writes the filter into the filter box so you can see it and adjust it.
  `tshark`'s expert tap reports how many times something happened and never
  which packets, so the row becomes a question about the packets instead —
  `_ws.expert.message == "…"`. A hierarchy row filters on its own protocol,
  which turns "what is in this file" into "show me that".

  The dialog is also a list rather than a block of text, so it scrolls. A
  capture with more distinct expert items than the terminal has rows previously
  produced a dialog whose bottom could not be reached.

  Copy mode no longer applies inside these two dialogs. Nothing documented
  promised it there, and the rows doing something is worth more.

- **The title bar says how many packets there are**, and whether a display
  filter is why the list is short: `7 packets`, `3 packets · filtered`,
  `40,000 packets · loading`. The count was previously not stated anywhere in
  the program — nothing distinguished a capture with seven packets from a
  filter that matched seven, and nothing said a number was still climbing.
  Wireshark's equivalent is the most-read part of its window.

- **The row under the cursor is amber**, not blue. `docs/Brand.md` had said so
  since the brand was written; the themes never did, so the document described
  an intention as a fact. Blue in this field means Wireshark, and the reason
  for choosing another colour applies to the program more than to the README.
  Black text on it rather than white: 8.3:1 against 2.5:1.

  Terminals with fewer than 256 colours cannot render it and keep the colour
  they had.

- **One key per analysis view**, the same key that now appears beside the entry
  in the Analysis menu: `p` capture file properties, `s` reassemble stream,
  `v` conversations, `e` expert information, `y` protocol hierarchy.

  None of them had a key. The menu was reachable by mouse, or by `esc` — which
  lands on Misc, not Analysis — and otherwise you had to know to type the
  command's full name at `:`. The `?` help named none of the five, so Expert
  Information, which the README calls the fastest way into a capture somebody
  just handed you, could be missed entirely. The help now lists them, generated
  from the bindings so the two cannot drift apart.

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

- **`:logs` and `:config` did not exist on Windows.** Both were compiled out
  there, along with their Misc menu entries, because the Unix versions run the
  user's pager inside a terminal widget and there is no pty to run it in. The
  README, the FAQ and the User Guide all tell the reader to run `:config` — so
  on one of the three supported platforms, three documents sent people to a
  command that answered "no such command", and a bug report could not include
  the log because there was no way to find it.

  Windows now reads the file and shows it in a scrollable dialog, headed by the
  path. Long logs show their end and say how many lines were left out.

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
