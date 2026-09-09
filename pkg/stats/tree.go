// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"strings"
)

//======================================================================

// TreeRow is one line of a tshark packet-counter table - the shape shared by
// `-z http,tree`, `-z dns,tree` and the rest of the ",tree" statistics.
//
// Depth is how far the row is indented, which is the only place tshark says
// that "404 Not Found" belongs under "4xx: Client Error".
//
// Parent is the name of the row this one is indented under, because the same
// name means different things in different sections: "A" under Query Type is
// what was asked for, and "A" under Answer Type is what came back.
//
// Percent and Average are kept as tshark printed them. They are formatted
// numbers, and the format follows the machine's locale: this project has
// already been bitten by digit grouping that appears on one CI runner and not
// another, and the same table prints "33,33%" where another machine prints
// "33.33%". Parsing them into floats would mean guessing which convention
// produced them, to print them again immediately.
type TreeRow struct {
	Depth   int
	Name    string
	Parent  string
	Count   int
	Average string
	Percent string
}

//======================================================================

// ParseTree reads a tshark packet-counter table.
//
// The format is a rule, a title, a header, a rule, the rows, and a closing
// rule:
//
//	HTTP / Packet Counter:
//	Packet Type                   Count         Average  ...  Percent  ...
//	-----------------------------------------------------------------------
//	Total HTTP Packets            3                           100%
//	 HTTP Response Packets        3                           100,00%
//	  2xx: Success                1                           33,33%
//	   200 OK                     1                           100,00%
//
// The columns are fixed-width and the header names them, so the header is what
// says where Count and Percent begin. Reading them by field position instead
// would break on the first row whose name contains a space, which is most of
// them.
//
// The header is found by the two column names this reads, not by the name of
// the first column: Wireshark calls it "Packet Type" on some builds and
// "Topic / Item" on others, and CI found that difference between two runners
// on the same day. Count and Percent are the columns being read, so they are
// the right thing to require.
func ParseTree(out string) []TreeRow {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")

	var cols *treeColumns
	var res []TreeRow

	// The name of the innermost row seen at each depth, so that a row can be
	// told what it sits under.
	var enclosing []string

	for _, line := range lines {
		if cols == nil {
			cols = treeHeader(line)
			continue
		}

		row, ok := cols.row(line)
		if !ok {
			continue
		}

		for len(enclosing) <= row.Depth {
			enclosing = append(enclosing, "")
		}
		enclosing = enclosing[:row.Depth+1]
		enclosing[row.Depth] = row.Name
		if row.Depth > 0 {
			row.Parent = enclosing[row.Depth-1]
		}

		res = append(res, row)
	}

	return res
}

// treeColumns is where every column of the table begins, and which of them are
// the three this reads.
//
// All of them, not just those three, because a column has to be read between
// its own start and the next one's. Most rows leave Average blank, and reading
// from its start to the end of the line would take the first number after it -
// which is Rate, several columns away.
type treeColumns struct {
	starts                  []int
	count, average, percent int
}

// treeHeader recognises the header by the columns this reads, not by the name
// of the first one: Wireshark calls that "Packet Type" on some builds and
// "Topic / Item" on others, and CI found the two on the same day.
func treeHeader(line string) *treeColumns {
	starts := columnStarts(line)

	cols := &treeColumns{starts: starts, count: -1, average: -1, percent: -1}
	for i := range starts {
		switch strings.TrimSpace(cols.text(line, i)) {
		case "Count":
			cols.count = i
		case "Average":
			cols.average = i
		case "Percent":
			cols.percent = i
		}
	}

	if cols.count < 0 || cols.percent < 0 {
		return nil
	}
	return cols
}

// columnStarts is where each column begins, which a header states by putting
// two or more spaces between them. Single spaces are inside names - "Min Val",
// "Burst Rate", "Rate (ms)".
func columnStarts(header string) []int {
	var starts []int
	gap := 2

	for i, r := range header {
		if r == ' ' {
			gap++
			continue
		}
		if gap >= 2 {
			starts = append(starts, i)
		}
		gap = 0
	}

	return starts
}

// text is the i'th column of a line: from where that column starts to where
// the next one does.
func (c *treeColumns) text(line string, i int) string {
	if i < 0 || i >= len(c.starts) || len(line) <= c.starts[i] {
		return ""
	}
	end := len(line)
	if i+1 < len(c.starts) && c.starts[i+1] < end {
		end = c.starts[i+1]
	}
	return line[c.starts[i]:end]
}

func (c *treeColumns) row(line string) (TreeRow, bool) {
	at := c.starts[c.count]
	if len(line) <= at {
		return TreeRow{}, false
	}

	name := strings.TrimSpace(line[:at])
	if name == "" {
		return TreeRow{}, false
	}

	// A name long enough to reach into the Count column leaves no number to
	// read there, and the row is dropped rather than guessed at. tshark widens
	// the first column to fit its longest name, so this is not expected - but
	// a misread count is worse than a missing row.
	count, ok := parseCount(field(c.text(line, c.count)))
	if !ok {
		return TreeRow{}, false
	}

	return TreeRow{
		Depth:   len(line) - len(strings.TrimLeft(line, " ")),
		Name:    name,
		Count:   count,
		Average: number(c.text(line, c.average)),
		Percent: percent(c.text(line, c.percent)),
	}, true
}

// number is a formatted number, or "" where the row leaves the column blank -
// most rows leave Average blank, and tshark writes "-" for some of them.
func number(col string) string {
	s := field(col)
	if s == "-" {
		return ""
	}
	return s
}

func percent(col string) string {
	s := field(col)
	if !strings.HasSuffix(s, "%") {
		return ""
	}
	return s
}

// field is the first whitespace-separated word of s, or "".
func field(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

//======================================================================

// NonEmptyRows drops the rows tshark counted at nothing.
//
// These tables are fixed skeletons: every HTTP status class it knows about,
// every DNS statistic it can collect, present in the capture or not. On a
// capture with three responses that is four useful lines under sixteen empty
// ones, and the empty ones are what the eye lands on first. What is in the
// capture is the question being asked.
//
// A row is kept if anything under it was counted, even when it was not itself:
// tshark's DNS table has "Query Stats" at zero with the query statistics
// beneath it, and dropping the heading would leave its rows indented under
// nothing.
func NonEmptyRows(rows []TreeRow) []TreeRow {
	res := make([]TreeRow, 0, len(rows))

	for i, r := range rows {
		if r.Count > 0 || countedBelow(rows, i) {
			res = append(res, r)
		}
	}

	return res
}

// countedBelow is whether anything indented under rows[i] was counted. The
// rows that follow it more deeply indented are its subtree, and the first one
// at its own depth or shallower ends it.
func countedBelow(rows []TreeRow, i int) bool {
	for _, r := range rows[i+1:] {
		if r.Depth <= rows[i].Depth {
			return false
		}
		if r.Count > 0 {
			return true
		}
	}
	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
