# Packaging

Two package definitions, ready to publish. Neither has been submitted — that is
a step somebody has to take by hand, and both are written so that the taking is
the only work left.

Why these two first: `go install` reaches Go developers, and this program's
audience is network engineers. Every install figure for a tool in this niche
exists because somebody packaged it. Measured on Homebrew, installs over one
year: `wireshark` 56,152 · `tcpdump` 5,679 · `sngrep` 718 · `termshark` 619.
The gap between the first and the last is the reachable audience.

## AUR — `aur/PKGBUILD`

The lowest barrier of any channel: no popularity threshold, no sponsor, anyone
with an account can submit. Arch's audience is also the right one — terminal,
networks, SSH.

```bash
git clone ssh://aur@aur.archlinux.org/pcaptui.git
cd pcaptui
cp .../packaging/aur/PKGBUILD .
makepkg --printsrcinfo > .SRCINFO
makepkg -si                       # build and install locally first
git add PKGBUILD .SRCINFO && git commit && git push
```

Builds from the source tag rather than the release binary, which is what Arch
expects, and runs the test suite in `check()` — including the tests that drive
a real `tshark`, so the package is checked against the `tshark` the user will
actually have.

## Homebrew — `homebrew/pcaptui.rb`

`homebrew-core` has no documented star threshold, but in practice maintainers
weigh how established a project is, and this one is days old. A personal tap
works immediately and needs nobody's approval:

```bash
gh repo create m1rwana12/homebrew-pcaptui --public
# put pcaptui.rb in Formula/pcaptui.rb, push, then:
brew tap m1rwana12/pcaptui
brew install pcaptui
```

Installs the release binaries rather than building from source, so what lands
is the artefact CI produced and the release checksums cover.

**When the version changes**, both files need the new version and new hashes:

```bash
curl -sL https://github.com/M1rwana12/pcaptui/releases/download/vX.Y.Z/checksums.txt
curl -sL https://github.com/M1rwana12/pcaptui/archive/refs/tags/vX.Y.Z.tar.gz | sha256sum
```

The first gives the four binary hashes the formula needs; the second gives the
source hash the PKGBUILD needs. They are different archives — do not reuse one
for the other.

## What was checked here, and what was not

Both definitions were written against the real v1.0.0 release, and every hash
above came from that release rather than from a guess.

The formula's test uses `--pass-thru=false` deliberately: under `brew test`
stdout is not a terminal, so a bare `--version` runs `tshark` and prints
*its* version, asserting nothing about this program. That was measured, not
assumed.

Neither package has been built on its target platform — there is no Arch or
macOS machine here. `makepkg` and `brew install --build-from-source` are the
first thing to run, before submitting anywhere.
