# Credits

`pcaptui` is built on [termshark](https://github.com/gcla/termshark), written by
**Graham Clark**, who died in 2024. The packet loader, the widget set and the great
majority of this code are his, used here under the MIT licence he chose. His
copyright notice stands in every file he wrote, and in [LICENSE](../LICENSE).

The terminal widgets are [gowid](https://github.com/gcla/gowid), also his, built
on [tcell](https://github.com/gdamore/tcell) by Garrett D'Amore.

## The fix that was already written

[Gilbert Ramirez](https://github.com/gilramir), a Wireshark core developer with
over nine hundred commits to Wireshark itself, fixed the Wireshark 4.x column
format change in [gcla/termshark#170](https://github.com/gcla/termshark/pull/170)
on 2025-10-20. It was reported by [@rbalint](https://github.com/rbalint) the day
before. The pull request was never merged; there was no longer anyone to merge
it. The fix in this tree was written before that PR was found and arrives at the
same design.

## Everyone who contributed to termshark

Their work is in this code. Listed as recorded in the original repository:

[@Hongarc](https://github.com/Hongarc), [@IAXES](https://github.com/IAXES), [@NicolaiSoeborg](https://github.com/NicolaiSoeborg), [@Piping](https://github.com/Piping), [@QuLogic](https://github.com/QuLogic), [@ReK2Fernandez](https://github.com/ReK2Fernandez),
[@RobertLarsen](https://github.com/RobertLarsen), [@Thann](https://github.com/Thann), [@abenson](https://github.com/abenson), [@basondole](https://github.com/basondole), [@cmosig](https://github.com/cmosig), [@dawidd6](https://github.com/dawidd6),
[@deliciouslytyped](https://github.com/deliciouslytyped), [@denyspozniak](https://github.com/denyspozniak), [@dragosmaftei](https://github.com/dragosmaftei), [@elig0n](https://github.com/elig0n), [@factorion](https://github.com/factorion), [@freddii](https://github.com/freddii),
[@gdluca](https://github.com/gdluca), [@gvanem](https://github.com/gvanem), [@herbygillot](https://github.com/herbygillot), [@hook-s3c](https://github.com/hook-s3c), [@inzel](https://github.com/inzel), [@jJit0](https://github.com/jJit0),
[@jboverfelt](https://github.com/jboverfelt), [@jerry73204](https://github.com/jerry73204), [@joelparker](https://github.com/joelparker), [@kevinhwang91](https://github.com/kevinhwang91), [@lennartkoopmann](https://github.com/lennartkoopmann), [@linsong](https://github.com/linsong),
[@loudsong](https://github.com/loudsong), [@luzpaz](https://github.com/luzpaz), [@mazball](https://github.com/mazball), [@mharjac](https://github.com/mharjac), [@mingrammer](https://github.com/mingrammer), [@mrash](https://github.com/mrash),
[@msenturk](https://github.com/msenturk), [@nmeum](https://github.com/nmeum), [@pocc](https://github.com/pocc), [@punkymaniac](https://github.com/punkymaniac), [@qbit](https://github.com/qbit), [@rongyi](https://github.com/rongyi),
[@rski](https://github.com/rski), [@sagis-tikal](https://github.com/sagis-tikal), [@sean-abbott](https://github.com/sean-abbott), [@szuecs](https://github.com/szuecs), [@the-c0d3r](https://github.com/the-c0d3r), [@thebyrdman-git](https://github.com/thebyrdman-git),
[@thejerrod](https://github.com/thejerrod), [@thordy](https://github.com/thordy), [@uzxmx](https://github.com/uzxmx), [@wfailla](https://github.com/wfailla), [@winpat](https://github.com/winpat), [@zi0r](https://github.com/zi0r),
[@zoulja](https://github.com/zoulja), 

Plus everyone who filed an issue that turned into a fix.
