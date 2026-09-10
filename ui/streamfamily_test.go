// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"testing"

	"github.com/gcla/gowid/gwutil"
	"github.com/m1rwana12/pcaptui/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// fakePacket stands in for a pdmltree.Model: the two questions the choice is
// made from, and nothing else.
type fakePacket struct {
	layers  map[string]bool
	indexes map[string]int
}

var _ streamPacket = fakePacket{}

func (f fakePacket) HasLayer(name string) bool { return f.layers[name] }

func (f fakePacket) FieldIndex(field string) gwutil.IntOption {
	if v, ok := f.indexes[field]; ok {
		return gwutil.SomeInt(v)
	}
	return gwutil.NoneInt()
}

func tcpPacket() fakePacket {
	return fakePacket{
		layers:  map[string]bool{"tcp": true},
		indexes: map[string]int{"tcp.stream": 4},
	}
}

func udpPacket() fakePacket {
	return fakePacket{
		layers:  map[string]bool{"udp": true},
		indexes: map[string]int{"udp.stream": 1},
	}
}

// A TLS packet is a TCP packet as well, and it has both indexes.
func tlsPacket() fakePacket {
	return fakePacket{
		layers:  map[string]bool{"tcp": true, "tls": true},
		indexes: map[string]int{"tcp.stream": 7, "tls.stream": 2},
	}
}

// A WebSocket packet has no index of its own: Wireshark has no
// websocket.stream field, and tshark indexes the stream by the TCP one.
func websocketPacket() fakePacket {
	return fakePacket{
		layers:  map[string]bool{"tcp": true, "http": true, "websocket": true},
		indexes: map[string]int{"tcp.stream": 3},
	}
}

func family(t *testing.T, token string) streams.Family {
	t.Helper()
	f, ok := streams.FamilyByToken(token)
	require.True(t, ok, token)
	return f
}

//======================================================================

func TestTheTransportStreamIsWhatTheKeyFollows(t *testing.T) {
	f, idx, err := pickStreamFamily(nil, tcpPacket())
	require.NoError(t, err)
	assert.Equal(t, streams.TCP, f.Proto)
	assert.Equal(t, 4, idx)

	f, idx, err = pickStreamFamily(nil, udpPacket())
	require.NoError(t, err)
	assert.Equal(t, streams.UDP, f.Proto)
	assert.Equal(t, 1, idx)
}

// The guarantee the whole design rests on. Following this packet as TLS would
// be an empty pane whenever there is no key log - measured: tshark answers a
// keyless TLS stream with a full banner, no payload and exit status 0 - so the
// key keeps giving the bytes, and TLS is asked for by name.
func TestATLSPacketIsStillFollowedAsTCPUnlessAsked(t *testing.T) {
	f, idx, err := pickStreamFamily(nil, tlsPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.TCP, f.Proto, "TLS was chosen for a user who did not ask for it")
	assert.Equal(t, 7, idx, "and with the TCP stream's index, not the TLS one")
}

func TestAskingForTLSFollowsTLS(t *testing.T) {
	want := family(t, "tls")

	f, idx, err := pickStreamFamily(&want, tlsPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.TLS, f.Proto)
	assert.Equal(t, 2, idx, "the tls.stream index, which is not the tcp.stream one")
	assert.Equal(t, "tls.stream eq 2", f.Filter(idx))
}

// A WebSocket stream is numbered by the TCP stream underneath it. Reading the
// index from the family's own name would find nothing here and the view would
// refuse to open on a packet that plainly has a stream.
func TestAWebSocketStreamIsNumberedByTheTCPStreamUnderIt(t *testing.T) {
	want := family(t, "websocket")

	f, idx, err := pickStreamFamily(&want, websocketPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.WebSocket, f.Proto)
	assert.Equal(t, 3, idx)
	assert.Equal(t, "tcp.stream eq 3", f.Filter(idx))
}

// The index alone cannot tell the families apart, because every TLS or
// WebSocket packet has a tcp.stream like any other. Asking for a family the
// packet does not carry has to be a refusal, not a stream of something else.
func TestAskingForAFamilyThePacketDoesNotHave(t *testing.T) {
	for _, token := range []string{"tls", "websocket"} {
		want := family(t, token)

		_, _, err := pickStreamFamily(&want, tcpPacket())

		require.Error(t, err, token)
		assert.Contains(t, err.Error(), want.Label, "the refusal should name what was asked for")
		assert.Contains(t, err.Error(), ":streams tcp",
			"and what can be followed here instead")
	}
}

// A layer tshark dissected but did not number. The field that would have
// carried the number is the only useful thing to say.
func TestALayerWithNoStreamNumberSaysWhichFieldIsMissing(t *testing.T) {
	want := family(t, "tls")
	pkt := fakePacket{
		layers:  map[string]bool{"tcp": true, "tls": true},
		indexes: map[string]int{"tcp.stream": 7},
	}

	_, _, err := pickStreamFamily(&want, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tls.stream")
}

func TestAPacketWithNoStreamAtAll(t *testing.T) {
	pkt := fakePacket{layers: map[string]bool{"icmp": true}}

	_, _, err := pickStreamFamily(nil, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TCP or UDP")
}

// The refusal lists what this packet does offer, so that a user who asked for
// the wrong family is told the right one rather than just told no.
func TestTheRefusalNamesEveryFamilyThePacketOffers(t *testing.T) {
	want := family(t, "tls")

	_, _, err := pickStreamFamily(&want, websocketPacket())

	require.Error(t, err)
	assert.Contains(t, err.Error(), ":streams tcp")
	assert.Contains(t, err.Error(), ":streams websocket")
	assert.NotContains(t, err.Error(), ":streams udp")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
