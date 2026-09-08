// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/stats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func someExpertRows() []stats.ExpertRow {
	return []stats.ExpertRow{
		{Severity: "Errors", Count: 5, Group: "Protocol", Protocol: "IPv4",
			Summary: "IPv4 total length exceeds packet length (52 bytes)"},
		{Severity: "Errors", Count: 1, Group: "Malformed", Protocol: "TELNET",
			Summary: "Malformed Packet (Exception occurred)"},
		{Severity: "Notes", Count: 1, Group: "Sequence", Protocol: "TCP",
			Summary: "This frame is a (suspected) retransmission"},
	}
}

func TestEveryExpertRowIsSelectable(t *testing.T) {
	v := expertLines(someExpertRows())

	require.Len(t, v.Rows, 3)
	for i, r := range v.Rows {
		assert.True(t, r.actionable(), "row %d cannot be acted on", i)
	}
}

// tshark prints the severity once, as a heading over a section. Flattening the
// sections into one list is what makes the rows selectable, so each row has to
// carry its own severity or the flattening loses it.
func TestASeverityTravelsWithItsRow(t *testing.T) {
	v := expertLines(someExpertRows())

	assert.Contains(t, v.Rows[0].Text, "Error")
	assert.Contains(t, v.Rows[2].Text, "Note")
}

// "Errors" over a column of single rows reads as a tally. The column says what
// one row is.
func TestTheSeverityColumnIsSingular(t *testing.T) {
	v := expertLines(someExpertRows())

	assert.NotContains(t, v.Rows[0].Text, "Errors")
	assert.NotContains(t, v.Rows[2].Text, "Notes")
}

func TestTheColumnTitlesAreAHeaderNotARow(t *testing.T) {
	v := expertLines(someExpertRows())

	require.Len(t, v.Header, 2)
	assert.Contains(t, v.Header[0], "Severity")
	assert.Contains(t, v.Header[0], "Summary")

	for _, r := range v.Rows {
		assert.NotContains(t, r.Text, "Severity")
	}
}

func TestColumnsLineUp(t *testing.T) {
	v := expertLines(someExpertRows())

	at := strings.Index(v.Rows[0].Text, "IPv4 total length")
	require.Positive(t, at)
	assert.Equal(t, at, strings.Index(v.Rows[1].Text, "Malformed Packet"),
		"the summaries start in different columns")
	assert.Equal(t, at, strings.Index(v.Header[0], "Summary"),
		"the summaries do not start where their title does")
}

// A rule that stops after the titles leaves most of the table with nothing
// under it, because the last column is free text and much the widest.
func TestTheRuleSpansTheWidestLine(t *testing.T) {
	v := expertLines(someExpertRows())

	widest := len([]rune(v.Header[0]))
	for _, r := range v.Rows {
		if n := len([]rune(r.Text)); n > widest {
			widest = n
		}
	}

	assert.Equal(t, widest, len([]rune(v.Header[1])))
}

func TestNoRowsIsAnEmptyView(t *testing.T) {
	assert.True(t, expertLines(nil).empty())
	assert.True(t, hierarchyLines(nil).empty())
}

//======================================================================

func someHierarchyRows() []stats.HierarchyRow {
	return []stats.HierarchyRow{
		{Depth: 0, Protocol: "frame", Frames: 7, Bytes: 1027},
		{Depth: 1, Protocol: "eth", Frames: 7, Bytes: 1027},
		{Depth: 4, Protocol: "http", Frames: 3, Bytes: 454},
	}
}

// The indentation is what says which protocol is carried by which. A flat list
// of names answers a different, less useful question.
func TestTheHierarchyKeepsItsShape(t *testing.T) {
	v := hierarchyLines(someHierarchyRows())

	require.Len(t, v.Rows, 3)
	assert.True(t, strings.HasPrefix(v.Rows[0].Text, "frame"))
	assert.True(t, strings.HasPrefix(v.Rows[1].Text, "  eth"))
	assert.True(t, strings.HasPrefix(v.Rows[2].Text, "        http"))
}

func TestEveryHierarchyRowFiltersOnItsProtocol(t *testing.T) {
	v := hierarchyLines(someHierarchyRows())

	assert.Equal(t, "frame", v.Rows[0].Filter)
	assert.Equal(t, "http", v.Rows[2].Filter)
}

// An indented name is wider than the name alone, so the counts have to be
// placed after the deepest one or the columns walk right along with the tree.
func TestTheCountsLineUpDespiteTheIndentation(t *testing.T) {
	v := hierarchyLines(someHierarchyRows())

	at := strings.Index(v.Rows[0].Text, "7")
	require.Positive(t, at)
	assert.Equal(t, at, strings.Index(v.Rows[1].Text, "7"))
}

//======================================================================

func someEndpointRows() []stats.EndpointRow {
	return []stats.EndpointRow{
		{Address: "10.0.0.1", Packets: 12, Bytes: 900, TxBytes: 400, RxBytes: 500},
		{Address: "10.0.0.2", Packets: 40, Bytes: 90000, TxBytes: 89000, RxBytes: 1000},
		{Address: "10.0.0.3", Packets: 3, Bytes: 300, TxBytes: 100, RxBytes: 200},
	}
}

// tshark prints endpoints in the order it met them, which buries the host you
// are looking for. The question the table answers is who is doing the most.
func TestEndpointsAreBusiestFirst(t *testing.T) {
	v := endpointLines(someEndpointRows())

	require.Len(t, v.Rows, 3)
	assert.True(t, strings.HasPrefix(v.Rows[0].Text, "10.0.0.2"))
	assert.True(t, strings.HasPrefix(v.Rows[2].Text, "10.0.0.3"))
}

func TestEveryEndpointRowFiltersOnItsAddress(t *testing.T) {
	v := endpointLines(someEndpointRows())

	assert.Equal(t, "ip.addr == 10.0.0.2", v.Rows[0].Filter)
}

// 90000 is not a number anyone reads at a glance, and this table exists to be
// glanced at.
func TestEndpointCountsAreGrouped(t *testing.T) {
	v := endpointLines(someEndpointRows())

	assert.Contains(t, v.Rows[0].Text, "90,000")
	assert.Contains(t, v.Rows[0].Text, "89,000")
}

func TestEndpointColumnsLineUp(t *testing.T) {
	v := endpointLines(someEndpointRows())

	at := strings.Index(v.Rows[0].Text, "90,000")
	require.Positive(t, at)
	for i, r := range v.Rows {
		assert.Equal(t, len([]rune(v.Rows[0].Text)), len([]rune(r.Text)),
			"row %d is a different width", i)
	}
}

func TestNoEndpointsIsAnEmptyView(t *testing.T) {
	assert.True(t, endpointLines(nil).empty())
}

//======================================================================

// Both statistics honour the display filter. A reader who has forgotten what
// is in the filter box would otherwise take a subset for the whole capture.
func TestTheHeadingNamesTheFilterInForce(t *testing.T) {
	assert.Equal(t, "Expert Information", statsHeading("Expert Information", ""))
	assert.Contains(t, statsHeading("Expert Information", "tcp.port == 23"), "tcp.port == 23")
}

func TestAnEmptyResultSaysWhichKindOfEmpty(t *testing.T) {
	assert.Contains(t, statsEmptyMessage(""), "capture")
	assert.Contains(t, statsEmptyMessage("udp"), "display filter")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
