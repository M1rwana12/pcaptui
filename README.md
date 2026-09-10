<p align="center">
  <img src=".github/assets/banner.svg" alt="pcaptui — read packet captures in your terminal" width="100%">
</p>

<p align="center">
  <a href="https://github.com/m1rwana12/pcaptui/actions/workflows/ci.yml"><img src="https://github.com/m1rwana12/pcaptui/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/m1rwana12/pcaptui/releases/latest"><img src="https://img.shields.io/github/v/release/m1rwana12/pcaptui?color=F0883E&label=release" alt="Latest release"></a>
  <a href="https://pkg.go.dev/github.com/m1rwana12/pcaptui"><img src="https://pkg.go.dev/badge/github.com/m1rwana12/pcaptui.svg" alt="Go reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/licence-MIT-blue" alt="MIT licence"></a>
  <img src="https://img.shields.io/badge/platforms-13-8B949E" alt="13 platforms">
</p>

<p align="center"><b>English</b> · <a href="README.uk.md">Українською</a></p>

<h3 align="center">Wireshark's analysis, where the capture already is.</h3>

<p align="center">
  <code>pcaptui</code> is a terminal interface to <code>tshark</code>: the packet list, the protocol
  tree and the hex view, over SSH, on a machine with no display — without copying the
  capture back to your desktop first.
</p>

<p align="center">
  <img src=".github/assets/demo.svg" alt="Opening a capture, asking what is in it, asking what is wrong with it, and landing on the packets that are wrong" width="100%">
</p>

**What the recording above shows**, on a 92-packet telnet capture:

- <kbd>y</kbd> — **what is in this file.** The protocol hierarchy, as a tree, with the
  share of packets each protocol accounts for.
- <kbd>esc</kbd> <kbd>e</kbd> — **what is wrong with it.** Expert Information: what
  Wireshark's own dissectors flagged, worst first. Here, a malformed packet.
- <kbd>enter</kbd> — **the packets it is about.** The row becomes a display filter and
  the packet list narrows to exactly those packets. Every table here works that way.

## Install

```bash
go install github.com/m1rwana12/pcaptui/cmd/pcaptui@latest
```

Or take a binary from the [latest release](https://github.com/m1rwana12/pcaptui/releases/latest)
— 13 platform/architecture pairs, with checksums. No package in any distribution yet;
the recipes are written and waiting in [`packaging/`](packaging/).

```bash
pcaptui -r traffic.pcap                # a file
pcaptui -r traffic.pcap 'tcp.port==80' # a file, filtered
pcaptui -i eth0 'port 443'             # an interface, with a capture filter
tcpdump -w - port 53 | pcaptui -r -    # a pipe (Unix; not supported on Windows)
```

**It needs [`tshark`](https://www.wireshark.org) on your `PATH`** — that is where every
dissector, filter and statistic comes from. 1.10.2 or newer is enforced at startup, and
if it is not there `pcaptui` says so and stops, rather than failing later in a way that
looks like a broken capture. Two of the views want more than the floor: Credentials
needs tshark 3.2, and following a TLS stream is best on 4.4 or newer.

---

## What you get

**A capture you can move around in.** A scrolling packet list, a protocol tree you
can expand, and a hex pane that highlights the bytes of whichever field you select.
`tcpdump` prints lines; this lets you look.

**Wireshark's display filters, checked before they cost you anything.** The same
syntax, with the filter box turning red while the expression is still wrong. A filter
given on the command line is checked too — either form, `-Y` or positional — so a typo
stops the program with `tshark`'s own explanation instead of opening three empty panes.

**Wireshark's dissectors, all of them.** `pcaptui` dissects nothing itself: there is no
protocol parsing in this repository and no opinions about protocols in it. It asks
`tshark` and draws the answer. When Wireshark learns a protocol, so does this.

**Answers, not just packets.** A capture answers four questions — what this file is,
what is wrong with it, what is in it, and who is on the wire — and the Overview asks
all four at once, in one pass. It opens by itself when a capture finishes loading,
because a screen full of packet number one answers none of them.

**It says when it is showing you less than everything.** A column too narrow for its
value ends in `…` rather than quietly showing a shorter address; the title bar says
how many packets there are and whether a filter is why the list looks short; a
statistic that matched nothing says so in words instead of opening an empty box; and a
capture too large to summarise for free is offered rather than summarised — see
[the limits](#what-it-does-not-do).

---

## Analysis

Eleven views, one key each — the same key that appears beside them in the Analysis
menu. Every row of every table can be pressed: it becomes a display filter.

| Key | | |
|---|---|---|
| <kbd>o</kbd> | **Overview** | all four questions at once, with a sparkline of when the traffic happened |
| <kbd>e</kbd> | Expert Information | what the dissectors think is wrong |
| <kbd>y</kbd> | Protocol Hierarchy | what is in this capture, as a tree |
| <kbd>t</kbd> | Endpoints | who is on the wire, busiest first, IPv4 and IPv6 |
| <kbd>a</kbd> | Credentials | logins this capture carries in the clear |
| <kbd>w</kbd> | HTTP | how the HTTP responses turned out, by status |
| <kbd>d</kbd> | DNS | what was asked for, and with which response codes |
| <kbd>v</kbd> | Conversations | who talked to whom, by packets and bytes |
| <kbd>s</kbd> | Reassemble stream | the stream this packet is in; `:streams tcp`\|`udp`\|`tls`\|`websocket`\|`http2` to pick a family |
| <kbd>p</kbd> | Capture file properties | size, duration, encapsulation, hashes |
| <kbd>x</kbd> | Export objects | write the files this capture carried out to disk |

<p align="center">
  <img src=".github/assets/screenshot-expert.svg" alt="Expert Information, reporting a suspected retransmission" width="100%">
</p>

The statistics respect the display filter in force, so you can ask them about a
subset, and each names the filter at the top of its result. Two exceptions, both
deliberate and both visible:

- **Credentials ignores it**, and says so in its own heading. `tshark` accepts a filter
  for that tap, exits successfully and reports the same logins anyway — measured with a
  filter that excludes every packet in the file — so claiming the filter applied would
  be worse than admitting it does not.
- **Conversations has a checkbox for it**, unticked by default: the whole point of that
  view is usually who is there at all, not who is left after filtering.

### Following a stream

<kbd>s</kbd> follows the transport stream — TCP, or UDP if the packet is not TCP.
Three more families are asked for by name, because each one can legitimately come back
empty and that should not be a surprise:

| | |
|---|---|
| `:streams websocket` | the messages, unmasked, without the HTTP upgrade or the framing bytes |
| `:streams http2` | one exchange out of a connection that multiplexes many, headers decoded to text |
| `:streams tls` | the decrypted payload, which needs a key log — see below |

## Decrypting TLS

With the session keys, TLS payloads become ordinary protocol layers and
`:streams tls` shows the plaintext:

```bash
export SSLKEYLOGFILE=~/keys.log     # then start your browser or run curl
pcaptui --tls-keylog ~/keys.log -r traffic.pcap
```

`pcaptui` checks the key log before it starts and refuses to run if the file is
missing or unreadable. `tshark` accepts a key log path that does not exist, starts
normally, exits zero and decrypts nothing — so a typo would otherwise be
indistinguishable from traffic whose keys you never had.

## Moving around

| | |
|---|---|
| <kbd>/</kbd> | display filter |
| <kbd>tab</kbd> | switch panes |
| <kbd>&#124;</kbd> <kbd>\\</kbd> | pane layout, pane zoom |
| <kbd>ctrl</kbd>+<kbd>f</kbd> | search — by filter, hex, text, or regex |
| <kbd>m</kbd><kbd>a</kbd> <kbd>'</kbd><kbd>a</kbd> | mark a packet, jump back to it |
| <kbd>c</kbd> | copy mode — packets, fields, or the whole stream |
| <kbd>?</kbd> | everything else |

Vim keys work throughout. So does the mouse, in most terminals.

## Is this for you?

**Yes, if** the capture is on a server, or is too big to move, or you already know
Wireshark's filters and want them where the traffic is.

**No, if** you are at a desktop with Wireshark installed and the file is in front of
you. Wireshark's GUI is better than any terminal can be, and this exists for when you
cannot have it. Reach for [`tshark`](https://www.wireshark.org) itself if you want one
answer in a script rather than a session, and for
[termshark](https://github.com/gcla/termshark) — whose code this is built on — if you
want the same idea from its author.

### What it does not do

- **No piped input on Windows.** `pcaptui -r -` works on Unix; the Windows build says
  so and stops.
- **Above 10 MB the Overview offers its statistics instead of running them.** They are
  one pass over every packet, and that pass slows down as the file grows — measured,
  85 MB is 19 s on Linux and 47 s on macOS. What still opens by itself is what the file
  says about itself, which costs a fraction of a second at any size. <kbd>o</kbd> runs
  the rest whenever you ask. `:set start-view packets` turns the whole thing off.
- **TLS following needs a key log**, and says so plainly rather than showing an empty
  pane: without keys there is nothing decrypted to follow.
- **No QUIC stream following.** Its streams are protected past the handshake, so
  without keys it would answer nothing at all — and there is no way to test it here
  honestly. Written down rather than half-built.
- **No live capture on Windows without [Npcap](https://npcap.com)**, which is
  `tshark`'s requirement, not this program's.

## Documentation

- [User Guide](docs/UserGuide.md) — every view, every key, every setting
- [FAQ](docs/FAQ.md) — colours, terminals, live capture, reporting a bug
- [Changelog](CHANGELOG.md) — what changed and why it was wrong before
- [Contributing](docs/Contributing.md) — how to build it and what CI checks

## Built with

[gowid](https://github.com/gcla/gowid) and [tcell](https://github.com/gdamore/tcell)
for the terminal interface, and [Wireshark](https://www.wireshark.org)'s `tshark`
for every byte of the analysis.

<sub>The two screenshots and the animation are drawn by the program itself, from
<a href="scripts/pcaps/demo.pcap"><code>demo.pcap</code></a> and
<a href="scripts/pcaps/telnet-cooked.pcap"><code>telnet-cooked.pcap</code></a> — CI
renders them again on every push and fails if the text no longer matches, so they
cannot quietly go stale. The banner and the logo are drawn by hand.</sub>

## Licence

MIT. See [LICENSE](LICENSE).

`pcaptui` is built on the codebase of [termshark](https://github.com/gcla/termshark)
by Graham Clark, used under the MIT licence.
