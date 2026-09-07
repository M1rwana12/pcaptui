# Brand

What this project looks and sounds like, so it stays consistent when somebody
adds a screenshot, writes a release note or draws a diagram a year from now.

## Voice

**Plain, measured, specific.** The audience is people debugging a network at
two in the morning. They want the fact, not the enthusiasm.

Four rules, in order of how often they matter:

1. **Say the number.** "Eleven vulnerabilities reachable from this code" beats
   "improved security". If there is no number, say what you actually observed.
2. **Name the failure.** "`tshark` accepts a key log that does not exist, exits
   zero and decrypts nothing" is the whole reason a feature exists. Write that
   down, not "added validation".
3. **No marketing verbs.** Nothing is blazing, seamless, powerful or
   revolutionary. It either works on your capture or it does not.
4. **Admit the edges.** "End-to-end decryption is not verified here; it needs a
   capture with a matching key log" costs one sentence and buys all the trust
   the rest of the document spends.

Tone shifts by context but voice does not: an error message is terse, the user
guide is patient, a release note is factual. None of them are chatty.

### Words

Say `capture` or `capture file`, not "pcap file" in prose — the file may be
pcapng. Say `display filter` and `capture filter`, never just "filter", because
they are different things and mixing them is the most common confusion in this
domain. Say `tshark` and `Wireshark` with their own capitalisation.

The program is `pcaptui`, lower case, always — including at the start of a
sentence. Rewrite the sentence if that reads badly.

## Colour

The palette is GitHub's dark surface plus one accent, chosen so a dark card
renders identically on light and dark themes and never fights the terminal
screenshots it sits next to.

| Role | Hex | Where |
|---|---|---|
| Ink | `#0D1117` | backgrounds, the logo card |
| Deep | `#010409` | the terminal well inside the banner |
| Line | `#21262D` | borders, dividers |
| Muted | `#30363D` | unselected packet rows, gutter marks |
| Grey | `#8B949E` | secondary text, captions |
| Paper | `#E6EDF3` | primary text on ink |
| **Accent** | **`#F0883E`** | the selected packet, the cursor, one thing per image |
| Signal | `#3FB950` | shell prompts, passing states |

**The accent marks the selection and nothing else.** One amber element per
image. The moment there are two, the picture stops saying "this is the packet
you are looking at".

Amber rather than blue on purpose: blue in this field means Wireshark, and this
project is not Wireshark and should not borrow its authority.

## Type

Monospace everywhere, including the wordmark:

```
ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas,
"Liberation Mono", monospace
```

The subject is a terminal program; a proportional wordmark would be the first
thing that felt untrue. Wordmark is bold with `-2` letter-spacing at display
sizes, regular everywhere else.

## Marks

| File | Use |
|---|---|
| `.github/assets/logo.svg` | square mark, avatars, favicons, anywhere under 128px |
| `.github/assets/banner.svg` | README header, social preview |

Both are SVG and hold up at any size. The logo is legible down to 32px — the
three-row list plus gutter is exactly as much detail as survives that.

**Do not:** recolour them, put them on a light background, stretch them, or add
a glow. If a placement needs a light background, put the dark card on it rather
than inverting the mark.

## Screenshots

Real terminal, real capture, default theme, at least 100 columns. No decorative
window frames, no drop shadows, no arrows drawn on top. If something needs
pointing at, crop to it instead.

Never publish a screenshot with real traffic in it. Capture files carry
addresses, hostnames and payloads; use a synthetic capture or one you have
checked line by line.
