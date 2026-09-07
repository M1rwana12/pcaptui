# Security

## Reporting a vulnerability

Use GitHub's [private vulnerability
reporting](https://github.com/m1rwana12/pcaptui/security/advisories/new) rather
than a public issue. It reaches the maintainer directly and stays private until
there is a fix.

Include what you sent in, what happened, and what you expected. A capture file
that triggers the problem is worth more than a description of it — attach one if
you can share it, redacted if you cannot.

You will get an acknowledgement within a few days.

## What is in scope

`pcaptui` reads capture files, which are **untrusted input by definition**:
anything that ends up in a pcap was put there by whoever was on the network.
That makes the following interesting:

- Anything that crashes, hangs or exhausts memory on a crafted capture file
- Anything that escapes the display — control sequences from packet data
  reaching the terminal unescaped
- Anything that causes a command to run that the user did not ask for, through
  a filename, a display filter, or a field value
- Reading files outside those the user named

## What is not

- **Bugs in `tshark`'s dissectors.** All protocol analysis happens in
  `tshark`; report those to
  [Wireshark](https://gitlab.com/wireshark/wireshark/-/issues) — they have a
  security process and a fuzzing programme far beyond this project.
- **Needing capture privileges.** Live capture requires the same permissions
  `tshark` requires. That is the operating system working correctly.
- **A key log file being readable.** If someone can read your
  `SSLKEYLOGFILE`, they can decrypt the traffic without this program.

## Supported versions

The latest release. This is a single-maintainer project; there is no long-term
support branch, and the honest answer is that a fix will land in the next
release rather than being backported.

## Dependencies

`govulncheck` runs on every push and every release, and the release is blocked
if it reports anything reachable from this code. That gate exists because a
parser with a known infinite loop is not a stale dependency when the input is a
capture file — it is the bug.
