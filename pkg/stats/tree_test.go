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
// Local Variables:
// mode: Go
// fill-column: 78
// End:
