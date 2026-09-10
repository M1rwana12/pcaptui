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

	tcp, ok := FamilyByToken("tcp")
	require.True(t, ok)

	follow := joinedArgs(t, MakeCommands().Stream("capture.pcap",
		Ref{Family: tcp, Index: 0}.FollowArg("raw")))
	assert.Contains(t, follow, "-o tls.keylog_file:/keys")
	assert.Contains(t, follow, "-d tcp.port==23,http")
	assert.Contains(t, follow, "-z follow,tcp,raw,0")

	// The index pass reads the same packets and has to agree with it.
	index := joinedArgs(t, MakeCommands().Indexer("capture.pcap", "tcp.stream eq 0"))
	assert.Contains(t, index, "-o tls.keylog_file:/keys")
	assert.Contains(t, index, "-d tcp.port==23,http")
	assert.Contains(t, index, "-Y tcp.stream eq 0")

	// A family layered over TCP asks tshark for that family, and indexes the
	// packets by the field that actually carries the number - which for
	// WebSocket is the TCP stream underneath, because there is no
	// websocket.stream field in Wireshark at all.
	ws, ok := FamilyByToken("websocket")
	require.True(t, ok)
	wsRef := Ref{Family: ws, Index: 2}
	wsFollow := joinedArgs(t, MakeCommands().Stream("capture.pcap", wsRef.FollowArg("raw")))
	assert.Contains(t, wsFollow, "-z follow,websocket,raw,2")
	wsIndex := joinedArgs(t, MakeCommands().Indexer("capture.pcap", wsRef.Filter()))
	assert.Contains(t, wsIndex, "-Y tcp.stream eq 2")

	tls, ok := FamilyByToken("tls")
	require.True(t, ok)
	tlsIndex := joinedArgs(t, MakeCommands().Indexer("capture.pcap",
		Ref{Family: tls, Index: 1}.Filter()))
	assert.Contains(t, tlsIndex, "-Y tls.stream eq 1")

	// A two-index family carries both numbers into both passes, and the index
	// pass filters on both fields.
	h2, ok := FamilyByToken("http2")
	require.True(t, ok)
	h2Ref := Ref{Family: h2, Index: 3, Sub: 5}
	h2Follow := joinedArgs(t, MakeCommands().Stream("capture.pcap", h2Ref.FollowArg("raw")))
	assert.Contains(t, h2Follow, "-z follow,http2,raw,3,5")
	h2Index := joinedArgs(t, MakeCommands().Indexer("capture.pcap", h2Ref.Filter()))
	assert.Contains(t, h2Index, "-Y tcp.stream eq 3 and http2.streamid eq 5")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
