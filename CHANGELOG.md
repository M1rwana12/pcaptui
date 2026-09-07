# Changelog

## [1.0.0] - 2026-09-07

First release under this name. `pcaptui` continues the codebase of
[termshark](https://github.com/gcla/termshark) by Graham Clark, which received
its last release in July 2022 and its last commit in November 2022. Graham died
in 2024.

Everything below is what changed between that last release and this one.

### Fixed

- **Column names were mangled against Wireshark 4.x.** `tshark -G
  column-formats` gained a third tab-separated field; the reader split on
  whitespace with a limit of two, which was right when the line held two fields
  and silently wrong once it held three. Every column title offered in the
  preferences carried its padding and the display filter field name along with
  it.

  Reported upstream by @rbalint as gcla/termshark#169 and fixed by @gilramir — a
  Wireshark core developer — in gcla/termshark#170, on 2025-10-20. That pull
  request was never merged, because there was no longer anyone to merge it. The
  fix here was written independently and reaches the same design; it also strips
  the carriage return `tshark` emits on Windows.

- **A context leaked whenever a capture file could not be read.** A cancellable
  context was built from the loader's context and then abandoned without being
  cancelled if `stat` failed, staying attached to the loader for the life of the
  session. Filter searches against a file that has rotated away hit this
  routinely.

- **Error text was passed to `Fprintf` as a format string** in seven places. Any
  percent sign in the message — in a display filter, a filename, a `tshark`
  diagnostic — was read as a verb and mangled the output. The message a user
  most needs to read was the one most likely to be destroyed.

- **The test suite had never passed on Windows.** One test set `TMPDIR` and
  expected `tshark` to honour it; Windows uses `TEMP`. CI ran on Linux only, so
  nothing ever said so.

### Added

- **Expert Information** — what the dissectors found wrong with a capture,
  grouped by severity. `:expert`, or the Analysis menu.

- **Protocol Hierarchy** — every protocol present, by packet and byte count.
  `:hierarchy`, or the Analysis menu.

  Both narrow to the display filter in force, and both say so when a filter
  matches nothing instead of showing an empty dialog. `tshark` carries these
  behind `-z`, which the program had used only for conversations and stream
  reassembly. Exposing more of `-z` was on Graham Clark's own list for the
  release he did not get to make.

- **TLS decryption** via `--tls-keylog <file>`, or `tls-keylog` in the config
  file — the key log written by browsers and other tools when `SSLKEYLOGFILE`
  is set. Asked for as gcla/termshark#145 in January 2023 and never answered.

  It was already reachable through the general `tshark-args` setting, but
  nothing connected that setting to TLS and nothing checked the path. `tshark`
  accepts a key log that does not exist, starts normally, exits zero, decrypts
  nothing and says nothing, so a mistyped path is indistinguishable from
  traffic whose keys you never had. `pcaptui` names the path and refuses to
  start — in the UI and when its output is piped alike.

### Changed

- **Eleven vulnerabilities reachable from this code are gone.** `x/net`,
  `antchfx/xpath` — an infinite loop reachable from `xmlquery.Parse`, which runs
  for every packet decoded — and `sirupsen/logrus` are upgraded; the rest are in
  the standard library and are fixed by building with a current Go toolchain.
  Capture files are untrusted input by definition, so this is not housekeeping.

- **CI runs on Linux, macOS and Windows**, with `go vet`, `gofmt` and
  `govulncheck` as gates. It previously built on Linux alone, on Go 1.19.

- Configuration, cache and log files move to a `pcaptui` directory, and the
  config file is `pcaptui.toml`. Settings are otherwise unchanged, so an old
  `termshark.toml` can be copied across as-is.

- The binary is `pcaptui`. Everything it does on the command line, it does the
  same way.
