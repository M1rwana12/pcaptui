// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"strings"
)

//======================================================================

// ExpertRow is one line of `tshark -z expert`.
//
// The severity is the section the row appeared under, not a column of its own:
// tshark prints "Errors (6)", a rule, a header and then the rows, and repeats
// that for warnings, notes and chats. Reading a row without its section loses
// the only thing that says how much it matters.
type ExpertRow struct {
	Severity string // Errors, Warns, Notes, Chats
	Count    int
	Group    string // Sequence, Protocol, Malformed, Checksum, ...
	Protocol string
	Summary  string
}

// DisplayFilter is an expression matching the packets this row is about.
//
// tshark's expert tap reports how many times something happened and never
// which packets, so the row cannot carry frame numbers. What it can carry is
// the text, and Wireshark exposes that as a filterable field - so the row
// turns into a question the packet list can answer.
func (r ExpertRow) DisplayFilter() string {
	return fmt.Sprintf("_ws.expert.message == %s", quoteFilterString(r.Summary))
}

// quoteFilterString wraps s as a display filter string literal.
//
// Backslashes have to be doubled before the quotes are added: several expert
// summaries contain them literally - `HTTP/1.1 200 OK\r\n` is a chat message
// whose text really does end in those four characters - and an unescaped one
// makes tshark read the next character as an escape.
func quoteFilterString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + r.Replace(s) + `"`
}

//======================================================================

// ParseExpert reads the output of `tshark -z expert`.
//
// Where the Summary column starts is read off the header line rather than
// assumed, because the summary is free text - it contains spaces, brackets and
// backslashes - so it can only be found by position, and the position differs
// between Wireshark releases.
func ParseExpert(out string) []ExpertRow {
	var res []ExpertRow

	severity := ""
	var cols []int

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "="):
			continue

		case isExpertSection(trimmed):
			severity = expertSectionName(trimmed)
			cols = nil

		case strings.HasPrefix(trimmed, "Frequency"):
			cols = expertColumns(line)

		case cols != nil && severity != "":
			if row, ok := expertRow(line, cols, severity); ok {
				res = append(res, row)
			}
		}
	}

	return res
}

// isExpertSection recognises the "Errors (6)" line that opens a section.
func isExpertSection(line string) bool {
	return expertSectionName(line) != ""
}

func expertSectionName(line string) string {
	open := strings.Index(line, " (")
	if open <= 0 || !strings.HasSuffix(line, ")") {
		return ""
	}

	name := line[:open]
	switch name {
	case "Errors", "Warns", "Notes", "Chats", "Comments", "Details":
		return name
	}
	return ""
}

// expertColumns returns the start of each column, read off the header.
func expertColumns(header string) []int {
	var res []int
	for _, title := range []string{"Frequency", "Group", "Protocol", "Summary"} {
		i := strings.Index(header, title)
		if i < 0 {
			return nil
		}
		res = append(res, i)
	}
	return res
}

func expertRow(line string, cols []int, severity string) (ExpertRow, bool) {
	// Only the start of Summary is used as a position. The three fields before
	// it are right-aligned under their headings, so a value wider than its
	// column overflows leftwards into the padding of the one before - and
	// "Response Code" is a real expert group that does exactly that. Splitting
	// the middle by whitespace survives it: a protocol name is a display
	// filter token and can never contain a space, so it is the last word, and
	// whatever precedes it is the group.
	summary := strings.TrimSpace(slice(line, cols[3], len(line)))
	head := strings.TrimSpace(slice(line, 0, cols[3]))

	fields := strings.Fields(head)
	if len(fields) < 3 {
		return ExpertRow{}, false
	}

	// Through parseCount for the same reason as the other two tables: a
	// frequency in the thousands may or may not arrive grouped, depending on
	// which of Wireshark's printing routines produced it.
	n, ok := parseCount(fields[0])
	if !ok {
		return ExpertRow{}, false
	}

	return ExpertRow{
		Severity: severity,
		Count:    n,
		Group:    strings.Join(fields[1:len(fields)-1], " "),
		Protocol: fields[len(fields)-1],
		Summary:  summary,
	}, true
}

func slice(s string, from, to int) string {
	if from > len(s) {
		return ""
	}
	if to > len(s) {
		to = len(s)
	}
	if to < from {
		return ""
	}
	return s[from:to]
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
