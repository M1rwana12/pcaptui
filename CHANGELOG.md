# Changelog

## [Unreleased]

### Added

- **Credentials** — `a`, `:credentials`, or the Analysis menu. The logins
  `tshark` can read in the clear: HTTP basic authentication, FTP, POP, IMAP,
  SMTP and telnet. "This capture is carrying a password, here it is, and here
  is the packet" is a stronger answer than most of what the other views give.

  It is the only view whose rows name a packet, so Enter lands on that exact
  frame rather than on everything matching some text.

  It also ignores the display filter, and says so instead of pretending
  otherwise: `tshark` accepts a filter for this tap, exits successfully and
  reports the same logins anyway — measured with a filter that excludes every
  packet in the file.

- **The Overview says what the file is.** When it was captured, how long it
  covers, how many packets, how large, and whether a snapshot length cut the
  packets short. None of the three statistics answers *when*, and "this is a
  forty-second slice from 1999, not the hour you asked for" is often the whole
  answer — it used to be behind a separate key on a separate screen. It costs
  almost nothing next to the rest: `capinfos` took 0.27 s on a 44 MB capture
  where the statistics pass over the same file took 9.1 to 11.6 s.

- **And when the traffic happened**, as one row of blocks with `·` for an
  interval that carried nothing. A burst at the start, a hole in the middle or
  "all of it arrived in the last three seconds" is a shape, and it rides in the
  pass that was already running. The interval is chosen from the length of the
  capture, about twenty marks, rounded to a number a person would have picked.

### Changed

- **The Overview fills the screen it is given.** Each section was cut to four
  rows whatever the terminal, so a tall one showed four rows and then blank
  space above the Close button. The limit now comes from the height of the
  dialog, less what is drawn around the rows, and never falls below three.

- **Conversations opens on the busiest tab rather than on Ethernet.** On any
  routed capture the Ethernet tab is one row per next-hop MAC address — true,
  and not what anybody opened the view to see. The counts were already computed
  for the tab labels. A tie goes to the more specific tab, so a capture with
  one IPv4 conversation and one TCP conversation opens on the one that names
  ports; that rule is written down rather than left to arrival order, because
  `tshark` prints its sections in the reverse of the order they were asked for.

- **A statistic can be cancelled.** Every `-z` view is one pass over the whole
  capture — measured here at 1.07 s for 2.8 MB, 3.51 s for 22 MB and 11.6 s for
  44 MB, and the Overview starts one by itself when a file finishes loading.
  The wait dialog had no buttons at all, and Escape closed it without stopping
  anything: `tshark` went on reading and the result opened over whatever you
  had moved on to. It now says Cancel, means it, and treats Escape the same
  way.

- **Expert Information is readable on a real capture.** Wireshark reports some
  items once per occurrence with a counter in the text, and the tap reports
  each of those as a separate kind of problem: on a capture of 376,832 packets
  `tshark -z expert` printed 4,108 rows, of which 4,095 were
  `Duplicate ACK (#n)`. Thirteen facts arriving as four thousand lines, with
  the Overview summarising them as "and 4104 more" — which reads as four
  thousand distinct problems.

  Those rows now fold into one, counting `143,325` under
  `Duplicate ACK (#1-#4095)`, and Enter on it filters the packet list to every
  one of them. The Overview now says "and 10 more".

- **The per-severity totals are shown.** `tshark` states them in its section
  headings — `Notes (430012)` — and the program had been reading them and
  throwing the number away. The dialog now opens with
  `20,481 errors · 430,012 notes · 16,384 chats`.

- **Every table groups its digits the same way.** The Overview printed
  `frame 7 1027` six lines above `192.0.2.10 7 1,027` — one quantity, two
  spellings, in one dialog.

### Fixed

- **IPv6 hosts were invisible to Endpoints and to the Overview.** Both asked
  `tshark` for IPv4 endpoints only, so a capture of IPv6 traffic answered
  "Nothing to report" to the question "who is on the wire" — six lines below
  the protocol hierarchy in the same dialog listing `ipv6`. Both families are
  now asked for in the same pass, which measured as free, and a row's filter
  uses `ipv6.addr` where the address has a colon in it: Wireshark has no one
  field name covering both, and `ip.addr` against an IPv6 address is rejected
  outright.

## [1.1.0] - 2026-09-09

### Added

- **Overview** — `o`, or the Analysis menu, **and it opens by itself** when a
  capture finishes loading. Once per capture, never during a live capture, and
  `:set start-view packets` turns it off. The three questions you have about a
  capture somebody just handed you, on one screen: what is wrong with it, what
  is in it, who is on the wire. Each section shows its first few rows and says
  how many it left out; every row keeps the filter its own table would have
  given it, so the summary is not a dead end.

  Problems come first because that is what you are looking for, and because a
  capture with nothing wrong is worth knowing in one line rather than as a
  missing section.

  One `tshark` run, not three: `tshark` accepts several `-z` arguments and
  produces all of them from a single pass over the file.

- **Endpoints** — `t`, or the Analysis menu. Every IPv4 address in the capture
  with what it accounts for and how much of that it sent versus received,
  busiest first, because the question it answers is who is doing the most;
  `tshark` prints them in the order it met them, which buries the host you are
  looking for. Enter on a row filters to `ip.addr == …`.

  Adding it needed one `Stat` entry and a parser: the key, the menu entry and
  the `?` help all come from the same list, so none of them had to be told.

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
  `v` conversations, `e` expert information, `y` protocol hierarchy. `t`
  endpoints joined them later.

  None of them had a key. The menu was reachable by mouse, or by `esc` — which
  lands on Misc, not Analysis — and otherwise you had to know to type the
  command's full name at `:`. The `?` help named none of the five, so Expert
  Information, which the README calls the fastest way into a capture somebody
  just handed you, could be missed entirely. The help now lists them, generated
  from the bindings so the two cannot drift apart.

- **HTTP responses, by status.** `w`, `:http`, or HTTP in the Analysis menu:
  how many responses succeeded, how many were not found, how many the server
  failed on, each code under its class. Enter on a row filters the packet list
  to those responses — a code row gives `http.response.code == 404`, a class
  row the range it stands for. Only what is in the capture is listed; tshark
  prints a fixed skeleton of every status class it knows, so three responses
  would otherwise arrive as four useful lines under sixteen zeroes.

- **DNS, what was asked and how it came back.** `d`, `:dns`, or DNS in the
  Analysis menu: the lookups, the record types, the response codes and how long
  the server took. Enter on a row filters the packet list wherever a filter can
  say what the row means — one row cannot, and says so: tshark reads the
  response-code bits of every DNS header, including queries, which carry a
  zero, so its "No error" count includes the questions while the filter finds
  only the answers.

  It comes with `scripts/pcaps/dns.pcap` and the script that builds it. Neither
  existing capture contains DNS, and a fixture nobody can regenerate is one
  nobody can extend.

- **`--two-pass`**, and `two-pass` in the configuration file. Some of what
  `tshark` knows about a packet depends on the packets after it: a request
  cannot say which frame answered it until that frame has been read. Off by
  default, and measured before deciding — in two-pass mode `tshark` prints
  nothing until it has read the whole file, so on 376,000 packets the first
  packet took 2.7 seconds to appear instead of 0.36, growing with the file,
  for 8% more total time. Asked for on a live capture or a pipe, which cannot
  be read twice, `pcaptui` says so and carries on with one pass rather than
  leaving the missing fields to look like a capture that has none.

- **Export objects** — `x`, `:export`, or the Analysis menu. The files a
  capture carried — a page fetched over HTTP, a mail body, a block copied over
  SMB — written out to disk, with a count and the folder they went to. The
  kinds on offer come from your `tshark`, not from a list inside `pcaptui`, so
  they match the Wireshark you have. `tshark` says nothing about what it wrote
  and exits successfully either way, so `pcaptui` reads the folder before and
  after: "no http objects in this capture" is a real answer and looks different
  from having written some.

### Changed

- **A million packets carry 1.9 MB of colour instead of 115 MB.** Measured
  before and after on a million packets: 115 MB and 8,000,084 allocations
  became 1.9 MB and 49. Every packet stored its own pair of colour interfaces,
  built by parsing the hex strings tshark printed - per packet, though a
  capture only ever uses the handful of pairs its colour rules define. Each
  distinct pair is now parsed once and packets carry an index into it.

- **The packet-number lookups cost 3.8 MB per million packets instead of
  92.7 MB.** Two `map[int]int` held one entry per packet each: which table row
  shows a packet number, and which packet number follows a given one. Both are
  now answered by searching the packet numbers themselves, which the loader
  already reads in order — measured at 92.7 MB and 159 ms of building against
  3.8 MB and 0.9 ms.

### Fixed

- **The hex pane was empty for every reassembled packet.** `tshark` labels each
  data source when a packet has more than one — a reassembled TCP stream, a
  decrypted TLS record, a decompressed body — and it labels the first one too.
  The reader treated that first label as "a second source starts here" and
  skipped the frame's own bytes. Measured on a two-segment HTTP response: the
  first packet showed its 93 bytes and the second showed none.

- **Every IPv6 conversation was missing from the Conversations table.**
  `tshark` does not bracket IPv6, so one end reads `2001:db8::1:50000` — an
  address with four colons of its own. The code split on every colon and
  required exactly two pieces, so the row was dropped without a word. The tab
  said `TCP (0)` on a capture with one IPv6 conversation in it.

- **A statistic that failed said "Nothing to report for this capture."** The
  handler had no `OnError`, so the error was type-asserted away and lost —
  including `tshark`'s own explanation of which display filter field is wrong.
  The Overview opens by itself on every capture, so that sentence had become
  the program's answer to a failure it had been told about in detail.

- **`--tls-keylog`, `-d` and `main.tshark-args` reached only the packet list.**
  The four other loaders — statistics, stream reassembly, conversations, object
  export — each built their own `tshark` command line from scratch and carried
  none of them. So the key log decrypted the packet list while the reassembled
  stream stayed ciphertext, which is the opposite of what the User Guide
  promises; and `-d tcp.port==23,http` changed what the list said a packet was
  without changing what the Protocol Hierarchy counted it as. Measured: the
  hierarchy said `telnet 46 4670` before and `http 2 622` after.

- **Two columns of the Conversations table could not be sorted at all.** Start
  and Duration went through a comparator that is `ParseFloat`, and `tshark`
  writes `0,000000000` on a machine whose numbers use a decimal comma — as this
  one's does. Clicking the header did nothing and said nothing. The byte
  columns had the opposite half of the same bug: every comma was stripped, so
  `1,5 kB` sorted as fifteen kilobytes.

- **pcaptui's own flags were handed to `tshark` in pass-thru.**
  `pcaptui --screencast -r x.pcap | cat` passed `--screencast` straight to
  `tshark`, which rejects it; and a dropped flag left its value behind, so
  `--profile work` became a read filter called `work`.

- **A third of the analysis views were missing from `:help cmdline`.** `http`,
  `dns` and `export` all worked at the colon prompt and none was in the list
  that exists to name them. The commands, the keys, the menu and both help
  screens now come from one list.

- **On an 8-colour terminal the hex pane never showed whether it had focus.**
  Three of its four highlights — the cursor byte, the protocol layer and the
  line-number gutter — were given the same colours focused and unfocused, so
  nothing on screen said where the next keystroke would go. The layer was also
  drawn exactly like the gutter. The cursor, layer and gutter now go plain when
  the pane loses focus, the field stays marked so you can still see which bytes
  it occupies, and the layer has a colour of its own. Tests now state both
  rules for every built-in theme.

- **A theme could be edited without the change reaching any build.** The themes
  the program uses are compiled into the binary; the files under `assets/` are
  their source, and nothing checked that the two agreed. A test now compares
  every embedded asset against the file it came from.

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

- **A stream that could not be reassembled to the end looked like a short
  one.** The parser's error went to the log and nowhere else, so a truncated
  conversation was presented as the whole conversation. It is now reported —
  except when the stream was cancelled, because closing the reader mid-parse is
  this program stopping, not the stream being unreadable.

- **Five goroutines read a `tshark` command field their next load overwrites.**
  Four of them already took a private copy at the top for exactly that reason
  and then read the shared field anyway when building an error message; the
  fifth took no copy at all. They all use the copy now.

- **A configuration file could stop the program from starting.** `search-type`
  or `search-target` set to anything the program did not recognise reached a
  `panic(nil)` — from `search.New`, called during `ui.Build`, so the panic
  happened before the interface appeared and the message was "panic called with
  nil argument", naming neither the key nor the file.

  Both values were read straight out of the configuration, bypassing the two
  functions written a few lines above for exactly this and used everywhere
  else. They now go through those, and every value they can return is handled.

  This is the second configuration that could make the program refuse to start,
  after hiding every column.

- **Settings failed to save without saying so.** Every preference the program
  remembers — recent files, theme, column layout, profile — is one write to the
  configuration file, and those writes discarded their error. A configuration
  directory that is read-only, full, or owned by someone else gave a program
  that appeared to accept every setting and forgot all of them.

  It now says so once, naming the file, and logs every failure.

- **A failed `tshark -G column-formats` was cached as a success, permanently.**
  Neither `Start` nor `Wait` was checked and the function always returned
  success, so a run that produced nothing wrote an empty list to the cache —
  and the cache is only rebuilt when the tshark binary is newer than it. One
  bad run left the install unable to name a column for good: Edit Columns
  offered nothing, and every configured column was discarded as unrecognised.

  No columns is now an error rather than an empty answer, an empty cache from a
  previous version is regenerated rather than trusted, and the failure is
  reported on stderr instead of only in a log file.

- **Three `panic(nil)` in the search left the loader's lock held.** They fired
  when the packet number map and the packet list disagreed, which means the
  loader changed underneath the search. Panicking before releasing the lock
  means that if anything ever recovered the panic, the program would not crash
  — it would hang, with every goroutine needing the loader blocked for good and
  nothing on screen to say why. The search now stops and reports it.

  The remaining `panic(nil)` calls either became errors — a hex search term
  that cannot be read is now reported the way an invalid regex already was — or
  now name the value that was not understood. There are none left.

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
