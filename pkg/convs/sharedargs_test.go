// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package convs

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func joinedArgs(t *testing.T, c pcap.IPcapCommand) string {
	t.Helper()

	cmd, ok := c.(*pcap.Command)
	require.True(t, ok, "expected a *pcap.Command")

	return strings.Join(cmd.Cmd.Args[1:], " ")
}

// The conversations have to be the ones the packet list is showing, which
// means the same decode-as rules and the same key log.
func TestTheConvsCommandCarriesTheSharedArguments(t *testing.T) {
	defer pcaptui.SetTsharkExtras(nil, nil)
	pcaptui.SetTsharkExtras([]string{"tcp.port==23,http"},
		[]string{"-o", "tls.keylog_file:/keys"})

	got := joinedArgs(t, MakeCommands().Convs("capture.pcap", []string{"tcp"}, "", false, true))

	assert.Contains(t, got, "-d tcp.port==23,http")
	assert.Contains(t, got, "-o tls.keylog_file:/keys")
	assert.Contains(t, got, "-z conv,tcp")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
