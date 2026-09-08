// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/m1rwana12/pcaptui/pkg/stats"
)

//======================================================================

// expertLines lays out Expert Information.
//
// The severity is a column here rather than a section heading, because the
// rows have to be one flat list for any of them to be selectable, and a row
// that has floated away from its heading has to carry its severity with it.
// Order is kept as tshark gives it, which is worst first.
func expertLines(rows []stats.ExpertRow) statsView {
	if len(rows) == 0 {
		return statsView{}
	}

	sevw, grpw, prow, cntw := len("Severity"), len("Group"), len("Protocol"), len("Count")
	for _, r := range rows {
		sevw = max(sevw, len(severityLabel(r.Severity)))
		grpw = max(grpw, len(r.Group))
		prow = max(prow, len(r.Protocol))
		cntw = max(cntw, len(fmt.Sprintf("%d", r.Count)))
	}

	line := func(sev, grp, pro, cnt, sum string) string {
		return fmt.Sprintf("%-*s  %-*s  %-*s  %*s  %s",
			sevw, sev, grpw, grp, prow, pro, cntw, cnt, sum)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(severityLabel(r.Severity), r.Group, r.Protocol,
				fmt.Sprintf("%d", r.Count), r.Summary),
			Filter: r.DisplayFilter(),
		})
	}

	titles := line("Severity", "Group", "Protocol", "Count", "Summary")
	v.Header = []string{titles, rule(titles, v.Rows)}

	return v
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
		framew = max(framew, len(fmt.Sprintf("%d", r.Frames)))
	}

	line := func(name, frames, bytes string) string {
		return fmt.Sprintf("%-*s  %*s  %s", namew, name, framew, frames, bytes)
	}

	var v statsView
	for _, r := range rows {
		v.Rows = append(v.Rows, statsLine{
			Text: line(
				strings.Repeat(" ", 2*r.Depth)+r.Protocol,
				fmt.Sprintf("%d", r.Frames),
				fmt.Sprintf("%d", r.Bytes),
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
	rows = stats.HTTPRows(rows)
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

// overviewLimit is how many rows of each section the overview shows.
//
// It is a summary, not a table: past a handful of rows the eye stops reading
// and starts scrolling, and each section has its own view a keypress away for
// when the whole thing is wanted.
const overviewLimit = 4

// overviewLines is the three statistics as one screen.
//
// Order is the order the questions get asked: what is wrong with this capture,
// what is in it, who is on the wire. Problems first because that is what
// someone handed a capture is looking for, and because a capture with none is
// itself worth knowing in one line.
//
// Every row keeps the filter its own table would have given it, so the summary
// is not a dead end - enter still lands on the packets.
func overviewLines(out string) statsView {
	var v statsView

	add := func(title string, body statsView, empty string) {
		if len(v.Rows) > 0 {
			v.Rows = append(v.Rows, statsLine{})
		}
		v.Rows = append(v.Rows, statsLine{Text: title}, statsLine{Text: rule(title, nil)})

		if len(body.Rows) == 0 {
			v.Rows = append(v.Rows, statsLine{Text: "  " + empty})
			return
		}

		shown := body.Rows
		if len(shown) > overviewLimit {
			shown = shown[:overviewLimit]
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

	add("What is wrong", expertLines(stats.ParseExpert(out)),
		"Nothing the dissectors object to.")
	add("What is in it", hierarchyLines(stats.ParseHierarchy(out)),
		"No protocols reported.")
	add("Who is on the wire", endpointLines(stats.ParseEndpoints(out)),
		"No IPv4 addresses.")

	return v
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
func statsHeading(name string, filter string) string {
	if filter == "" {
		return name
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
