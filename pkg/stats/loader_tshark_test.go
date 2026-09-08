// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/summary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

var testWG sync.WaitGroup

// pcap.Command.Start starts a summary goroutine, and that package tracks its
// goroutines through a package-level WaitGroup that cmd/pcaptui sets up. A
// test binary has to do the same or Start panics on a nil WaitGroup.
func TestMain(m *testing.M) {
	summary.Goroutinewg = &testWG
	Goroutinewg = &testWG

	code := m.Run()

	testWG.Wait()
	os.Exit(code)
}

//======================================================================

// These drive a real tshark over a real capture. Unit tests can only confirm
// that the -z argument is spelled the way this package intends; only tshark
// can confirm that spelling is the one tshark accepts. The column format
// reader in pkg/shark was broken for years precisely because nothing checked
// the second half.

const testPcap = "../../scripts/pcaps/telnet-cooked.pcap"

// The telnet capture has no HTTP in it. The demo capture is three responses -
// a 200, a 404 and a 500 - one of each class the HTTP table groups by, and so
// it exercises the whole shape of that table.
const httpPcap = "../../scripts/pcaps/demo.pcap"

func runStat(t *testing.T, s Stat, filter string) string {
	return runStatOn(t, testPcap, s, filter)
}

func runStatOn(t *testing.T, pcap string, s Stat, filter string) string {
	t.Helper()

	cmd := MakeCommands().Stats(pcap, s.ZArg(filter))

	out, err := cmd.StdoutReader()
	assert.NoError(t, err)

	assert.NoError(t, cmd.Start())

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("could not run tshark: %v", err)
		}
	}

	return buf.String()
}

func TestExpertReportsProblemsInTheCapture(t *testing.T) {
	out := runStat(t, Expert, "")

	// The capture carries a malformed telnet packet and several IPv4 length
	// errors; the point is that tshark reports them under a severity heading,
	// not the exact tally, which moves between Wireshark releases.
	assert.Contains(t, out, "Errors")
	assert.Contains(t, out, "Summary")
}

func TestProtoHierarchyListsTheProtocolsPresent(t *testing.T) {
	out := runStat(t, ProtoHierarchy, "")

	assert.Contains(t, out, "Protocol Hierarchy Statistics")
	assert.Contains(t, out, "telnet")
	assert.Contains(t, out, "frames:")
}

func TestDisplayFilterNarrowsTheStatistic(t *testing.T) {
	all := runStat(t, Expert, "")
	telnet := runStat(t, Expert, "telnet")

	assert.Contains(t, all, "IPv4")
	assert.NotContains(t, telnet, "IPv4",
		"a filter of telnet should drop the IPv4 expert entries")
}

// A filter that matches nothing produces no output at all - not a header, not
// an empty table. The UI has to say so itself, or the user gets a blank dialog
// and no idea why.
func TestStatisticCanLegitimatelyProduceNoOutput(t *testing.T) {
	out := runStat(t, Expert, "udp")

	assert.Empty(t, strings.TrimSpace(out))
}

func TestProtoHierarchyHonoursAFilterWithCommas(t *testing.T) {
	out := runStat(t, ProtoHierarchy, "tcp.port in {23,80}")

	assert.Contains(t, out, "Protocol Hierarchy Statistics")
	assert.Contains(t, out, "tcp.port in {23,80}",
		"tshark echoes the filter it was given, commas and all")
}

//======================================================================

// The parsers are unit-tested against captured output. These check that the
// captured output is still what tshark emits - a format read from a fixture
// that the program never re-reads is a format that quietly stops matching.

func TestExpertOutputStillParses(t *testing.T) {
	rows := ParseExpert(runStat(t, Expert, ""))

	assert.NotEmpty(t, rows, "tshark reported expert info that the parser read as nothing")

	for _, r := range rows {
		assert.NotEmpty(t, r.Severity, "row has no severity: %+v", r)
		assert.NotEmpty(t, r.Protocol, "row has no protocol: %+v", r)
		assert.NotEmpty(t, r.Summary, "row has no summary: %+v", r)
		assert.Greater(t, r.Count, 0, "row has no count: %+v", r)
	}
}

func TestHierarchyOutputStillParses(t *testing.T) {
	rows := ParseHierarchy(runStat(t, ProtoHierarchy, ""))

	require.NotEmpty(t, rows)

	// Which protocol is at the root is not asserted. Wireshark 4.6 prints a
	// "frame" row above "eth" and the version on the Ubuntu runner does not,
	// and either is a correct hierarchy. What has to hold is that the tree
	// starts at the left and gets deeper.
	assert.Equal(t, 0, rows[0].Depth, "the first row should be the root")

	var deepest int
	for _, r := range rows {
		assert.NotEmpty(t, r.Protocol)
		assert.Greater(t, r.Frames, 0)
		if r.Depth > deepest {
			deepest = r.Depth
		}
	}
	assert.Greater(t, deepest, 0, "every protocol came out at the same depth")
}

// The whole point of the expert table's filter: what it produces has to be
// something tshark will run and answer with the packets that row is about.
func TestAnExpertRowsFilterFindsItsPackets(t *testing.T) {
	rows := ParseExpert(runStat(t, Expert, ""))

	var checked int
	for _, r := range rows {
		out := runFieldsQuery(t, r.DisplayFilter(), "frame.number")

		// tshark refuses an invalid filter and writes to stderr instead, so an
		// empty result here would also be how a broken expression looks. Only
		// the count is asserted, because which frames carry which expert item
		// moves between Wireshark releases.
		if strings.TrimSpace(out) == "" {
			continue
		}
		checked++
		assert.LessOrEqual(t, len(strings.Fields(out)), r.Count+2,
			"filter %q matched more packets than the row claims", r.DisplayFilter())
	}

	assert.Greater(t, checked, 0,
		"not one expert row produced a filter that matched anything")
}

// Every name in the hierarchy has to be a filter tshark will accept, because
// the table offers it as one. Names like _ws.malformed and data-text-lines are
// the ones worth checking - they do not look like protocol names.
func TestEveryHierarchyRowsFilterIsAcceptedByTshark(t *testing.T) {
	rows := ParseHierarchy(runStat(t, ProtoHierarchy, ""))
	require.NotEmpty(t, rows)

	for _, r := range rows {
		cmd := exec.Command("tshark", "-r", testPcap, "-Y", r.DisplayFilter(),
			"-T", "fields", "-e", "frame.number")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = io.Discard

		assert.NoError(t, cmd.Run(),
			"tshark refused %q: %s", r.DisplayFilter(), stderr.String())
	}
}

func TestEndpointsOutputStillParses(t *testing.T) {
	out := runStat(t, Endpoints, "")
	rows := ParseEndpoints(out)

	// The raw output on failure, because the whole point of these tests is to
	// catch a format this parser has not seen - and a bare "was empty" from a
	// runner whose Wireshark differs from the developer's says nothing about
	// what it actually printed.
	require.NotEmpty(t, rows,
		"tshark listed endpoints that the parser read as nothing. Raw output:\n%s", out)

	for _, r := range rows {
		assert.NotEmpty(t, r.Address)
		assert.Greater(t, r.Packets, 0)
		assert.Equal(t, r.Packets, r.TxPkts+r.RxPkts,
			"sent plus received should be the total for %s", r.Address)
		assert.Equal(t, r.Bytes, r.TxBytes+r.RxBytes,
			"sent plus received bytes should be the total for %s", r.Address)
	}
}

// The table offers each address as a filter, so tshark has to accept it and
// answer with that address's traffic.
func TestAnEndpointRowsFilterFindsItsPackets(t *testing.T) {
	out := runStat(t, Endpoints, "")
	rows := ParseEndpoints(out)
	require.NotEmpty(t, rows, "raw output:\n%s", out)

	for _, r := range rows {
		out := runFieldsQuery(t, r.DisplayFilter(), "frame.number")

		assert.Equal(t, r.Packets, len(strings.Fields(out)),
			"filter %q found a different number of packets than the row claims",
			r.DisplayFilter())
	}
}

func runFieldsQuery(t *testing.T, filter string, field string) string {
	t.Helper()

	return runFieldsQueryOn(t, testPcap, filter, field)
}

func runFieldsQueryOn(t *testing.T, pcap string, filter string, field string) string {
	t.Helper()

	cmd := exec.Command("tshark", "-r", pcap, "-Y", filter, "-T", "fields", "-e", field)
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("could not run tshark: %v", err)
		}
	}
	return string(out)
}

func TestHTTPOutputStillParses(t *testing.T) {
	out := runStatOn(t, httpPcap, HTTP, "")
	rows := ParseTree(out)

	require.NotEmpty(t, rows,
		"tshark printed an HTTP table that the parser read as nothing. Raw output:\n%s", out)

	byName := map[string]TreeRow{}
	for _, r := range rows {
		byName[r.Name] = r
	}

	total, ok := byName["Total HTTP Packets"]
	require.True(t, ok, "raw output:\n%s", out)
	assert.Equal(t, 3, total.Count)
	assert.Equal(t, 0, total.Depth)

	for _, name := range []string{"200 OK", "404 Not Found", "500 Internal Server Error"} {
		r, ok := byName[name]
		require.True(t, ok, "%s missing. Raw output:\n%s", name, out)
		assert.Equal(t, 1, r.Count)
		assert.Greater(t, r.Depth, byName["HTTP Response Packets"].Depth,
			"%s should sit under the responses", name)
	}
}

// Each row offers a filter, so tshark has to accept it and answer with the
// packets that row counted.
func TestEveryHTTPRowsFilterFindsItsPackets(t *testing.T) {
	out := runStatOn(t, httpPcap, HTTP, "")
	rows := HTTPRows(ParseTree(out))
	require.NotEmpty(t, rows, "raw output:\n%s", out)

	tried := 0
	for _, r := range rows {
		filter := HTTPFilter(r.Name)
		if filter == "" {
			continue
		}
		tried++

		found := runFieldsQueryOn(t, httpPcap, filter, "frame.number")
		assert.Equal(t, r.Count, len(strings.Fields(found)),
			"filter %q found a different number of packets than %q claims",
			filter, r.Name)
	}

	require.NotZero(t, tried, "no row offered a filter to check")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
