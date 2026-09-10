// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package streams

import "fmt"

//======================================================================

// A Family is one of tshark's -z follow,<family> taps, together with the four
// things this program has to know to drive it.
//
// Measured against tshark 4.6.8, which offers thirteen of them. Four are here;
// the rest are left out for reasons written down beside them below.
type Family struct {
	Proto Protocol

	// Token is the spelling tshark wants in -z follow,<token>,raw,<idx>. It is
	// also the spelling the parser has to accept in the "Follow:" header line.
	Token string

	// Label is what the user is told - it appears in the view's title.
	Label string

	// IndexField is the display filter field that carries the stream index.
	// It is not always the family's own name: a WebSocket stream is indexed by
	// the TCP stream underneath it, and tshark itself emits `tcp.stream eq N`
	// as the filter for follow,websocket. There is no websocket.stream field
	// at all - asking tshark for one is an error.
	IndexField string

	// Transport is the family whose length field says whether a packet carried
	// payload, which is how the indexer decides which packets a chunk can jump
	// to. For everything layered over TCP that is tcp.
	Transport string

	// Layer is the PDML <proto name=...> that must be present for this family
	// to apply to a packet. For tcp and udp it is the same as Token; for the
	// families layered on top it is what distinguishes them from plain TCP.
	Layer string

	// EmptyReason, when set, is why this family can legitimately come back
	// with nothing for a stream that plainly exists - said to the user instead
	// of leaving them with an empty pane.
	//
	// This is not a hypothetical. Measured on a synthesised TLS capture:
	// `-z follow,tls,raw,0` with no key log answers with a complete banner,
	// `Node 0: :0`, no payload and exit status 0 - byte for byte the shape of
	// "there is no such stream" - while `-z follow,tcp,raw,0` over the same
	// stream returns the bytes.
	EmptyReason string

	// Explicit families are not chosen for the user. See PickFor.
	Explicit bool
}

// The order matters: PickFor walks this list and takes the first family whose
// layer the packet has, so the transports come first.
var families = []Family{
	{
		Proto:      TCP,
		Token:      "tcp",
		Label:      "TCP",
		IndexField: "tcp.stream",
		Transport:  "tcp",
		Layer:      "tcp",
	},
	{
		Proto:      UDP,
		Token:      "udp",
		Label:      "UDP",
		IndexField: "udp.stream",
		Transport:  "udp",
		Layer:      "udp",
	},
	{
		Proto:      TLS,
		Token:      "tls",
		Label:      "TLS",
		IndexField: "tls.stream",
		Transport:  "tcp",
		Layer:      "tls",
		Explicit:   true,
		EmptyReason: "TLS following shows the decrypted payload, so without a key log there is " +
			"nothing for it to show. Start with --tls-keylog, or follow the TCP stream to see " +
			"the encrypted bytes.",
	},
	{
		Proto:      WebSocket,
		Token:      "websocket",
		Label:      "WebSocket",
		IndexField: "tcp.stream",
		Transport:  "tcp",
		Layer:      "websocket",
		Explicit:   true,
	},
}

// Families layered over a transport are asked for, never guessed at. A packet
// carrying TLS is also a TCP packet, and choosing TLS for it would replace the
// bytes the user used to get with an empty pane whenever there is no key log -
// which is the common case, and which tshark reports as success.
//
// Deliberately absent, with the measurement that rules each one out:
//
//   - http2 and quic need TWO indexes - `follow,http2,raw,<tcp stream>,<http2
//     streamid>`. One index is refused: "Error creating filter for this
//     stream", exit 1. The loader's Stream/Indexer/StartLoad take a single
//     int, and follow.peg's FollowExpr is [a-zA-Z,]+, which rejects the digit
//     in a `Follow: http2,raw` header - verified by feeding one to ParseReader.
//   - mp2t and mpeg-pes need two indexes as well, spell their filter with
//     ==/&& rather than eq/and, and give the PID in decimal while printing it
//     as hex.
//   - sip is registered by tshark and unusable from the command line: every
//     index form, and the Call-ID form, is refused with exit 1.
//   - http reverses the node order relative to follow,tcp on the same stream
//     and returns the reassembled body as a single untabbed chunk, losing the
//     request - it is a worse answer than the TCP stream, not a better one.
//   - dccp, dtls and usbcom are transports this program has no captures of and
//     no way to make one on this machine.
func Families() []Family {
	res := make([]Family, len(families))
	copy(res, families)
	return res
}

// FamilyOf returns the family for a Protocol. Every value of the enum has one;
// a value that does not is a programming error and says so.
func FamilyOf(p Protocol) Family {
	for _, f := range families {
		if f.Proto == p {
			return f
		}
	}
	panic(fmt.Sprintf("no stream family for protocol %v", int(p)))
}

// FamilyByToken looks up what the user typed after :streams.
func FamilyByToken(tok string) (Family, bool) {
	for _, f := range families {
		if f.Token == tok {
			return f, true
		}
	}
	return Family{}, false
}

// Tokens is the list offered for completion and named in error messages.
func Tokens() []string {
	res := make([]string, 0, len(families))
	for _, f := range families {
		res = append(res, f.Token)
	}
	return res
}

// Filter is the display filter for one stream of this family. It is the same
// expression tshark prints in the Filter: line of its own output, which is
// what makes the packet list and the stream view agree about what is being
// shown.
func (f Family) Filter(idx int) string {
	return fmt.Sprintf("%s eq %d", f.IndexField, idx)
}

// Resolve adjusts the index field to the fields this tshark actually has.
//
// tls.stream does not exist before Wireshark 4.4. The TLS follow tap is there
// in 4.2 and reports `tcp.stream eq N` as its own filter - measured on the CI
// runner's 4.2.2, against 4.6.8 here, which reports tls.stream. That matters
// beyond a label: a display filter naming a field this build does not know is
// rejected outright, so the index pass would fail rather than come back empty,
// and following a TLS stream would break on a Wireshark this program is
// supposed to support.
//
// known answers whether a field name exists; nil means "assume it does",
// which is what happens before the field list has loaded.
func (f Family) Resolve(known func(string) bool) Family {
	if known == nil || known(f.IndexField) {
		return f
	}

	f.IndexField = f.Transport + ".stream"
	return f
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
