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
	ref, err := pickStreamFamily(nil, tcpPacket())
	require.NoError(t, err)
	assert.Equal(t, streams.TCP, ref.Family.Proto)
	assert.Equal(t, 4, ref.Index)

	ref, err = pickStreamFamily(nil, udpPacket())
	require.NoError(t, err)
	assert.Equal(t, streams.UDP, ref.Family.Proto)
	assert.Equal(t, 1, ref.Index)
}

// The guarantee the whole design rests on. Following this packet as TLS would
// be an empty pane whenever there is no key log - measured: tshark answers a
// keyless TLS stream with a full banner, no payload and exit status 0 - so the
// key keeps giving the bytes, and TLS is asked for by name.
func TestATLSPacketIsStillFollowedAsTCPUnlessAsked(t *testing.T) {
	ref, err := pickStreamFamily(nil, tlsPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.TCP, ref.Family.Proto, "TLS was chosen for a user who did not ask for it")
	assert.Equal(t, 7, ref.Index, "and with the TCP stream's index, not the TLS one")
}

func TestAskingForTLSFollowsTLS(t *testing.T) {
	want := family(t, "tls")

	ref, err := pickStreamFamily(&want, tlsPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.TLS, ref.Family.Proto)
	assert.Equal(t, 2, ref.Index, "the tls.stream index, which is not the tcp.stream one")
	assert.Equal(t, "tls.stream eq 2", ref.Filter())
}

// A WebSocket stream is numbered by the TCP stream underneath it. Reading the
// index from the family's own name would find nothing here and the view would
// refuse to open on a packet that plainly has a stream.
func TestAWebSocketStreamIsNumberedByTheTCPStreamUnderIt(t *testing.T) {
	want := family(t, "websocket")

	ref, err := pickStreamFamily(&want, websocketPacket())

	require.NoError(t, err)
	assert.Equal(t, streams.WebSocket, ref.Family.Proto)
	assert.Equal(t, 3, ref.Index)
	assert.Equal(t, "tcp.stream eq 3", ref.Filter())
}

// The index alone cannot tell the families apart, because every TLS or
// WebSocket packet has a tcp.stream like any other. Asking for a family the
// packet does not carry has to be a refusal, not a stream of something else.
func TestAskingForAFamilyThePacketDoesNotHave(t *testing.T) {
	for _, token := range []string{"tls", "websocket"} {
		want := family(t, token)

		_, err := pickStreamFamily(&want, tcpPacket())

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

	_, err := pickStreamFamily(&want, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tls.stream")
}

func TestAPacketWithNoStreamAtAll(t *testing.T) {
	pkt := fakePacket{layers: map[string]bool{"icmp": true}}

	_, err := pickStreamFamily(nil, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TCP or UDP")
}

// On a Wireshark with no tls.stream - anything before 4.4 - the TLS family is
// indexed by the TCP stream underneath, which is what that tshark reports as
// the filter for its own follow,tls tap.
func TestOnAnOlderWiresharkTLSIsIndexedByTheTCPStream(t *testing.T) {
	want := family(t, "tls")
	noTLSStream := func(name string) bool { return name != "tls.stream" }
	pkt := fakePacket{
		layers:  map[string]bool{"tcp": true, "tls": true},
		indexes: map[string]int{"tcp.stream": 7},
	}

	ref, err := pickStreamFamilyWith(&want, pkt, noTLSStream)

	require.NoError(t, err, "following TLS should still work where tls.stream does not exist")
	assert.Equal(t, streams.TLS, ref.Family.Proto)
	assert.Equal(t, 7, ref.Index)
	assert.Equal(t, "tcp.stream eq 7", ref.Filter())
}

// An HTTP/2 stream is two numbers, and both come off the packet: the TCP
// stream it travelled in and the stream id within it.
func TestAnHTTP2StreamIsTwoNumbersFromTheSamePacket(t *testing.T) {
	want := family(t, "http2")
	pkt := fakePacket{
		layers:  map[string]bool{"tcp": true, "http2": true},
		indexes: map[string]int{"tcp.stream": 0, "http2.streamid": 1},
	}

	ref, err := pickStreamFamily(&want, pkt)

	require.NoError(t, err)
	assert.Equal(t, streams.HTTP2, ref.Family.Proto)
	assert.Equal(t, 0, ref.Index)
	assert.Equal(t, 1, ref.Sub)
	assert.Equal(t, "tcp.stream eq 0 and http2.streamid eq 1", ref.Filter())
	assert.Equal(t, "follow,http2,raw,0,1", ref.FollowArg("raw"))
}

// The frames that set a connection up - the preface and SETTINGS - belong to
// stream 0 and carry no http2.streamid field of their own. Following them
// would hand tshark one number where it needs two, which is an error and not
// an empty pane, so the refusal has to happen here and say what is missing.
func TestAnHTTP2PacketWithNoStreamIdIsRefusedWithItsName(t *testing.T) {
	want := family(t, "http2")
	pkt := fakePacket{
		layers:  map[string]bool{"tcp": true, "http2": true},
		indexes: map[string]int{"tcp.stream": 0},
	}

	_, err := pickStreamFamily(&want, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "http2.streamid")
	assert.Contains(t, err.Error(), "multiplexes")
}

// And such a packet is not offered as a followable HTTP/2 stream either - the
// offer list has to apply the same rule as the choice.
func TestTheOfferListLeavesOutAFamilyMissingItsSecondNumber(t *testing.T) {
	want := family(t, "tls")
	pkt := fakePacket{
		layers:  map[string]bool{"tcp": true, "http2": true},
		indexes: map[string]int{"tcp.stream": 0},
	}

	_, err := pickStreamFamily(&want, pkt)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ":streams tcp")
	assert.NotContains(t, err.Error(), ":streams http2")
}

// The refusal lists what this packet does offer, so that a user who asked for
// the wrong family is told the right one rather than just told no.
func TestTheRefusalNamesEveryFamilyThePacketOffers(t *testing.T) {
	want := family(t, "tls")

	_, err := pickStreamFamily(&want, websocketPacket())

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
