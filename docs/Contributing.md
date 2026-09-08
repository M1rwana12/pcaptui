# Contributing

## Building

```bash
git clone https://github.com/m1rwana12/pcaptui
cd pcaptui
go build ./cmd/pcaptui
```

Go 1.25 or newer. There is nothing to generate and no code generation step.

## Testing

```bash
go test ./...
```

**`tshark` must be on your `PATH`.** Several tests drive a real `tshark` over a
real capture, and without it they fail for reasons that look like code bugs but
are not. This is deliberate: the parts of this program that talk to `tshark` are
exactly the parts that break when Wireshark changes, and a unit test cannot see
that coming.

The column format reader was broken for years — every column title in the
preferences came out mangled against Wireshark 4.x — because the only tests
covering it asserted what the code already did. If you are testing anything that
consumes `tshark` output, run `tshark`.

## Before sending a change

```bash
gofmt -l .                      # must print nothing
go vet -composites=false ./...
go test ./...
```

CI runs these on Linux, macOS and Windows, plus `govulncheck`. All three
platforms matter: a test that quietly only worked on Linux is how a Windows bug
survived for years here.

### The images in the README

`scripts/screenshots.sh` regenerates all of them; `--check` compares without
touching the committed files, and CI runs that. Two stills and one animation:
the animation is recorded by `--screencast`, which photographs the screen after
each keystroke and writes the frames as one looping SVG, plus a `.txt` of every
frame so the same gate covers it.

They have to be the **Linux** rendering — `gowid` picks its frame characters
from `runtime.GOOS`, so a Windows or macOS run produces files that can never
match CI. Download the `screenshots-linux` artifact from any CI run instead.

When that gate goes red, re-run the job before believing it. A capture load
arrives in two halves and a screenshot taken between them looks like a
regression; that is fixed, but a gate driven by a real program on a shared
runner is never entirely above suspicion.

### If you change anything under `assets/`

The themes are compiled into the binary by `statik`, not read from the source
tree, so editing `assets/themes/*.toml` does nothing on its own — a rebuild
will happily keep using the old colours and give you no reason why.

```bash
go install github.com/rakyll/statik@latest
cd assets && statik -src=. -f      # or: go generate ./assets
gofmt -w assets/statik/statik.go   # statik does not format what it writes
```

Commit the regenerated `assets/statik/statik.go` with your change.

### About `-composites=false`

`go vet` flags 34 unkeyed struct literals, all of them `gowid` types. They are a
real fragility — if `gowid` ever reorders a field, every one of them breaks
silently, and `gowid` is unmaintained so nobody would be warned. They are not
fixed yet, and the check is disabled explicitly in the workflow rather than
quietly. `go vet -composites=true ./...` lists them if you want to help.

## Commit messages

Say what changed and why it was wrong before. The subject line is a sentence
about the defect, not a label:

> `fix: error text was being passed as a format string`

not `fix: printf issues`. Anyone reading the log later wants to know whether
this commit is the one that explains the behaviour they are seeing.

Include the measurement if there was one. "tshark accepts a keylog path that
does not exist, exits zero and decrypts nothing" is worth more in the log than
"added validation".

## Releasing

1. Set the version in `version.go`.
2. Update `CHANGELOG.md`.
3. `gofmt -l .` clean, `go vet -composites=false ./...`, `go test ./...` with
   `tshark` present.
4. Push, and **wait for CI to go green on that exact commit** — check the
   commit SHA, not just the most recent run.
5. Tag and let the release workflow build the binaries.

## Where things are

| | |
|---|---|
| `cmd/pcaptui` | entry point, flag handling, start-up |
| `pkg/pcap` | running `tshark`, loading packets, the cache |
| `pkg/pdmltree` | the protocol tree model |
| `pkg/shark` | column formats, Wireshark configuration |
| `pkg/stats` | Expert Information, Protocol Hierarchy |
| `pkg/streams` | TCP/UDP reassembly |
| `pkg/convs` | conversations |
| `pkg/cli` | command-line options |
| `ui` | everything on screen |
| `widgets` | reusable terminal widgets |

The widget set is [gowid](https://github.com/gcla/gowid), built on
[tcell](https://github.com/gdamore/tcell). It is unmaintained upstream, so
treat it as a fixed dependency: work around its limitations rather than
expecting fixes.

## Copyright headers

Leave the copyright header at the top of a file alone when you edit it. The MIT
licence this project is under requires those notices to travel with the code.
