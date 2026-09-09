// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func cmdArgs(t *testing.T, c pcap.IPcapCommand) []string {
	t.Helper()

	cmd, ok := c.(*pcap.Command)
	require.True(t, ok, "expected a *pcap.Command")

	return cmd.Cmd.Args[1:]
}

func joinedArgs(t *testing.T, c pcap.IPcapCommand) string {
	t.Helper()
	return strings.Join(cmdArgs(t, c), " ")
}

// A statistic that does not carry the decode-as rules answers a different
// question than the packet list does about the same capture: with
// -d tcp.port==23,http the list showed HTTP and the protocol hierarchy went
// on counting telnet.
func TestTheStatsCommandCarriesTheSharedArguments(t *testing.T) {
	defer pcaptui.SetTsharkExtras(nil, nil)
	pcaptui.SetTsharkExtras([]string{"tcp.port==23,http"},
		[]string{"-o", "tls.keylog_file:/keys"})

	got := joinedArgs(t, MakeCommands().Stats("capture.pcap", "io,phs"))

	assert.Contains(t, got, "-d tcp.port==23,http")
	assert.Contains(t, got, "-o tls.keylog_file:/keys")
	assert.Contains(t, got, "-z io,phs")
	assert.Contains(t, got, "-r capture.pcap")
}

func TestWithNothingSharedTheCommandIsUnchanged(t *testing.T) {
	pcaptui.SetTsharkExtras(nil, nil)

	got := cmdArgs(t, MakeCommands().Stats("capture.pcap", "expert"))

	assert.Equal(t, []string{"-q", "-z", "expert", "-r", "capture.pcap"}, got)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
