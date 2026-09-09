// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"regexp"
	"strings"
)

//======================================================================

// CollapseNumbered folds the rows tshark numbers into one row each.
//
// Wireshark reports some expert items once per occurrence with a counter in
// the text - "Duplicate ACK (#1)", "Duplicate ACK (#2)" - and the tap reports
// each of those as a separate kind of problem. Measured on a capture of
// 376,832 packets: `tshark -z expert` printed 4,108 data rows, of which 4,095
// were duplicate ACKs. Thirteen facts, arriving as four thousand lines, with
// the four that matter at the top and no way to see that from the summary,
// which said "and 4104 more".
//
// The rows are merged in place: the first of a family keeps its position, so
// the order tshark chose - worst first - survives.
//
// The count is the sum, and it is the right number: on that same capture the
// merged row said 143,325 and the filter it offers found exactly 143,325
// packets.
func CollapseNumbered(rows []ExpertRow) []ExpertRow {
	res := make([]ExpertRow, 0, len(rows))
	at := map[expertFamily]int{}

	for _, r := range rows {
		prefix, num, ok := splitNumbered(r.Summary)
		if !ok {
			res = append(res, r)
			continue
		}

		key := expertFamily{
			severity: r.Severity,
			group:    r.Group,
			protocol: r.Protocol,
			prefix:   prefix,
		}

		i, seen := at[key]
		if !seen {
			r.Summary = prefix
			r.Merged = 1
			r.FirstNum, r.LastNum = num, num
			at[key] = len(res)
			res = append(res, r)
			continue
		}

		res[i].Count += r.Count
		res[i].Merged++
		if num < res[i].FirstNum {
			res[i].FirstNum = num
		}
		if num > res[i].LastNum {
			res[i].LastNum = num
		}
	}

	// A family of one was never folded, so it gets its number back. Without
	// this it kept the shortened text and offered a filter for a message that
	// does not exist: "Duplicate ACK" alone matches nothing, because what
	// tshark wrote was "Duplicate ACK (#4)".
	for _, i := range at {
		if res[i].Merged == 1 {
			res[i].Summary = fmt.Sprintf("%s (#%d)", res[i].Summary, res[i].FirstNum)
			res[i].Merged = 0
			res[i].FirstNum, res[i].LastNum = 0, 0
		}
	}

	return res
}

// expertFamily is what makes two numbered rows the same problem. The severity,
// group and protocol are part of it because the same wording under a different
// heading is a different thing.
type expertFamily struct {
	severity string
	group    string
	protocol string
	prefix   string
}

// splitNumbered pulls "Duplicate ACK (#12)" apart into "Duplicate ACK" and 12.
func splitNumbered(summary string) (prefix string, num int, ok bool) {
	m := numberedSummary.FindStringSubmatch(summary)
	if m == nil {
		return "", 0, false
	}

	n, ok := parseCount(m[2])
	if !ok {
		return "", 0, false
	}

	return m[1], n, true
}

var numberedSummary = regexp.MustCompile(`^(.*[^ ]) \(#(\d+)\)$`)

//======================================================================

// mergedFilter is the display filter for a folded row: every numbered variant
// of the same text and nothing else.
//
// The brackets are written as character classes rather than escaped, because
// Wireshark's display filter refuses a backslash escape it does not know
// inside a string literal - `matches "^Duplicate ACK \(#[0-9]+\)$"` is
// rejected with "\( is not a valid character escape sequence", while
// `[(]` is accepted and matches the same thing.
//
// A prefix carrying a backslash, a caret or a quote is not folded at all, and
// the caller keeps tshark's own rows: a caret cannot be written as a one
// character class, and the other two would need the escapes this cannot use.
// No Wireshark expert message contains them, so this is a guard rather than a
// case.
func mergedFilter(prefix string) (string, bool) {
	re, ok := regexLiteral(prefix)
	if !ok {
		return "", false
	}

	return fmt.Sprintf(`_ws.expert.message matches "^%s [(]#[0-9]+[)]$"`, re), true
}

// regexLiteral turns text into a regex matching exactly that text.
func regexLiteral(s string) (string, bool) {
	var b strings.Builder

	for _, r := range s {
		switch r {
		case '\\', '^', '"':
			return "", false
		case '.', '$', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|':
			// A one-character class is literal for all of these, and needs no
			// escape the filter parser might reject.
			b.WriteString("[" + string(r) + "]")
		default:
			b.WriteRune(r)
		}
	}

	return b.String(), true
}

//======================================================================

// SeverityTotal is how many expert items of one severity the whole capture
// has, as tshark states it in each section heading: "Notes (430012)".
//
// The number was being read and thrown away - expertSectionName kept the word
// and dropped what was in the brackets - so the program knew a capture had
// 430,012 note-level events and printed nothing resembling that anywhere.
type SeverityTotal struct {
	Severity string
	Count    int
}

// ExpertTotals reads the section headings of `tshark -z expert`, in the order
// tshark printed them, which is worst first.
func ExpertTotals(out string) []SeverityTotal {
	var res []SeverityTotal

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)

		name := expertSectionName(trimmed)
		if name == "" {
			continue
		}

		n, ok := parseCount(strings.TrimSuffix(trimmed[len(name)+2:], ")"))
		if !ok {
			continue
		}

		res = append(res, SeverityTotal{Severity: name, Count: n})
	}

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
