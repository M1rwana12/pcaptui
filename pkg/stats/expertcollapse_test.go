// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func numbered(n int, count int) ExpertRow {
	return ExpertRow{
		Severity: "Notes",
		Count:    count,
		Group:    "Sequence",
		Protocol: "TCP",
		Summary:  fmt.Sprintf("Duplicate ACK (#%d)", n),
	}
}

// The case this exists for: a capture of 376,832 packets gave 4,108 rows, of
// which 4,095 were numbered duplicate ACKs - thirteen facts arriving as four
// thousand lines.
func TestANumberedFamilyBecomesOneRow(t *testing.T) {
	var rows []ExpertRow
	for i := 1; i <= 4095; i++ {
		rows = append(rows, numbered(i, 35))
	}

	got := CollapseNumbered(rows)

	require.Len(t, got, 1)
	assert.Equal(t, "Duplicate ACK", got[0].Summary)
	assert.Equal(t, 4095, got[0].Merged)
	assert.Equal(t, 1, got[0].FirstNum)
	assert.Equal(t, 4095, got[0].LastNum)
	assert.Equal(t, 4095*35, got[0].Count, "the count is the sum of what was folded")
}

// A row that is not numbered is not touched, and keeps a Merged of zero so the
// view can tell the two apart.
func TestAnOrdinaryRowPassesThrough(t *testing.T) {
	rows := []ExpertRow{
		{Severity: "Errors", Count: 20480, Group: "Protocol", Protocol: "IPv4",
			Summary: "IPv4 total length exceeds packet length (52 bytes)"},
	}

	got := CollapseNumbered(rows)

	assert.Equal(t, rows, got)
	assert.Zero(t, got[0].Merged)
}

// The order tshark chose is worst first, so a folded row has to stay where its
// first member was rather than moving to the end.
func TestTheFirstOfAFamilyKeepsItsPlace(t *testing.T) {
	rows := []ExpertRow{
		{Severity: "Errors", Count: 1, Group: "Malformed", Protocol: "TELNET", Summary: "Malformed Packet"},
		numbered(1, 35),
		{Severity: "Notes", Count: 200656, Group: "Sequence", Protocol: "TCP",
			Summary: "This frame is a (suspected) retransmission"},
		numbered(2, 35),
	}

	got := CollapseNumbered(rows)

	require.Len(t, got, 3)
	assert.Equal(t, "Malformed Packet", got[0].Summary)
	assert.Equal(t, "Duplicate ACK", got[1].Summary)
	assert.Equal(t, "This frame is a (suspected) retransmission", got[2].Summary)
}

// The same wording under a different heading is a different thing.
func TestTheSameTextInAnotherSectionIsNotTheSameFamily(t *testing.T) {
	a := numbered(1, 35)
	b := numbered(2, 35)
	b.Severity = "Warns"
	c := numbered(3, 35)
	c.Protocol = "UDP"
	d := numbered(4, 35)
	d.Group = "Protocol"

	got := CollapseNumbered([]ExpertRow{a, b, c, d})

	assert.Len(t, got, 4)
}

// Numbers need not start at one or arrive in order; the label has to say what
// was really covered.
func TestTheRangeIsTheLowestAndHighestSeen(t *testing.T) {
	got := CollapseNumbered([]ExpertRow{numbered(7, 1), numbered(3, 1), numbered(9, 1)})

	require.Len(t, got, 1)
	assert.Equal(t, 3, got[0].FirstNum)
	assert.Equal(t, 9, got[0].LastNum)
}

// A hash in the middle, or a number that is not one, is not a family.
func TestOnlyATrailingNumberInBracketsCounts(t *testing.T) {
	for _, summary := range []string{
		"Duplicate ACK (#1) and more",
		"Duplicate ACK (#)",
		"Duplicate ACK (#one)",
		"Duplicate ACK #1",
		"(#1)",
	} {
		got := CollapseNumbered([]ExpertRow{{Summary: summary, Count: 1}})

		require.Len(t, got, 1, summary)
		assert.Equal(t, summary, got[0].Summary, "%q should be left alone", summary)
		assert.Zero(t, got[0].Merged, summary)
	}
}

func TestNoRowsFoldToNoRows(t *testing.T) {
	assert.Empty(t, CollapseNumbered(nil))
}

//======================================================================

// A folded row stands for every numbered variant, so its filter has to select
// all of them - and this is the exact string that found 143,325 packets on the
// capture the row was measured against.
func TestAFoldedRowFiltersOnTheWholeFamily(t *testing.T) {
	got := CollapseNumbered([]ExpertRow{numbered(1, 35), numbered(2, 35)})

	require.Len(t, got, 1)
	assert.Equal(t,
		`_ws.expert.message matches "^Duplicate ACK [(]#[0-9]+[)]$"`,
		got[0].DisplayFilter())
}

// A family of one was never a family. It gets its number back, because the
// shortened text names a message that does not exist - "Duplicate ACK" alone
// matches nothing, and a filter that finds nothing is worse than no folding.
func TestASingleNumberedRowKeepsItsNumber(t *testing.T) {
	got := CollapseNumbered([]ExpertRow{numbered(4, 35)})

	require.Len(t, got, 1)
	assert.Zero(t, got[0].Merged)
	assert.Equal(t, "Duplicate ACK (#4)", got[0].Summary)
	assert.Equal(t, `_ws.expert.message == "Duplicate ACK (#4)"`, got[0].DisplayFilter())
}

// Wireshark's display filter rejects "\(" inside a string literal outright, so
// the brackets are written as one-character classes instead.
func TestRegexSpecialsBecomeCharacterClasses(t *testing.T) {
	got, ok := regexLiteral("a.b*c(d)e[f]g{h}i|j+k?l$m")

	require.True(t, ok)
	assert.Equal(t, "a[.]b[*]c[(]d[)]e[[]f[]]g[{]h[}]i[|]j[+]k[?]l[$]m", got)
}

// A caret cannot be written as a one-character class, and a backslash or a
// quote would need the escapes this filter language will not take. Such a row
// is not folded at all and keeps tshark's own rows.
func TestTextThatCannotBeWrittenAsARegexIsNotFolded(t *testing.T) {
	for _, s := range []string{`a^b`, `a\b`, `a"b`} {
		_, ok := regexLiteral(s)
		assert.False(t, ok, "%q", s)

		_, ok = mergedFilter(s)
		assert.False(t, ok, "%q", s)
	}
}

// If the text cannot become a regex, the row still offers something true: the
// one message it can name exactly.
func TestAnUnwritableFamilyStillOffersOneExactMessage(t *testing.T) {
	a := ExpertRow{Severity: "Notes", Group: "G", Protocol: "P", Count: 1, Summary: `odd ^ thing (#5)`}
	b := ExpertRow{Severity: "Notes", Group: "G", Protocol: "P", Count: 1, Summary: `odd ^ thing (#6)`}

	got := CollapseNumbered([]ExpertRow{a, b})

	require.Len(t, got, 1)
	assert.Equal(t, `_ws.expert.message == "odd ^ thing (#5)"`, got[0].DisplayFilter())
}

//======================================================================

// tshark states these in its section headings and the program read them and
// threw them away.
func TestTheSeverityTotalsAreRead(t *testing.T) {
	out := `
Errors (20481)
=============
   Frequency      Group           Protocol  Summary
       20480   Protocol               IPv4  IPv4 total length exceeds packet length (52 bytes)

Notes (430012)
=============
   Frequency      Group           Protocol  Summary
      200656   Sequence                TCP  This frame is a (suspected) retransmission

Chats (16384)
=============
   Frequency      Group           Protocol  Summary
        4096   Sequence                TCP  Connection establish request (SYN): server port 23
`
	got := ExpertTotals(out)

	assert.Equal(t, []SeverityTotal{
		{Severity: "Errors", Count: 20481},
		{Severity: "Notes", Count: 430012},
		{Severity: "Chats", Count: 16384},
	}, got)
}

// A total in the thousands may arrive grouped, like every other count tshark
// prints.
func TestAGroupedTotalIsRead(t *testing.T) {
	got := ExpertTotals("Notes (430,012)\n")

	require.Len(t, got, 1)
	assert.Equal(t, 430012, got[0].Count)
}

func TestOutputWithNoSectionsHasNoTotals(t *testing.T) {
	assert.Empty(t, ExpertTotals(""))
	assert.Empty(t, ExpertTotals("tshark: something went wrong\n"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
