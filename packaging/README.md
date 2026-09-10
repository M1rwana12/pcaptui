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
cp .../packaging/aur/{PKGBUILD,.SRCINFO} .   # .SRCINFO is generated already
makepkg -si                                  # build and install locally first
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

## What was checked, and what was not

Both definitions are at **v1.2.0**, and every hash in them came from that
release: the source tarball's own hash for the PKGBUILD, and the four archive
hashes out of `checksums.txt` for the formula. Those are different archives and
one cannot stand in for the other.

**The PKGBUILD is built and installed on Arch by the `Packaging` workflow** —
`workflow_dispatch`, so it runs when somebody asks. It is not a syntax check:
`namcap` lints the recipe and then the built package, `makepkg` verifies the
source hash against the published tag, `check()` runs the test suite with a
real `wireshark-cli` installed, and the resulting package is installed with
`pacman -U` and the binary run out of `/usr/bin`.

Measured on v1.2.0: **namcap clean on both**, all 26 test packages pass with
tshark present rather than skipped, and `/usr/bin/pcaptui --pass-thru=false
--version` answers `pcaptui v1.2.0`.

`.SRCINFO` is committed beside the PKGBUILD, because that is the file the AUR
actually serves - the site reads the version, the dependencies and the sources
from it, and one left stale after a version bump lies quietly while the package
still builds. The workflow regenerates it and fails if it differs from the
committed copy. Re-run the workflow after changing the version, before pushing
to the AUR.

**The formula is not tested anywhere, and cannot be until a release is
public.** It installs release archives by URL, and a draft release's assets
need authentication - `brew` gets a 404. So the order is: publish the release,
then `brew install --formula packaging/homebrew/pcaptui.rb` on a Mac, then push
the tap.

The formula's test uses `--pass-thru=false` deliberately: under `brew test`
stdout is not a terminal, so a bare `--version` runs `tshark` and prints
*its* version, asserting nothing about this program. That was measured, not
assumed - and the same trap is why the Arch job passes the flag too.
