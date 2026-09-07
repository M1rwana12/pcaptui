// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/summary"
	"github.com/stretchr/testify/assert"
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

func runStat(t *testing.T, s Stat, filter string) string {
	t.Helper()

	cmd := MakeCommands().Stats(testPcap, s.ZArg(filter))

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
// Local Variables:
// mode: Go
// fill-column: 78
// End:
