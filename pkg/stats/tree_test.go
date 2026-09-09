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

// Real output of `tshark -z http,tree` on this project's demo capture, cut to
// the columns that matter. The decimal comma is not a typo: this is what the
// machine it was taken from printed, and it is why Percent is kept as text.
const httpTree = `
=========================================================
HTTP / Packet Counter:
Packet Type                   Count         Average       Percent       
---------------------------------------------------------
Total HTTP Packets            3                           100%          
 HTTP Response Packets        3                           100,00%       
  5xx: Server Error           1                           33,33%        
   500 Internal Server Error  1                           100,00%       
  4xx: Client Error           1                           33,33%        
   404 Not Found              1                           100,00%       
  2xx: Success                1                           33,33%        
   200 OK                     1                           100,00%       
  ???: broken                 0                           0,00%         
 HTTP Request Packets         0                           0,00%         
---------------------------------------------------------
`

func TestATreeTableIsReadInFull(t *testing.T) {
	rows := ParseTree(httpTree)

	require.Len(t, rows, 10)
	assert.Equal(t, "Total HTTP Packets", rows[0].Name)
	assert.Equal(t, 3, rows[0].Count)
	assert.Equal(t, "100%", rows[0].Percent)
}

// The indentation is the only place tshark says a 404 is one of the 4xx.
func TestIndentationBecomesDepthInATree(t *testing.T) {
	rows := ParseTree(httpTree)

	depths := map[string]int{}
	for _, r := range rows {
		depths[r.Name] = r.Depth
	}

	assert.Equal(t, 0, depths["Total HTTP Packets"])
	assert.Equal(t, 1, depths["HTTP Response Packets"])
	assert.Equal(t, 2, depths["4xx: Client Error"])
	assert.Equal(t, 3, depths["404 Not Found"])
}

// The rows tshark counted at zero are still rows; dropping them is the
// caller's decision, not the parser's.
func TestRowsCountedAtZeroAreStillParsed(t *testing.T) {
	rows := ParseTree(httpTree)

	var names []string
	for _, r := range rows {
		if r.Count == 0 {
			names = append(names, r.Name)
		}
	}

	assert.Equal(t, []string{"???: broken", "HTTP Request Packets"}, names)
}

// The columns are found from the header, so a table whose first column is
// wider is read the same way.
func TestAWiderFirstColumnIsStillRead(t *testing.T) {
	wide := `
Packet Type                                  Count         Percent       
-------------------------------------------------------------------------
Total DNS Packets                            17            100%          
 request-response time (msec)                4             23,52%        
`
	rows := ParseTree(wide)

	require.Len(t, rows, 2)
	assert.Equal(t, "Total DNS Packets", rows[0].Name)
	assert.Equal(t, 17, rows[0].Count)
	assert.Equal(t, "request-response time (msec)", rows[1].Name)
	assert.Equal(t, 4, rows[1].Count)
	assert.Equal(t, "23,52%", rows[1].Percent)
}

// Some of Wireshark's printing routines group thousands and some do not, and
// which one a machine uses has already differed between CI runners.
func TestAGroupedCountIsRead(t *testing.T) {
	grouped := `
Packet Type                   Count         Percent       
-----------------------------------------------------------
Total HTTP Packets            7,748         100%          
`
	rows := ParseTree(grouped)

	require.Len(t, rows, 1)
	assert.Equal(t, 7748, rows[0].Count)
}

// Everything before the header is preamble, and the closing rule is not a row.
func TestThePreambleAndRulesAreNotRows(t *testing.T) {
	for _, r := range ParseTree(httpTree) {
		assert.NotContains(t, r.Name, "=")
		assert.NotContains(t, r.Name, "---")
		assert.NotEqual(t, "HTTP / Packet Counter:", r.Name)
	}
}

// Output that is not a packet-counter table at all yields nothing, rather than
// rows made of whatever happened to be at those columns.
func TestOutputWithNoHeaderYieldsNothing(t *testing.T) {
	assert.Empty(t, ParseTree("tshark: some error\n"))
	assert.Empty(t, ParseTree(""))
}

// A name long enough to run into the Count column leaves no number to read,
// and a dropped row is better than a misread one.
func TestARowWithNoCountIsDropped(t *testing.T) {
	overflowing := `
Packet Type      Count         Percent       
---------------------------------------------
A name that is far too long to fit here       
Total HTTP Packets  3           100%          
`
	rows := ParseTree(overflowing)

	for _, r := range rows {
		assert.NotContains(t, r.Name, "far too long")
	}
}

// Windows line endings come from the same tshark on the same capture.
func TestCarriageReturnsDoNotReachTheRows(t *testing.T) {
	rows := ParseTree(strings.ReplaceAll(httpTree, "\n", "\r\n"))

	require.NotEmpty(t, rows)
	for _, r := range rows {
		assert.NotContains(t, r.Name, "\r")
		assert.NotContains(t, r.Percent, "\r")
	}
}

// The first column has two names in the wild: Wireshark prints "Packet Type"
// on some builds and "Topic / Item" on others. CI found the difference between
// two runners on the same day, so the header is recognised by the columns this
// reader actually uses rather than by the name of the one it does not.
func TestTheOtherNameForTheFirstColumnIsAlsoAHeader(t *testing.T) {
	ubuntu := `
=========================================================
HTTP/Packet Counter:
Topic / Item                  Count         Percent       
---------------------------------------------------------
Total HTTP Packets            3             100%          
`
	rows := ParseTree(ubuntu)

	require.Len(t, rows, 1)
	assert.Equal(t, "Total HTTP Packets", rows[0].Name)
	assert.Equal(t, 3, rows[0].Count)
	assert.Equal(t, "100%", rows[0].Percent)
}

//======================================================================

// Real output of `tshark -z dns,tree`, cut to the columns that matter. Blank
// Average cells with numbers further along the line are the point: reading
// Average from its column to the end of the line would take the Rate.
const dnsTree = `
==============================================================================
DNS:
Packet Type                    Count         Average       Min Val       Rate (ms)     Percent       
------------------------------------------------------------------------------
Total Packets                  4                                         0,0013        100%          
rcode                          4                                         0,0013        100%          
 No error                      3                                         0,0010        75,00%        
 No such name                  1                                         0,0003        25,00%        
Query Stats                    0                                         0,0000        100%          
 Qname Len                     2             12,00         12            0,0007                      
 Label Stats                   0                                         0,0000                      
  2nd Level                    2                                         0,0007                      
Service Stats                  0                                         0,0000        100%          
 no. of retransmissions        0                                         0,0000                      
------------------------------------------------------------------------------
`

// A row that leaves Average blank has to come back blank. The columns after it
// are not empty, and a reader that scanned to the end of the line would report
// the Rate as the average.
func TestABlankAverageIsNotTheNextColumn(t *testing.T) {
	rows := ParseTree(dnsTree)
	require.NotEmpty(t, rows)

	byName := map[string]TreeRow{}
	for _, r := range rows {
		byName[r.Name] = r
	}

	assert.Equal(t, "", byName["Total Packets"].Average)
	assert.Equal(t, "", byName["No error"].Average)
	assert.Equal(t, "12,00", byName["Qname Len"].Average)
}

// The same name appears in more than one section of the DNS table, so a row
// has to know what it sits under before anything can be said about it.
func TestARowKnowsWhatItSitsUnder(t *testing.T) {
	rows := ParseTree(dnsTree)

	parents := map[string]string{}
	for _, r := range rows {
		parents[r.Name] = r.Parent
	}

	assert.Equal(t, "", parents["Total Packets"], "a top-level row has no parent")
	assert.Equal(t, "rcode", parents["No such name"])
	assert.Equal(t, "Query Stats", parents["Qname Len"])
	assert.Equal(t, "Label Stats", parents["2nd Level"])
}

// Leaving a section and entering the next one has to forget the first one's
// rows: "Service Stats" is not inside "Query Stats".
func TestLeavingASectionForgetsIt(t *testing.T) {
	rows := ParseTree(dnsTree)

	for _, r := range rows {
		if r.Name == "no. of retransmissions" {
			assert.Equal(t, "Service Stats", r.Parent)
			return
		}
	}
	t.Fatal("the row was not parsed at all")
}

//======================================================================

// tshark's tables are fixed skeletons of everything it can count, so most of
// any one of them is zeroes on a real capture.
func TestTheRowsCountedAtNothingAreDropped(t *testing.T) {
	rows := NonEmptyRows(ParseTree(dnsTree))

	var names []string
	for _, r := range rows {
		names = append(names, r.Name)
	}

	assert.NotContains(t, names, "Service Stats")
	assert.NotContains(t, names, "no. of retransmissions")
}

// A heading tshark counts at zero still has to stay when its own rows were
// counted, or they end up indented under nothing.
func TestAZeroHeadingWithSomethingUnderItStays(t *testing.T) {
	rows := NonEmptyRows(ParseTree(dnsTree))

	var names []string
	for _, r := range rows {
		names = append(names, r.Name)
	}

	assert.Contains(t, names, "Query Stats", "its query statistics were counted")
	assert.Contains(t, names, "Label Stats", "2nd Level under it was counted")
	assert.Contains(t, names, "2nd Level")
}

func TestATableOfNothingButZeroesIsDroppedEntirely(t *testing.T) {
	rows := []TreeRow{
		{Depth: 0, Name: "Total HTTP Packets", Count: 0},
		{Depth: 1, Name: "HTTP Response Packets", Count: 0},
	}

	assert.Empty(t, NonEmptyRows(rows))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
