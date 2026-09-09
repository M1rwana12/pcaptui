// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package streams

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

// The User Guide says that with the session keys "stream reassembly shows the
// plaintext". Without the key log on this command line it shows ciphertext.
func TestTheStreamCommandsCarryTheSharedArguments(t *testing.T) {
	defer pcaptui.SetTsharkExtras(nil, nil)
	pcaptui.SetTsharkExtras([]string{"tcp.port==23,http"},
		[]string{"-o", "tls.keylog_file:/keys"})

	follow := joinedArgs(t, MakeCommands().Stream("capture.pcap", "tcp", 0))
	assert.Contains(t, follow, "-o tls.keylog_file:/keys")
	assert.Contains(t, follow, "-d tcp.port==23,http")
	assert.Contains(t, follow, "-z follow,tcp,raw,0")

	// The index pass reads the same packets and has to agree with it.
	index := joinedArgs(t, MakeCommands().Indexer("capture.pcap", "tcp", 0))
	assert.Contains(t, index, "-o tls.keylog_file:/keys")
	assert.Contains(t, index, "-d tcp.port==23,http")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
