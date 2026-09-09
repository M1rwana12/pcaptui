// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/m1rwana12/pcaptui/pkg/capinfo"
	"github.com/m1rwana12/pcaptui/pkg/stats"
)

//======================================================================

// expertLines lays out Expert Information.
//
// The severity is a column here rather than a section heading, because the
// rows have to be one flat list for any of them to be selectable, and a row
// that has floated away from its heading has to carry its severity with it.
// Order is kept as tshark gives it, which is worst first.
//
// The numbered rows are folded first. Without that, a capture of 376,832
// packets opened on 4,108 rows of which 4,095 said "Duplicate ACK (#n)" - the
// four that mattered were at the top and nothing on screen said the rest were
// one fact repeated.
func expertLines(rows []stats.ExpertRow, totals []stats.SeverityTotal) statsView {
	rows = stats.CollapseNumbered(rows)
	if len(rows) == 0 {
		return statsView{}
	}

	sevw, grpw, prow, cntw := len("Severity"), len("Group"), len("Protocol"), len("Count")
	for _, r := range rows {
		sevw = max(sevw, len(severityLabel(r.Severity)))
		grpw = max(grpw, len(r.Group))
		prow = max(prow, len(r.Protocol))
		cntw = max(cntw, len(groupDigits(r.Count)))
	}

	line := func(sev, grp, pro, cnt, sum string) string {
		return fmt.Sprintf("%-*s  %-*s  %-*s  %*s  %s",
			sevw, sev, grpw, grp, prow, pro, cntw, cnt, sum)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(severityLabel(r.Severity), r.Group, r.Protocol,
				groupDigits(r.Count), expertSummary(r)),
			Filter: r.DisplayFilter(),
		})
	}

	titles := line("Severity", "Group", "Protocol", "Count", "Summary")
	v.Header = []string{titles, rule(titles, v.Rows)}

	// tshark states the per-severity totals in its section headings and this
	// program read them and threw them away, so a capture with 430,012
	// note-level events said nothing resembling that anywhere.
	if line := severityTotals(totals); line != "" {
		v.Header = append([]string{line, ""}, v.Header...)
	}

	return v
}

// expertSummary says when a row stands for a family rather than one message,
// and which numbers it covers - "Duplicate ACK (#1-#4095)". A folded count
// with no sign that it was folded would read as one problem seen that often.
func expertSummary(r stats.ExpertRow) string {
	if r.Merged > 1 {
		return fmt.Sprintf("%s (#%d-#%d)", r.Summary, r.FirstNum, r.LastNum)
	}
	return r.Summary
}

// severityTotals is the line tshark already knew and the program did not say.
func severityTotals(totals []stats.SeverityTotal) string {
	parts := make([]string, 0, len(totals))
	for _, t := range totals {
		if t.Count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %s", groupDigits(t.Count), severityWord(t.Severity, t.Count)))
	}
	return strings.Join(parts, " · ")
}

// severityWord reads as a quantity of things rather than as tshark's own
// section name: "16,384 chats", not "16,384 Chats" - and "1 note", because a
// capture with one of them said "1 notes".
func severityWord(s string, n int) string {
	if n == 1 {
		return strings.ToLower(severityLabel(s))
	}
	if s == "Warns" {
		return "warnings"
	}
	return strings.ToLower(s)
}

// severityLabel drops tshark's plural. A column of "Errors / Errors / Notes"
// reads as a count of something; the column says what one row is.
func severityLabel(s string) string {
	switch s {
	case "Errors":
		return "Error"
	case "Warns":
		return "Warning"
	case "Notes":
		return "Note"
	case "Chats":
		return "Chat"
	}
	return s
}

//======================================================================

// hierarchyLines lays out Protocol Hierarchy, keeping the indentation that
// says what is carried by what.
func hierarchyLines(rows []stats.HierarchyRow) statsView {
	if len(rows) == 0 {
		return statsView{}
	}

	namew, framew := len("Protocol"), len("Frames")
	for _, r := range rows {
		namew = max(namew, 2*r.Depth+len(r.Protocol))
		framew = max(framew, len(groupDigits(r.Frames)))
	}

	line := func(name, frames, bytes string) string {
		return fmt.Sprintf("%-*s  %*s  %s", namew, name, framew, frames, bytes)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(
				strings.Repeat(" ", 2*r.Depth)+r.Protocol,
				// Grouped like every other table here. The overview used to print
				// "frame 7 1027" six lines above "192.0.2.10 7 1,027" - the same
				// quantity in two spellings, in one dialog.
				groupDigits(r.Frames),
				groupDigits(r.Bytes),
			),
			Filter: r.DisplayFilter(),
		})
	}

	titles := line("Protocol", "Frames", "Bytes")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
}

// httpLines lays out the HTTP packet counter, keeping the indentation that
// says a "404 Not Found" is one of the "4xx: Client Error" responses.
//
// Only the rows tshark counted are shown. Its table is a fixed skeleton of
// every status class it knows, so a capture with three responses prints four
// useful lines under sixteen zeroes, and the zeroes are what the eye lands on.
func httpLines(rows []stats.TreeRow) statsView {
	rows = stats.NonEmptyRows(rows)
	if len(rows) == 0 {
		return statsView{}
	}

	namew, cntw := len("Response"), len("Count")
	for _, r := range rows {
		namew = max(namew, r.Depth+len(r.Name))
		cntw = max(cntw, len(groupDigits(r.Count)))
	}

	line := func(name, count, pct string) string {
		return fmt.Sprintf("%-*s  %*s  %s", namew, name, cntw, count, pct)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(
				strings.Repeat(" ", r.Depth)+r.Name,
				groupDigits(r.Count),
				r.Percent,
			),
			Filter: stats.HTTPFilter(r.Name),
		})
	}

	titles := line("Response", "Count", "Percent")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
}

//======================================================================

// credentialLines lays out the logins tshark could read in the clear.
//
// The only table here whose rows name a packet, so the Packet column comes
// first and Enter on a row lands on that exact frame rather than on everything
// matching some text.
func credentialLines(rows []stats.CredentialRow) statsView {
	if len(rows) == 0 {
		return statsView{}
	}

	pktw, prow, userw := len("Packet"), len("Protocol"), len("Username")
	for _, r := range rows {
		pktw = max(pktw, len(groupDigits(r.Packet)))
		prow = max(prow, len(r.Protocol))
		userw = max(userw, len(r.Username))
	}

	line := func(pkt, proto, user, info string) string {
		return strings.TrimRight(fmt.Sprintf("%*s  %-*s  %-*s  %s",
			pktw, pkt, prow, proto, userw, user, info), " ")
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text:   line(groupDigits(r.Packet), r.Protocol, r.Username, r.Info),
			Filter: r.DisplayFilter(),
		})
	}

	titles := line("Packet", "Protocol", "Username", "Info")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
}

//======================================================================

// dnsLines lays out the DNS statistics, keeping the indentation that says
// which section a row belongs to - the same name appears in more than one.
//
// An Average column, which the HTTP table does not need and this one does: the
// row that says how long the server took to answer has its count in packets
// and its answer in milliseconds, and the count alone says nothing. It is only
// shown when some row has one.
func dnsLines(rows []stats.TreeRow) statsView {
	rows = stats.NonEmptyRows(rows)
	if len(rows) == 0 {
		return statsView{}
	}

	namew, cntw, avgw := len("Statistic"), len("Count"), 0
	for _, r := range rows {
		namew = max(namew, r.Depth+len(r.Name))
		cntw = max(cntw, len(groupDigits(r.Count)))
		avgw = max(avgw, len(r.Average))
	}
	if avgw > 0 {
		avgw = max(avgw, len("Average"))
	}

	line := func(name, count, avg, pct string) string {
		if avgw == 0 {
			return fmt.Sprintf("%-*s  %*s  %s", namew, name, cntw, count, pct)
		}
		return fmt.Sprintf("%-*s  %*s  %*s  %s",
			namew, name, cntw, count, avgw, avg, pct)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(
				strings.Repeat(" ", r.Depth)+r.Name,
				groupDigits(r.Count),
				r.Average,
				r.Percent,
			),
			Filter: stats.DNSFilter(r.Parent, r.Name),
		})
	}

	titles := line("Statistic", "Count", "Average", "Percent")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
}

//======================================================================

// endpointLines lays out Endpoints, busiest first.
//
// tshark prints them in the order it met them, which for a two-host capture is
// no order at all and for a busy one buries the host you are looking for. The
// question this table answers is "who is doing the most", so it is sorted by
// that.
func endpointLines(rows []stats.EndpointRow) statsView {
	if len(rows) == 0 {
		return statsView{}
	}

	sorted := make([]stats.EndpointRow, len(rows))
	copy(sorted, rows)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Bytes > sorted[j].Bytes
	})

	addrw := len("Address")
	pktw, bytew := len("Packets"), len("Bytes")
	txw, rxw := len("Sent"), len("Received")
	for _, r := range sorted {
		addrw = max(addrw, len(r.Address))
		pktw = max(pktw, len(groupDigits(r.Packets)))
		bytew = max(bytew, len(groupDigits(r.Bytes)))
		txw = max(txw, len(groupDigits(r.TxBytes)))
		rxw = max(rxw, len(groupDigits(r.RxBytes)))
	}

	line := func(addr, pkts, bytes, tx, rx string) string {
		return fmt.Sprintf("%-*s  %*s  %*s  %*s  %*s",
			addrw, addr, pktw, pkts, bytew, bytes, txw, tx, rxw, rx)
	}

	var v statsView
	for _, r := range sorted {
		v.Rows = append(v.Rows, statsLine{
			Text: line(r.Address,
				groupDigits(r.Packets), groupDigits(r.Bytes),
				groupDigits(r.TxBytes), groupDigits(r.RxBytes)),
			Filter: r.DisplayFilter(),
		})
	}

	// Bytes rather than packets in the two direction columns: a host sending
	// many small acknowledgements and one sending few large payloads look the
	// same by packet count and nothing alike by volume.
	titles := line("Address", "Packets", "Bytes", "Sent", "Received")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
}

// overviewLimitFor is how many rows of each section the overview shows on a
// screen of this height.
//
// It used to be a constant four, whatever the terminal. Measured at 120x36:
// four expert rows, "… and 6 more", and then eight blank lines above the Close
// button - every one of the hidden rows would have fitted.
//
// The budget is the dialog's height, which is four fifths of the screen, less
// what is drawn around the rows: its frame, the heading and the blank under
// it, the rule and the button at the foot, a title and a rule for each of the
// four sections, a blank line between them, and the file's own facts, which
// are never cut.
//
// Three or more, always. A screen too short for that scrolls, and a summary
// showing two rows of each section is still a summary; one showing none is a
// list of headings.
func overviewLimitFor(screenHeight int, factRows int) int {
	const (
		dialogFraction = 0.8
		frame          = 2
		heading        = 2
		foot           = 2
		sections       = 4
		perSection     = 2 // its title and the rule under it
		separators     = 3
		fewest         = 3
	)

	available := int(float64(screenHeight)*dialogFraction) -
		frame - heading - foot - sections*perSection - separators - factRows

	// The three statistics share what is left; the file's facts are already
	// taken out of it.
	return max(fewest, available/3)
}

// overviewLines is what the capture is, plus the three statistics, as one
// screen.
//
// Order is the order the questions get asked: what is this file, what is wrong
// with it, what is in it, who is on the wire. Problems before contents because
// that is what someone handed a capture is looking for, and because a capture
// with none is itself worth knowing in one line.
//
// The file's own facts come first and are the cheapest of the four by a long
// way - capinfos took 0.27 s on a 44 MB capture where the -z pass over the
// same file took 9.1 to 11.6 s. They are also the only ones that answer
// "when": "this is a forty-second slice from 1999, not the hour you asked
// for" is often the whole answer, and it used to be behind a separate key on
// a separate screen.
//
// Every row keeps the filter its own table would have given it, so the summary
// is not a dead end - enter still lands on the packets.
func overviewLines(out string, info capinfo.Info, limit int) statsView {
	var v statsView

	section := func(title string) {
		if len(v.Rows) > 0 {
			v.Rows = append(v.Rows, statsLine{})
		}
		v.Rows = append(v.Rows, statsLine{Text: title}, statsLine{Text: rule(title, nil)})
	}

	add := func(title string, body statsView, empty string) {
		section(title)

		if len(body.Rows) == 0 {
			v.Rows = append(v.Rows, statsLine{Text: "  " + empty})
			return
		}

		shown := body.Rows
		if len(shown) > limit {
			shown = shown[:limit]
		}
		for _, r := range shown {
			v.Rows = append(v.Rows, statsLine{Text: "  " + r.Text, Filter: r.Filter})
		}
		if len(body.Rows) > len(shown) {
			v.Rows = append(v.Rows, statsLine{
				Text: fmt.Sprintf("  … and %d more", len(body.Rows)-len(shown)),
			})
		}
	}

	// Not through add: this is a fixed handful of facts rather than the head of
	// a table, so there is nothing to cut and nothing to say "and N more"
	// about. It is skipped entirely when capinfos said nothing, which is what a
	// missing or failed capinfos looks like from here.
	if facts := fileFactsLines(info, stats.ParseIOStat(out)); len(facts) > 0 {
		section("What this file is")
		v.Rows = append(v.Rows, facts...)
	}

	add("What is wrong", expertLines(stats.ParseExpert(out), nil),
		"Nothing the dissectors object to.")
	add("What is in it", hierarchyLines(stats.ParseHierarchy(out)),
		"No protocols reported.")
	add("Who is on the wire", endpointLines(stats.ParseEndpoints(out)),
		"No addresses on the wire.")

	return v
}

// fileFactsLines is the file's own properties, laid out as label and value.
//
// Five of the twenty-odd capinfos prints, plus the traffic over time. The rest
// - the hashes, the bit rate, the average packet size - answer questions
// nobody has yet when they open a capture, and `p` still shows all of them.
//
// The snapshot length is only shown when there is one. capinfos writes
// "file hdr: (not set)" otherwise, and a line saying a limit is not set is a
// line spent on nothing.
func fileFactsLines(info capinfo.Info, traffic []stats.IOStatRow) []statsLine {
	type fact struct{ label, value string }

	facts := []fact{
		{"Captured", info.Earliest},
		{"Duration", info.Duration},
		{"Packets", info.Packets},
		{"Size", info.Size},
	}
	if info.SnapLen != "" && !strings.Contains(info.SnapLen, "not set") {
		facts = append(facts, fact{"Packet limit", info.SnapLen})
	}
	if line := sparkline(traffic); line != "" {
		facts = append(facts, fact{"Traffic", line})
	}

	width := 0
	for _, f := range facts {
		if f.value != "" {
			width = max(width, len(f.label))
		}
	}

	res := make([]statsLine, 0, len(facts))
	for _, f := range facts {
		if f.value == "" {
			continue
		}
		res = append(res, statsLine{
			Text: fmt.Sprintf("  %-*s  %s", width, f.label, f.value),
		})
	}

	return res
}

// sparkline draws the traffic over time as one row of blocks, tallest bucket
// full height and an empty one left as a gap you can see.
//
// One line rather than a table of twenty. What is being asked here is a shape
// - did it all arrive in the last three seconds, is there a hole in the middle
// - and a shape reads better drawn than counted. The full numbers are one
// keypress away in the packet list.
//
// The blocks are eighths of a cell, which every terminal font that can draw a
// box can draw. A bucket that is not empty never renders as the shortest one
// unless it really is the smallest: rounding a lone packet down to nothing
// would turn "quiet" into "silent", and the gaps are the point.
func sparkline(rows []stats.IOStatRow) string {
	if len(rows) == 0 {
		return ""
	}

	most := 0
	for _, r := range rows {
		most = max(most, r.Frames)
	}
	if most == 0 {
		return ""
	}

	blocks := []rune("▁▂▃▄▅▆▇█")

	var b strings.Builder
	for _, r := range rows {
		if r.Frames == 0 {
			// Not a block at all: a gap has to look like one.
			b.WriteRune('·')
			continue
		}

		// Ceiling, so the quietest interval that carried anything still gets
		// the shortest block rather than none.
		i := (r.Frames*len(blocks) + most - 1) / most
		b.WriteRune(blocks[min(i, len(blocks))-1])
	}

	return b.String()
}

// rule draws the line under the column titles, as wide as the widest thing it
// sits over. Sizing it from the titles alone would leave it stopping short of
// most of the table, since the last column is free text.
func rule(titles string, rows []statsLine) string {
	w := len([]rune(titles))
	for _, r := range rows {
		if n := len([]rune(r.Text)); n > w {
			w = n
		}
	}
	return strings.Repeat("─", w)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//======================================================================

// statsHeading names the statistic and the filter it was produced under.
//
// The filter belongs in the title because both statistics honour it: a reader
// who has forgotten what is in the filter box would otherwise take a subset
// for the whole capture.
func statsHeading(name string, filter string, ignoresFilter bool) string {
	if filter == "" {
		return name
	}
	if ignoresFilter {
		// Saying "display filter: tcp" over a table computed from the whole
		// capture would be a plain untruth. Measured on credentials: tshark
		// takes the filter, exits zero and reports the same rows.
		return fmt.Sprintf("%s  ·  the whole capture: this one ignores the display filter", name)
	}
	return fmt.Sprintf("%s  ·  display filter: %s", name, filter)
}

// statsEmptyMessage explains an empty result, which tshark reports by printing
// absolutely nothing - no header, no empty table.
func statsEmptyMessage(filter string) string {
	if filter != "" {
		return "Nothing to report for this display filter."
	}
	return "Nothing to report for this capture."
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
