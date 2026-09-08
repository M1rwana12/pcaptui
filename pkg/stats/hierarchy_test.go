// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Verbatim from `tshark -q -z io,phs -r scripts/pcaps/demo.pcap`,
// Wireshark 4.6.8.
const hierarchyOutput = `
===================================================================
Protocol Hierarchy Statistics
Filter:

frame                                    frames:7 bytes:1027
  eth                                    frames:7 bytes:1027
    ip                                   frames:7 bytes:1027
      tcp                                frames:7 bytes:1027
        http                             frames:3 bytes:454
          data-text-lines                frames:3 bytes:454
===================================================================
`

func TestEveryProtocolIsARow(t *testing.T) {
	rows := ParseHierarchy(hierarchyOutput)

	require.Len(t, rows, 6)
	assert.Equal(t, "frame", rows[0].Protocol)
	assert.Equal(t, "data-text-lines", rows[5].Protocol)
}

func TestTheCountsAreRead(t *testing.T) {
	rows := ParseHierarchy(hierarchyOutput)

	assert.Equal(t, 7, rows[0].Frames)
	assert.Equal(t, 1027, rows[0].Bytes)
	assert.Equal(t, 3, rows[4].Frames)
	assert.Equal(t, 454, rows[4].Bytes)
}

// The indentation is the only place tshark says what carries what. Losing it
// turns a tree into a list and the answer to "what is inside what" with it.
func TestIndentationBecomesDepth(t *testing.T) {
	rows := ParseHierarchy(hierarchyOutput)

	assert.Equal(t, 0, rows[0].Depth) // frame
	assert.Equal(t, 1, rows[1].Depth) // eth
	assert.Equal(t, 2, rows[2].Depth) // ip
	assert.Equal(t, 3, rows[3].Depth) // tcp
	assert.Equal(t, 4, rows[4].Depth) // http
	assert.Equal(t, 5, rows[5].Depth) // data-text-lines
}

// The rule, the title and the "Filter:" line are not protocols. The rule is
// printed both before and after the table, so position cannot be relied on.
func TestThePreambleAndTheClosingRuleAreNotRows(t *testing.T) {
	for _, line := range []string{
		"===================================================================",
		"Protocol Hierarchy Statistics",
		"Filter: tcp.port == 80",
		"",
		"   ",
	} {
		assert.Empty(t, ParseHierarchy(line), "should not have read a row from %q", line)
	}
}

func TestNoOutputIsNoRows(t *testing.T) {
	assert.Empty(t, ParseHierarchy(""))
}

//======================================================================

// The name in this table is already the protocol's display filter name, which
// is what makes "what is in this file" one keypress away from "show me that".
func TestARowFiltersOnItsOwnProtocol(t *testing.T) {
	assert.Equal(t, "http", HierarchyRow{Protocol: "http"}.DisplayFilter())
	assert.Equal(t, "data-text-lines", HierarchyRow{Protocol: "data-text-lines"}.DisplayFilter())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
