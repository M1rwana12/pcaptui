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
// Percent is kept as tshark printed it. It is a formatted number, and the
// format follows the machine's locale: this project has already been bitten
// by digit grouping that appears on one CI runner and not another, and the
// same table prints "33,33%" where another machine prints "33.33%". Parsing it
// into a float would mean guessing which convention produced it, to display it
// again immediately.
type TreeRow struct {
	Depth   int
	Name    string
	Count   int
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

	countCol, pctCol := -1, -1
	var res []TreeRow

	for _, line := range lines {
		if countCol < 0 {
			if c, p := strings.Index(line, "Count"), strings.Index(line, "Percent"); c >= 0 && p > c {
				countCol, pctCol = c, p
			}
			continue
		}

		row, ok := treeRow(line, countCol, pctCol)
		if ok {
			res = append(res, row)
		}
	}

	return res
}

func treeRow(line string, countCol int, pctCol int) (TreeRow, bool) {
	if len(line) <= countCol {
		return TreeRow{}, false
	}

	name := strings.TrimSpace(line[:countCol])
	if name == "" {
		return TreeRow{}, false
	}

	// A name long enough to reach into the Count column leaves no number to
	// read here, and the row is dropped rather than guessed at. tshark widens
	// the first column to fit its longest name, so this is not expected - but
	// a misread count is worse than a missing row.
	count, ok := parseCount(field(line[countCol:]))
	if !ok {
		return TreeRow{}, false
	}

	return TreeRow{
		Depth:   len(line) - len(strings.TrimLeft(line, " ")),
		Name:    name,
		Count:   count,
		Percent: percentField(line, pctCol),
	}, true
}

func percentField(line string, pctCol int) string {
	if pctCol < 0 || len(line) <= pctCol {
		return ""
	}
	pct := field(line[pctCol:])
	if !strings.HasSuffix(pct, "%") {
		return ""
	}
	return pct
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
// Local Variables:
// mode: Go
// fill-column: 78
// End:
