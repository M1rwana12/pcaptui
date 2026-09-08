# pcaptui User Guide

- [Getting started](#getting-started)
- [Reading a capture](#reading-a-capture)
- [Moving around](#moving-around)
- [Filtering](#filtering)
- [Searching](#searching)
- [Marks](#marks)
- [Copy mode](#copy-mode)
- [Following a stream](#following-a-stream)
- [Conversations](#conversations)
- [Expert Information](#expert-information)
- [Protocol Hierarchy](#protocol-hierarchy)
- [Decrypting TLS](#decrypting-tls)
- [Columns](#columns)
- [The command line](#the-command-line)
- [Configuration](#configuration)
- [Profiles](#profiles)
- [Colours and themes](#colours-and-themes)
- [When something goes wrong](#when-something-goes-wrong)

## Getting started

`pcaptui` needs `tshark` on your `PATH`. It does no protocol analysis itself —
every dissector, display filter and statistic comes from Wireshark. Version
1.10.2 or newer works; in practice anything you can install today is fine.

```bash
go install github.com/m1rwana12/pcaptui/cmd/pcaptui@latest
pcaptui -r capture.pcap
```

If `tshark` cannot be found, `pcaptui` says so and stops rather than failing
later in a way that looks like a bug in the capture.

## Reading a capture

**A file:**

```bash
pcaptui -r capture.pcap
pcaptui -r capture.pcapng 'tcp.port == 443'
```

The positional argument is a display filter when you are reading a file, and a
capture filter when you are reading an interface.

**Standard input, or a fifo:**

```bash
tcpdump -w - port 53 | pcaptui -r -
mkfifo /tmp/cap && pcaptui -r /tmp/cap
```

Reading from stdin or a fifo is treated as a live source: packets appear as they
arrive and the view keeps up.

**An interface:**

```bash
pcaptui -i eth0
pcaptui -i eth0 -i wlan0 'not port 22'
```

Multiple `-i` flags read several interfaces at once. Capturing needs the same
permissions `tshark` needs — on Linux the Wireshark packages usually set
`dumpcap` up for you; see [running without
root](#capturing-without-root).

**Piped output.** If standard output is not a terminal, `pcaptui` runs `tshark`
with your arguments instead of drawing a UI, so it behaves sensibly inside a
pipeline. Force it either way with `--pass-thru=true|false`.

## Moving around

The window has three panes: the packet list, the protocol tree for the selected
packet, and its bytes.

| Key | |
|---|---|
| `tab` | switch panes |
| <code>&#124;</code> | cycle pane layouts |
| `\` | zoom the current pane, and back |
| `+` `-` | adjust the horizontal split |
| `<` `>` | adjust the vertical split |
| `esc` | open the menus |
| `z` | maximise or restore a dialog |
| `?` | help |
| `q` | quit |

Vim keys work throughout:

| Key | |
|---|---|
| `h` `j` `k` `l` | move left, down, up, right |
| `gg` `G` | top, bottom of the current table |
| `5gg` | go to row 5 |
| `C-w C-w` | switch panes |
| `C-w =` | equalise the panes |
| `ZZ` | quit without confirming |

Most terminals pass the mouse through, so clicking and scrolling work. Use
`shift`-click to select text with the terminal's own selection when you want to
bypass the UI.

## Filtering

Press `/` to reach the filter box and type any Wireshark display filter.

The box validates as you type — it turns red while the expression is incomplete
or wrong, so you know before you press enter. Tab completes field names.

The filter applies to the packet list, and, importantly, to Expert Information
and Protocol Hierarchy as well.

`:filter` opens a list of filters you have used before.

## Searching

`ctrl-f` opens the search bar. You can search four different ways, chosen from
the dropdown next to the box:

- **by display filter** — find the next packet matching an expression
- **by hex bytes** — find a byte sequence anywhere in the packet data
- **in the packet structure** — text in the dissected tree
- **in the packet list** — text in the columns as displayed

Searching runs over the whole capture, not just the packets currently loaded,
and can be interrupted.

## Marks

Marks let you jump between points of interest.

| Key | |
|---|---|
| `ma` … `mz` | mark the current packet, within this capture |
| `'a` … `'z` | jump to that mark |
| `mA` … `mZ` | mark the packet *and* the capture it is in |
| `'A` … `'Z` | jump there, loading that capture if needed |
| `''` | jump back to where you were before the last jump |

Lower-case marks are forgotten when you load a different capture; upper-case
marks persist, which is what makes them useful across files. `:marks` lists
them.

## Copy mode

Press `c` to enter copy mode. The selected region is highlighted; `left` widens
the selection to the enclosing structure, `right` narrows it. `ctrl-c` copies.

| Key | |
|---|---|
| `c` | enter copy mode |
| `left` `right` | widen or narrow the selection |
| `ctrl-c` | copy |
| `q` `c` | leave copy mode |

Copy mode works in the packet list (ranges of packets), the protocol tree
(a field, a layer, or the whole packet) and the stream view.

On a remote machine there is no local clipboard to write to, so `pcaptui` falls
back to your terminal's OSC 52 support if it has any. If nothing works, the
copy dialog shows the text so you can select it by hand.

## Following a stream

Select a TCP or UDP packet, then choose **Reassemble stream** from the Analysis
menu, or run `:streams`.

The reassembled conversation is shown in ASCII, hex or raw, and you can look at
one direction alone or both interleaved. Each chunk is labelled with the
direction it travelled. From here you can filter the packet list down to just
this stream.

## Conversations

`:convs`, or **Conversations** in the Analysis menu, lists the conversations in
the capture by protocol — Ethernet, IPv4, IPv6, TCP, UDP — with packet and byte
counts in each direction.

Select a row and **Prepare** writes a display filter for that conversation into
the filter box without running it, so you can adjust it first; **Apply** runs it
immediately. You can pick direction: both ways, one way, or by source or
destination alone.

## Expert Information

`:expert`, or **Expert Information** in the Analysis menu.

This is what Wireshark's dissectors think is wrong with the capture:
retransmissions, malformed packets, checksum failures, protocol violations —
grouped by severity with errors first, each line giving how many times it
happened, the protocol that raised it, and a summary.

It is usually the quickest way into a capture somebody has just handed you.

## Protocol Hierarchy

`:hierarchy`, or **Protocol Hierarchy** in the Analysis menu.

Every protocol present in the capture, as a tree, with packet and byte counts.
Each protocol is indented under the one carrying it.

Both this and Expert Information respect the display filter currently applied,
so you can ask them about a subset. The filter in force is named at the top of
the result. If the filter matches nothing, `pcaptui` says so — an empty result
is a real answer, not a blank window.

Both are produced by running `tshark` over the file, so each takes a pass over
the capture. They are not cached, because the answer depends on the filter as
well as the file.

## Decrypting TLS

With the session keys, TLS payloads become ordinary protocol layers — HTTP
inside TLS, and so on — and stream reassembly shows the plaintext.

Browsers and many command-line tools write those keys when `SSLKEYLOGFILE` is
set:

```bash
export SSLKEYLOGFILE=~/keys.log
curl https://example.com          # or launch your browser from this shell
```

Then:

```bash
pcaptui --tls-keylog ~/keys.log -r traffic.pcap
```

Or permanently, in `pcaptui.toml`:

```toml
[main]
  tls-keylog = "/home/me/keys.log"
```

The flag wins over the config file.

`pcaptui` checks the file before starting and refuses to run if it is missing or
is not a readable file. This is on purpose. `tshark` accepts a key log path that
does not exist — also an empty path, also a directory — starts normally, exits
zero and decrypts nothing, without a word. A mistyped path then looks exactly
like traffic whose keys you never had, which is the worst way for this
particular feature to fail. An existing but empty file is fine: the browser
creates the log when it starts and fills it in as sessions are negotiated.

## Columns

`:columns`, or **Edit Columns** in the Misc menu.

Add, remove and reorder the columns of the packet list. Columns can be any of
Wireshark's built-in ones, or a custom display filter field — so
`http.host` or `dns.qry.name` can have a column of its own.

You can also add a column straight from a packet: select a field in the protocol
tree and choose **Apply as column**.

## The command line

Press `:` for the command line. Tab lists and completes.

| | |
|---|---|
| `capinfo` | capture file properties |
| `clear-filter` | clear the display filter and apply |
| `clear-packets` | clear the current capture |
| `columns` | choose the columns to display |
| `config` | show the config file |
| `convs` | conversations |
| `expert` | expert information |
| `filter` | pick a recently-used display filter |
| `help` | the help dialogs |
| `hierarchy` | protocol hierarchy |
| `load` | load a capture from the filesystem |
| `logs` | show the log file |
| `map` | map a key to a key sequence |
| `marks` | show current marks |
| `menu` | open the Misc menu |
| `no-theme` | clear the theme for this colour mode |
| `profile` | create, use or delete a profile |
| `quit` | quit |
| `recents` | load a recently-used capture |
| `set` | change a setting |
| `streams` | stream reassembly |
| `theme` | choose a theme |
| `unmap` | remove a key mapping |
| `wormhole` | send the current capture |

`:set` changes settings interactively — `auto-scroll`, `dark-mode`,
`packet-colors`, `term`, `copy-timeout`, `suppress-tshark-errors`, `pager` and a
few others. Type `:set` and press tab.

## Configuration

Settings live in `pcaptui.toml` in your platform's configuration directory.
`:config` shows the path. Cache and log files sit alongside, under a `pcaptui`
directory.

Everything has a working default; the file exists for when you want something
else.

**Where the tools are**

```toml
[main]
  tshark = "/usr/local/bin/tshark"
  dumpcap = "/usr/local/bin/dumpcap"
  capinfos = "/usr/local/bin/capinfos"
```

**Extra arguments for tshark.** `tshark-args` is added to every invocation
`pcaptui` makes; `pdml-args` and `psml-args` are added only when generating the
protocol tree or the packet list respectively.

```toml
[main]
  tshark-args = ["-d", "udp.port==2055,cflow"]
```

**Memory and disk.** `pcap-cache-size` is how many bundles of packets are kept
in memory, `pcap-bundle-size` how many packets are in a bundle, and
`disk-cache-size-mb` caps the on-disk cache. The defaults suit an ordinary
laptop; raise them if you routinely open very large captures and have the room.

**Other settings you may want**

| Key | |
|---|---|
| `tls-keylog` | TLS key log file, as `--tls-keylog` |
| `auto-scroll` | follow a live capture as packets arrive |
| `packet-colors` | colour the packet list using Wireshark's rules |
| `dark-mode` | start in dark mode |
| `term` | override the terminal type `pcaptui` assumes |
| `suppress-tshark-errors` | keep `tshark` diagnostics out of the UI. Off by default: a capture tshark cannot read otherwise shows as an empty window with no explanation |
| `always-keep-pcap` | keep the temporary file written during a live capture |
| `wireshark-profile` | borrow colours and columns from a Wireshark profile |

## Profiles

A profile is a named set of configuration — columns, colours, filters. `:profile`
creates one, switches to one, or deletes one; `-C name` starts in one.

A profile can be linked to a Wireshark profile, in which case its colour rules
and column layout come from Wireshark's own configuration, so a setup you
already have carries over.

## Colours and themes

`:theme` picks a theme for the current colour mode. `default`, `dracula`,
`solarized` and `base16` are built in, and a theme file dropped into the
`themes` directory beside the config file is picked up automatically.

Terminals differ in what they can show. `pcaptui` uses `TERM` to decide, and
`COLORTERM=truecolor` unlocks 24-bit colour if the terminal supports it. If a
theme is only defined for 256 colours it is used in 256-colour mode even when
more is available.

If you use `base16-shell`, colours 0–21 of the palette are remapped and will not
match what a theme intends. `pcaptui` skips them as match candidates when it
sees `BASE16_SHELL` set; force that with `ignore-base16-colors`.

`d` toggles dark mode from the Misc menu, or `:set dark-mode`.

## When something goes wrong

**Nothing decrypts.** Check the key log actually has `CLIENT_RANDOM` lines in
it. `pcaptui` verifies the file exists and can be read, but it cannot tell
whether the keys in it match the traffic in your capture.

**A big capture loads slowly.** Packets load in bundles as you scroll rather
than all at once, which is why the scrollbar settles as you go. Raising
`pcap-cache-size` and `pcap-bundle-size` trades memory for fewer `tshark` runs.

**Colours look wrong.** Almost always `TERM`. See [colours and
themes](#colours-and-themes).

<a name="capturing-without-root"></a>
**Capture says permission denied.** `pcaptui` captures through `dumpcap`, and
`dumpcap` needs privileges the Wireshark packages normally grant. On Linux:

```bash
sudo dpkg-reconfigure wireshark-common   # answer yes
sudo usermod -a -G wireshark $USER       # then log out and back in
```

**Reporting a bug.** `:logs` shows the log file, which records the exact
`tshark` commands that were run — usually the fastest way to see what actually
happened. Run with `--debug` for more, and a profiling server on
`127.0.0.1:6060`.
