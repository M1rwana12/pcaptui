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

`pcaptui` is a terminal interface to `tshark`. It gives you the packet list, the
protocol tree and the hex view you would get from Wireshark, over SSH, on a
machine with no display, without copying a 2GB capture back to your desktop
first.

```bash
pcaptui -r traffic.pcap
pcaptui -i eth0 'port 443'
```

<p align="center">
  <img src=".github/assets/screenshot-packets.svg" alt="The packet list, protocol tree and hex view" width="100%">
</p>

<sub>That image is drawn by the program itself from
<a href="scripts/pcaps/demo.pcap"><code>scripts/pcaps/demo.pcap</code></a>, and CI
fails if it stops matching what the program renders — so it cannot go stale.</sub>

## Why

Wireshark is the best protocol analyser there is, and `tshark` carries all of
it — every dissector, every display filter, every statistic. What `tshark`
does not give you is a way to move around a capture: no scrolling packet list,
no expandable protocol tree, no clicking a field to see the bytes it came from.

`pcaptui` supplies exactly that, and delegates all the analysis to `tshark`.
It has no dissectors of its own and no opinions about protocols. When Wireshark
learns a new protocol, so does this.

## Requirements

`tshark`, from [Wireshark](https://www.wireshark.org). Version 1.10.2 or newer;
anything from the last decade will do. `pcaptui` will tell you if it cannot
find it.

Live capture also needs permission to capture — on Linux that usually means
`dumpcap` with the right capabilities, which the Wireshark packages set up for
you.

## Install

```bash
go install github.com/m1rwana12/pcaptui/cmd/pcaptui@latest
```

The binary lands in `~/go/bin`. It is a single static executable with no
runtime dependencies beyond `tshark` itself.

## What it does

**Read captures or watch live traffic.** Files, fifos, stdin, or an interface.
A live capture streams into the same view, and you can filter it while it runs.

**Filter with Wireshark's display filters.** The same syntax, checked as you
type — the filter box turns red when the expression is not valid.

**Follow TCP and UDP streams.** Reassembled, in ASCII or hex, either direction
or both.

**See conversations by protocol.** Sorted by packets or bytes, and you can turn
any row straight into a display filter.

**Search.** By display filter, by hex bytes, by text in the packet structure, or
by text in the packet list.

**Copy anything.** Ranges of packets, single fields, the reassembled stream —
into the system clipboard, or out through a terminal that supports OSC 52 when
you are on a remote machine.

### Expert Information

What the dissectors think is wrong with a capture: retransmissions, malformed
packets, checksum failures, protocol violations — grouped by severity, worst
first.

```
:expert
```

It is usually the fastest way to find the problem in a capture you have just
been handed, without reading it packet by packet.

### Protocol Hierarchy

Every protocol present, by packet and byte count, as a tree. Answers "what is
actually in this file" in one screen.

```
:hierarchy
```

Both views narrow to the display filter you have applied, the way Wireshark
does, and both say so plainly when the filter matches nothing rather than
showing you an empty box.

### Decrypting TLS

If you have the session keys, you see the plaintext:

```bash
export SSLKEYLOGFILE=~/keys.log     # then start your browser or run curl
pcaptui --tls-keylog ~/keys.log -r traffic.pcap
```

`pcaptui` checks the key log before it starts and refuses to run if the file is
missing or unreadable. That check matters more than it sounds: `tshark` accepts
a key log path that does not exist, starts normally, exits zero and decrypts
nothing — so a typo is indistinguishable from traffic whose keys you never had.

## Configuration

Settings live in `pcaptui.toml`, under your platform's config directory. Run
`:config` to see where. Everything has a working default; the file is for when
you want something different.

## Documentation

- [User Guide](docs/UserGuide.md) — every view, every key, every setting
- [FAQ](docs/FAQ.md) — the questions that actually come up
- [Contributing](docs/Contributing.md) — building, testing, releasing
- [Brand](docs/Brand.md) — how this project looks and sounds
- [Security policy](SECURITY.md) — what counts as a vulnerability here

## Built with

[gowid](https://github.com/gcla/gowid) for the terminal widgets, on top of
[tcell](https://github.com/gdamore/tcell).

## Licence

MIT. Copyright (c) 2026 m1rwana12 — see [LICENSE](LICENSE).
