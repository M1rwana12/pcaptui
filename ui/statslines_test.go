// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/capinfo"
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
	v := expertLines(someExpertRows(), nil)

	require.Len(t, v.Rows, 3)
	for i, r := range v.Rows {
		assert.True(t, r.actionable(), "row %d cannot be acted on", i)
	}
}

// tshark prints the severity once, as a heading over a section. Flattening the
// sections into one list is what makes the rows selectable, so each row has to
// carry its own severity or the flattening loses it.
func TestASeverityTravelsWithItsRow(t *testing.T) {
	v := expertLines(someExpertRows(), nil)

	assert.Contains(t, v.Rows[0].Text, "Error")
	assert.Contains(t, v.Rows[2].Text, "Note")
}

// "Errors" over a column of single rows reads as a tally. The column says what
// one row is.
func TestTheSeverityColumnIsSingular(t *testing.T) {
	v := expertLines(someExpertRows(), nil)

	assert.NotContains(t, v.Rows[0].Text, "Errors")
	assert.NotContains(t, v.Rows[2].Text, "Notes")
}

func TestTheColumnTitlesAreAHeaderNotARow(t *testing.T) {
	v := expertLines(someExpertRows(), nil)

	require.Len(t, v.Header, 2)
	assert.Contains(t, v.Header[0], "Severity")
	assert.Contains(t, v.Header[0], "Summary")

	for _, r := range v.Rows {
		assert.NotContains(t, r.Text, "Severity")
	}
}

func TestColumnsLineUp(t *testing.T) {
	v := expertLines(someExpertRows(), nil)

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
	v := expertLines(someExpertRows(), nil)

	widest := len([]rune(v.Header[0]))
	for _, r := range v.Rows {
		if n := len([]rune(r.Text)); n > widest {
			widest = n
		}
	}

	assert.Equal(t, widest, len([]rune(v.Header[1])))
}

func TestNoRowsIsAnEmptyView(t *testing.T) {
	assert.True(t, expertLines(nil, nil).empty())
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

// One tshark run with three -z arguments produces all three tables in one
// stream, in this order. Verbatim from
// `tshark -q -z expert -z io,phs -z endpoints,ip -r telnet-cooked.pcap`.
const overviewOutput = `================================================================================
IPv4 Endpoints
Filter:<No Filter>
                       | Packets | |  Bytes  | | Tx Packets | | Tx Bytes | | Rx Packets | | Rx Bytes |
192.168.0.2                    92   7748 bytes         48      3465 bytes          44      4283 bytes
192.168.0.1                    92   7748 bytes         44      4283 bytes          48      3465 bytes
================================================================================

===================================================================
Protocol Hierarchy Statistics
Filter:

frame                                    frames:92 bytes:7748
  eth                                    frames:92 bytes:7748
    ip                                   frames:92 bytes:7748
      tcp                                frames:92 bytes:7748
        telnet                           frames:46 bytes:4670
          _ws.malformed                  frames:1 bytes:67
===================================================================

Errors (6)
=============
   Frequency      Group           Protocol  Summary
           5   Protocol               IPv4  IPv4 total length exceeds packet length (52 bytes)
           1  Malformed             TELNET  Malformed Packet (Exception occurred)
`

func TestTheOverviewAnswersAllThreeQuestions(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})

	text := strings.Join(rowTexts(v), "\n")
	assert.Contains(t, text, "What is wrong")
	assert.Contains(t, text, "What is in it")
	assert.Contains(t, text, "Who is on the wire")
}

// Problems first: that is what somebody handed a capture is looking for.
func TestTheOverviewLeadsWithProblems(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})

	assert.Equal(t, "What is wrong", v.Rows[0].Text)
}

// Three tables arrive in one stream, so each parser has to take its own rows
// and leave the others alone. An expert line has seven fields once the word
// "bytes" is dropped often enough to be a real risk of inventing an endpoint.
func TestEachSectionTakesOnlyItsOwnRows(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})
	text := strings.Join(rowTexts(v), "\n")

	assert.Contains(t, text, "IPv4 total length exceeds")
	assert.Contains(t, text, "192.168.0.2")
	assert.Contains(t, text, "frame")

	// The hierarchy's protocol names must not appear as endpoint addresses,
	// nor expert frequencies as packet counts.
	for _, r := range v.Rows {
		if strings.Contains(r.Text, "192.168.0.") {
			assert.Equal(t, "ip.addr == 192.168.0.2", r.Filter,
				"an endpoint row should filter on its address")
			break
		}
	}
}

// A summary that cannot be acted on is a dead end; every row keeps the filter
// its own table would have given it.
func TestOverviewRowsStillCarryTheirFilters(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})

	var actionable int
	for _, r := range v.Rows {
		if r.actionable() {
			actionable++
		}
	}

	assert.GreaterOrEqual(t, actionable, 5,
		"the overview should be mostly rows that lead somewhere")
}

func TestASectionSaysHowManyItLeftOut(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})

	assert.Contains(t, strings.Join(rowTexts(v), "\n"), "and 2 more",
		"six hierarchy rows shown four at a time should say so")
}

// A capture with nothing wrong is itself worth knowing, and in one line rather
// than as a missing section.
func TestAnEmptySectionSaysSoInWords(t *testing.T) {
	onlyEndpoints := `192.168.0.2  92  7748 bytes  48  3465 bytes  44  4283 bytes
`
	v := overviewLines(onlyEndpoints, capinfo.Info{})
	text := strings.Join(rowTexts(v), "\n")

	assert.Contains(t, text, "Nothing the dissectors object to")
	assert.Contains(t, text, "No protocols reported")
	assert.Contains(t, text, "192.168.0.2")
}

func rowTexts(v statsView) []string {
	res := make([]string, 0, len(v.Rows))
	for _, r := range v.Rows {
		res = append(res, r.Text)
	}
	return res
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

func someHTTPRows() []stats.TreeRow {
	return []stats.TreeRow{
		{Depth: 0, Name: "Total HTTP Packets", Count: 3, Percent: "100%"},
		{Depth: 1, Name: "HTTP Response Packets", Count: 3, Percent: "100,00%"},
		{Depth: 2, Name: "5xx: Server Error", Count: 1, Percent: "33,33%"},
		{Depth: 3, Name: "500 Internal Server Error", Count: 1, Percent: "100,00%"},
		{Depth: 2, Name: "1xx: Informational", Count: 0, Percent: "0,00%"},
		{Depth: 1, Name: "HTTP Request Packets", Count: 0, Percent: "0,00%"},
	}
}

// tshark's table is a skeleton of every class it knows; on a real capture most
// of it is zeroes, and the zeroes are what the eye lands on first.
func TestTheEmptyHTTPRowsAreNotShown(t *testing.T) {
	v := httpLines(someHTTPRows())

	require.Len(t, v.Rows, 4)
	for _, r := range v.Rows {
		assert.NotContains(t, r.Text, "1xx")
		assert.NotContains(t, r.Text, "Request")
	}
}

func TestTheHTTPTableKeepsItsShape(t *testing.T) {
	v := httpLines(someHTTPRows())

	assert.True(t, strings.HasPrefix(v.Rows[0].Text, "Total"))
	assert.True(t, strings.HasPrefix(v.Rows[1].Text, " HTTP Response"))
	assert.True(t, strings.HasPrefix(v.Rows[2].Text, "  5xx"))
	assert.True(t, strings.HasPrefix(v.Rows[3].Text, "   500"))
}

func TestAStatusRowFiltersOnItsCode(t *testing.T) {
	v := httpLines(someHTTPRows())

	assert.Equal(t, "http.response.code == 500", v.Rows[3].Filter)
	assert.True(t, v.Rows[3].actionable())
}

// The percentage is printed as tshark gave it, decimal comma and all, because
// re-formatting it means first guessing which convention produced it.
func TestThePercentageIsShownAsTsharkPrintedIt(t *testing.T) {
	v := httpLines(someHTTPRows())

	assert.True(t, strings.HasSuffix(strings.TrimRight(v.Rows[2].Text, " "), "33,33%"))
}

func TestTheHTTPCountsLineUpDespiteTheIndentation(t *testing.T) {
	v := httpLines(someHTTPRows())

	at := strings.Index(v.Rows[0].Text, "3")
	require.Positive(t, at)
	assert.Equal(t, at, strings.Index(v.Rows[1].Text, "3"))
	assert.Equal(t, at, strings.Index(v.Rows[2].Text, "1"))
	assert.Equal(t, at, strings.Index(v.Rows[3].Text, "1"))
}

// A capture with no HTTP in it still prints the whole skeleton, at zero. That
// is not a table worth opening.
func TestACaptureWithNoHTTPIsAnEmptyView(t *testing.T) {
	rows := []stats.TreeRow{
		{Depth: 0, Name: "Total HTTP Packets", Count: 0, Percent: "100%"},
		{Depth: 1, Name: "HTTP Response Packets", Count: 0, Percent: "0,00%"},
	}

	assert.True(t, httpLines(rows).empty())
	assert.True(t, httpLines(nil).empty())
}

//======================================================================

func someDNSRows() []stats.TreeRow {
	return []stats.TreeRow{
		{Depth: 0, Name: "Total Packets", Count: 4, Percent: "100%"},
		{Depth: 0, Name: "rcode", Count: 4, Percent: "100%"},
		{Depth: 1, Name: "No error", Parent: "rcode", Count: 3, Percent: "75,00%"},
		{Depth: 1, Name: "No such name", Parent: "rcode", Count: 1, Percent: "25,00%"},
		{Depth: 0, Name: "Query Stats", Count: 0, Percent: "100%"},
		{Depth: 1, Name: "Qname Len", Parent: "Query Stats", Count: 2, Average: "12,00"},
		{Depth: 0, Name: "Service Stats", Count: 0, Percent: "100%"},
		{Depth: 1, Name: "no. of retransmissions", Parent: "Service Stats", Count: 0},
	}
}

// A heading counted at zero stays when its own rows were counted; one with
// nothing under it goes.
func TestAnEmptyDNSSectionGoesAndAFullOneStays(t *testing.T) {
	v := dnsLines(someDNSRows())

	var text string
	for _, r := range v.Rows {
		text += r.Text + "\n"
	}

	assert.Contains(t, text, "Query Stats")
	assert.Contains(t, text, "Qname Len")
	assert.NotContains(t, text, "Service Stats")
	assert.NotContains(t, text, "retransmissions")
}

// The row that says how long the server took has its count in packets and its
// answer in milliseconds, so the count alone says nothing.
func TestTheDNSAverageHasAColumn(t *testing.T) {
	v := dnsLines(someDNSRows())

	assert.Contains(t, v.Header[0], "Average")

	at := strings.Index(v.Header[0], "Average")
	require.Positive(t, at)
	for _, r := range v.Rows {
		if strings.Contains(r.Text, "Qname Len") {
			assert.Contains(t, r.Text[at:], "12,00")
			return
		}
	}
	t.Fatal("the row with the average was not shown")
}

// A table where nothing has an average should not carry an empty column for
// one. The HTTP table is the case: every row is a count.
func TestATableWithNoAveragesHasNoAverageColumn(t *testing.T) {
	rows := []stats.TreeRow{
		{Depth: 0, Name: "Total Packets", Count: 4, Percent: "100%"},
		{Depth: 0, Name: "rcode", Count: 4, Percent: "100%"},
	}

	v := dnsLines(rows)

	assert.NotContains(t, v.Header[0], "Average")
}

func TestADNSRowFiltersThroughItsSection(t *testing.T) {
	v := dnsLines(someDNSRows())

	byText := map[string]statsLine{}
	for _, r := range v.Rows {
		byText[strings.TrimSpace(strings.Split(r.Text, "  ")[0])] = r
	}

	assert.Equal(t, "dns.flags.rcode == 3", byText["No such name"].Filter)
	assert.Equal(t, "dns", byText["Total Packets"].Filter)
	assert.Equal(t, "", byText["Qname Len"].Filter,
		"no filter selects the packets behind an average")
	assert.False(t, byText["Qname Len"].actionable())
}

func TestACaptureWithNoDNSIsAnEmptyView(t *testing.T) {
	rows := []stats.TreeRow{
		{Depth: 0, Name: "Total Packets", Count: 0, Percent: "100%"},
		{Depth: 0, Name: "rcode", Count: 0, Percent: "100%"},
	}

	assert.True(t, dnsLines(rows).empty())
	assert.True(t, dnsLines(nil).empty())
}

//======================================================================

func someNumberedRows() []stats.ExpertRow {
	rows := []stats.ExpertRow{
		{Severity: "Errors", Count: 20480, Group: "Protocol", Protocol: "IPv4",
			Summary: "IPv4 total length exceeds packet length (52 bytes)"},
	}
	for i := 1; i <= 4095; i++ {
		rows = append(rows, stats.ExpertRow{
			Severity: "Notes", Count: 35, Group: "Sequence", Protocol: "TCP",
			Summary: fmt.Sprintf("Duplicate ACK (#%d)", i),
		})
	}
	return rows
}

// A capture of 376,832 packets opened on 4,108 rows, 4,095 of which were one
// fact repeated with a counter in the text.
func TestTheNumberedRowsArriveAsOne(t *testing.T) {
	v := expertLines(someNumberedRows(), nil)

	require.Len(t, v.Rows, 2)
	assert.Contains(t, v.Rows[1].Text, "Duplicate ACK (#1-#4095)",
		"a folded count with no sign it was folded reads as one problem seen that often")
	assert.Contains(t, v.Rows[1].Text, "143,325")
}

func TestAFoldedRowIsStillSelectable(t *testing.T) {
	v := expertLines(someNumberedRows(), nil)

	assert.Equal(t,
		`_ws.expert.message matches "^Duplicate ACK [(]#[0-9]+[)]$"`,
		v.Rows[1].Filter)
	assert.True(t, v.Rows[1].actionable())
}

// tshark counts these in its section headings, and the program used to read
// them and drop them.
func TestTheSeverityTotalsAreShown(t *testing.T) {
	v := expertLines(someExpertRows(), []stats.SeverityTotal{
		{Severity: "Errors", Count: 20481},
		{Severity: "Warns", Count: 3},
		{Severity: "Notes", Count: 430012},
	})

	require.NotEmpty(t, v.Header)
	assert.Equal(t, "20,481 errors · 3 warnings · 430,012 notes", v.Header[0])
}

// A capture with one of something said "1 notes".
func TestOneOfSomethingIsSingular(t *testing.T) {
	v := expertLines(someExpertRows(), []stats.SeverityTotal{
		{Severity: "Notes", Count: 1},
		{Severity: "Warns", Count: 1},
		{Severity: "Chats", Count: 3},
	})

	require.NotEmpty(t, v.Header)
	assert.Equal(t, "1 note · 1 warning · 3 chats", v.Header[0])
}

func TestWithNoTotalsThereIsNoTotalsLine(t *testing.T) {
	v := expertLines(someExpertRows(), nil)

	require.NotEmpty(t, v.Header)
	assert.Contains(t, v.Header[0], "Severity", "the column titles come first")
}

// The same quantity was printed two ways in one dialog: the hierarchy with
// %d and everything else grouped.
func TestEveryTableGroupsItsDigits(t *testing.T) {
	hier := hierarchyLines([]stats.HierarchyRow{
		{Depth: 0, Protocol: "frame", Frames: 376832, Bytes: 44483480},
	})

	require.Len(t, hier.Rows, 1)
	assert.Contains(t, hier.Rows[0].Text, "376,832")
	assert.Contains(t, hier.Rows[0].Text, "44,483,480")
}

//======================================================================

func someFileFacts() capinfo.Info {
	return capinfo.Info{
		Packets:  "92",
		Size:     "9244 bytes",
		Duration: "39,571274 seconds",
		Earliest: "1999-11-28 04:12:38,387203",
		SnapLen:  "file hdr: 1514 bytes",
	}
}

// The three statistics answer what is wrong, what is in it and who is on the
// wire. None of them answers when - and "this is a forty-second slice from
// 1999, not the hour you asked for" is often the whole answer.
func TestTheOverviewSaysWhatTheFileIs(t *testing.T) {
	v := overviewLines(overviewOutput, someFileFacts())

	var text string
	for _, r := range v.Rows {
		text += r.Text + "\n"
	}

	assert.Contains(t, text, "What this file is")
	assert.Contains(t, text, "1999-11-28 04:12:38,387203")
	assert.Contains(t, text, "39,571274 seconds")
	assert.Contains(t, text, "file hdr: 1514 bytes")
}

// It comes first because it is the cheapest and the most general: capinfos
// took 0.27 s on a capture whose -z pass took 9 to 11 s.
func TestTheFileComesBeforeTheStatistics(t *testing.T) {
	v := overviewLines(overviewOutput, someFileFacts())

	require.NotEmpty(t, v.Rows)
	assert.Equal(t, "What this file is", v.Rows[0].Text)
}

// capinfos missing or failing is not a reason to say nothing about the rest.
func TestWithNoFileFactsTheSectionIsNotDrawn(t *testing.T) {
	v := overviewLines(overviewOutput, capinfo.Info{})

	require.NotEmpty(t, v.Rows)
	assert.Equal(t, "What is wrong", v.Rows[0].Text)
	for _, r := range v.Rows {
		assert.NotContains(t, r.Text, "What this file is")
	}
}

// capinfos writes "file hdr: (not set)" when there is no snapshot length, and
// a line saying a limit is not set is a line spent on nothing.
func TestASnapshotLimitThatIsNotSetIsNotShown(t *testing.T) {
	info := someFileFacts()
	info.SnapLen = "file hdr: (not set)"

	lines := fileFactsLines(info, nil)

	require.NotEmpty(t, lines)
	for _, l := range lines {
		assert.NotContains(t, l.Text, "Packet limit")
	}
}

// A fact capinfos did not report is left out rather than shown blank: an empty
// value reads as something looked up and found missing.
func TestAFactThatIsNotThereIsNotAnEmptyLine(t *testing.T) {
	lines := fileFactsLines(capinfo.Info{Packets: "92"}, nil)

	require.Len(t, lines, 1)
	assert.Contains(t, lines[0].Text, "Packets")
	assert.Contains(t, lines[0].Text, "92")
}

// These say what the file is, not which packets to look at, so none of them
// pretends to be selectable.
func TestTheFileFactsAreNotSelectable(t *testing.T) {
	for _, l := range fileFactsLines(someFileFacts(), nil) {
		assert.False(t, l.actionable())
	}
}

//======================================================================

func trafficRows(frames ...int) []stats.IOStatRow {
	rows := make([]stats.IOStatRow, 0, len(frames))
	for i, f := range frames {
		rows = append(rows, stats.IOStatRow{
			Label:  fmt.Sprintf("%d <> %d", i*5, (i+1)*5),
			Frames: f,
		})
	}
	return rows
}

// The shape is the answer: a burst at the start, a hole in the middle. Reading
// it off twenty numbers is work; reading it off one row is not.
func TestTheTrafficIsDrawnAsAShape(t *testing.T) {
	got := sparkline(trafficRows(44, 6, 0, 2, 14, 14, 0, 12))

	assert.Equal(t, 8, len([]rune(got)))
	assert.Equal(t, '█', []rune(got)[0], "the busiest interval is full height")
}

// A gap is the point, so an empty interval is not drawn as a block at all.
func TestAnEmptyIntervalIsAGap(t *testing.T) {
	got := []rune(sparkline(trafficRows(10, 0, 10)))

	require.Len(t, got, 3)
	assert.Equal(t, '·', got[1])
}

// Rounding a lone packet down to nothing would turn "quiet" into "silent",
// and silence is what the gaps mean.
func TestOnePacketIsNotSilence(t *testing.T) {
	got := []rune(sparkline(trafficRows(1000, 1)))

	require.Len(t, got, 2)
	assert.NotEqual(t, '·', got[1])
	assert.Equal(t, '▁', got[1], "the shortest block, but a block")
}

func TestAFlatCaptureIsFlat(t *testing.T) {
	got := sparkline(trafficRows(5, 5, 5))

	assert.Equal(t, "███", got)
}

// Nothing to draw is not a row of nothing: with no traffic at all, or no
// buckets, the line is left out.
func TestNoTrafficDrawsNoLine(t *testing.T) {
	assert.Equal(t, "", sparkline(nil))
	assert.Equal(t, "", sparkline(trafficRows(0, 0, 0)))
}

func TestTheTrafficLineJoinsTheFileFacts(t *testing.T) {
	lines := fileFactsLines(someFileFacts(), trafficRows(4, 0, 2))

	var text string
	for _, l := range lines {
		text += l.Text + "\n"
	}

	assert.Contains(t, text, "Traffic")
	assert.Contains(t, text, "█·▄")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
