// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Verbatim from `tshark -q -z expert -r scripts/pcaps/telnet-cooked.pcap`,
// Wireshark 4.6.8. Every shape the format has is here: three sections, a
// two-word group, a count above one, and summaries containing spaces and
// brackets.
const expertOutput = `
Errors (6)
=============
   Frequency      Group           Protocol  Summary
           5   Protocol               IPv4  IPv4 total length exceeds packet length (52 bytes)
           1  Malformed             TELNET  Malformed Packet (Exception occurred)

Notes (5)
=============
   Frequency      Group           Protocol  Summary
           1   Protocol                TCP  The SYN packet does not contain a SACK PERM option
           1   Sequence                TCP  This frame is a (suspected) retransmission

Chats (4)
=============
   Frequency      Group           Protocol  Summary
           1   Sequence                TCP  Connection establish request (SYN): server port 23
           2   Sequence                TCP  Connection finish (FIN)
`

func TestEveryRowIsRead(t *testing.T) {
	rows := ParseExpert(expertOutput)

	require.Len(t, rows, 6)
}

// The severity is the section heading, not a column. A row read without it
// says something happened and not whether it matters.
func TestARowCarriesTheSectionItWasUnder(t *testing.T) {
	rows := ParseExpert(expertOutput)

	assert.Equal(t, "Errors", rows[0].Severity)
	assert.Equal(t, "Errors", rows[1].Severity)
	assert.Equal(t, "Notes", rows[2].Severity)
	assert.Equal(t, "Chats", rows[4].Severity)
}

func TestTheColumnsOfARow(t *testing.T) {
	rows := ParseExpert(expertOutput)

	assert.Equal(t, 5, rows[0].Count)
	assert.Equal(t, "Protocol", rows[0].Group)
	assert.Equal(t, "IPv4", rows[0].Protocol)
	assert.Equal(t, "IPv4 total length exceeds packet length (52 bytes)", rows[0].Summary)
}

// The summary is free text and the only reliable way to find it is the column
// the header puts it in - it contains spaces, brackets and, for some HTTP
// chats, backslashes.
func TestASummaryWithSpacesAndBracketsSurvivesWhole(t *testing.T) {
	rows := ParseExpert(expertOutput)

	assert.Equal(t, "Connection establish request (SYN): server port 23", rows[4].Summary)
}

// "Response Code" is a real Wireshark expert group. It is wider than the
// column the header allots it, so it overflows leftwards into the frequency
// padding, and anything splitting the head on fixed offsets would cut it in
// half.
//
// This row is constructed rather than captured, so the test first proves it is
// aligned the way tshark aligns one - otherwise it would be asserting against
// a format that does not exist.
func TestATwoWordGroupIsNotSplit(t *testing.T) {
	const header = "   Frequency      Group           Protocol  Summary"
	const row = "           1  Response Code       HTTP      HTTP/1.1 404 Not Found"

	require.Equal(t, strings.Index(header, "Summary"), strings.Index(row, "HTTP/1.1"),
		"the constructed row does not line up with the header")

	rows := ParseExpert("\nWarns (1)\n=============\n" + header + "\n" + row + "\n")

	require.Len(t, rows, 1)
	assert.Equal(t, "Response Code", rows[0].Group)
	assert.Equal(t, "HTTP", rows[0].Protocol)
	assert.Equal(t, "HTTP/1.1 404 Not Found", rows[0].Summary)
}

// A statistic narrowed by a filter that matches nothing prints nothing at all.
func TestNoOutputIsNoRowsRatherThanOneBadRow(t *testing.T) {
	assert.Empty(t, ParseExpert(""))
	assert.Empty(t, ParseExpert("\n\n"))
}

func TestTheRuleAndTheHeaderAreNotRows(t *testing.T) {
	out := `
Notes (1)
=============
   Frequency      Group           Protocol  Summary
           1   Sequence                TCP  This frame is a (suspected) retransmission
`
	rows := ParseExpert(out)

	require.Len(t, rows, 1)
	assert.Equal(t, "This frame is a (suspected) retransmission", rows[0].Summary)
}

//======================================================================

func TestARowBecomesAFilterForItsOwnPackets(t *testing.T) {
	r := ExpertRow{Summary: "This frame is a (suspected) retransmission"}

	assert.Equal(t,
		`_ws.expert.message == "This frame is a (suspected) retransmission"`,
		r.DisplayFilter())
}

// `HTTP/1.1 200 OK\r\n` is a real chat summary: those are four literal
// characters, not a line ending. Left unescaped, tshark reads the backslash as
// starting an escape and the filter means something else.
func TestABackslashInASummaryIsEscaped(t *testing.T) {
	r := ExpertRow{Summary: `HTTP/1.1 200 OK\r\n`}

	assert.Equal(t,
		`_ws.expert.message == "HTTP/1.1 200 OK\\r\\n"`,
		r.DisplayFilter())
}

func TestAQuoteInASummaryIsEscaped(t *testing.T) {
	r := ExpertRow{Summary: `unknown "value"`}

	assert.Equal(t, `_ws.expert.message == "unknown \"value\""`, r.DisplayFilter())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
