# FAQ

## What is this, exactly?

A terminal interface to `tshark`. It draws the packet list, protocol tree and
hex view; `tshark` does all the analysis. There are no dissectors here and no
protocol knowledge — when Wireshark learns a protocol, so does `pcaptui`,
because it is the same code doing the work.

## How is it related to termshark?

`pcaptui` is built on [termshark](https://github.com/gcla/termshark), written by
Graham Clark, who died in 2024. Most of this code is his, used under the MIT
licence he chose. See the [credit section of the README](../README.md#credit).

The practical differences: it works against current Wireshark releases, eleven
known vulnerabilities are gone, it has Expert Information, Protocol Hierarchy
and first-class TLS decryption, and its tests pass on Windows and macOS as well
as Linux.

## Can I use my old termshark config?

Yes. Copy `termshark.toml` to `pcaptui.toml` in the new config directory —
`:config` shows where that is. The settings are unchanged apart from the new
`tls-keylog` key.

## Why does it need tshark? Why not read the pcap directly?

Because Wireshark's dissectors are the point. There are around three thousand
of them, maintained by hundreds of people over twenty-five years. Reimplementing
even a fraction badly would make this a worse tool, not a more independent one.

## Does it work over SSH?

That is most of why it exists. It draws in a terminal and reads files on the
machine it runs on, so you analyse the capture where it already is instead of
copying gigabytes back to your desktop.

Copying to the clipboard over SSH uses OSC 52 if your terminal supports it;
otherwise the text is shown for you to select by hand.

## Can I capture live traffic?

Yes, with `-i`. It needs the same privileges `tshark` needs. On Linux the
Wireshark packages usually configure `dumpcap` for you — if not:

```bash
sudo dpkg-reconfigure wireshark-common
sudo usermod -a -G wireshark $USER
```

Then log out and back in.

## Nothing decrypts even though I passed --tls-keylog

`pcaptui` checks that the file exists and can be read, and refuses to start if
not. What it cannot check is whether the keys inside match your capture. Look
for `CLIENT_RANDOM` lines in the log, and make sure the log was being written
during the capture, not afterwards.

Note that TLS 1.3 session keys are per-connection: a key log captured on a
different day will not decrypt today's traffic.

## It uses a lot of memory on a large capture

Packets load in bundles as you move through the file rather than all at once,
and finished bundles are cached. `pcap-cache-size` controls how many are kept
in memory and `pcap-bundle-size` how big each one is. Lowering them trades
speed for memory; raising them does the opposite.

The protocol tree for a packet is generated on demand, so scrolling the list is
much cheaper than expanding every packet.

## Why is my terminal showing the wrong colours?

`TERM`, nearly always. Set `COLORTERM=truecolor` if your terminal supports
24-bit colour. If you use `base16-shell`, colours 0–21 are remapped and no theme
can predict them; `pcaptui` avoids those slots when it sees `BASE16_SHELL`.

## The UI is unreadable after I resized / changed themes

`:no-theme` clears the theme for the current colour mode and returns to the
built-in default, which is defined for every terminal.

## How do I see what tshark is actually being run?

`:logs`. Every command is logged with its full argument list. This is the
fastest way to understand unexpected output — usually the answer is visible in
the arguments.

## Is there a Windows build?

It builds and its tests pass on Windows, and `tshark` is available there. Live
capture needs Npcap installed, which the Wireshark installer offers.

## Are there distribution packages?

Not yet. `go install` is the way in for now, and releases carry binaries.

## Where do I report a bug?

[Issues](https://github.com/m1rwana12/pcaptui/issues). Include the output of
`:logs` if the problem involves `tshark` doing something unexpected.
