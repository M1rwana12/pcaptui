// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package streams

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/m1rwana12/pcaptui/internal/tsharktest"
	"github.com/m1rwana12/pcaptui/pkg/summary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

var testWG sync.WaitGroup

// pcap.Command.Start starts a summary goroutine, and that package tracks its
// goroutines through a package-level WaitGroup that cmd/pcaptui sets up. A
// test binary that runs a real tshark has to do the same or Start panics on a
// nil WaitGroup - which is what happened when these tests were first written.
func TestMain(m *testing.M) {
	summary.Goroutinewg = &testWG
	Goroutinewg = &testWG

	code := m.Run()

	testWG.Wait()
	os.Exit(code)
}

//======================================================================

func TestEveryFamilyHasANameAndATable(t *testing.T) {
	for _, f := range Families() {
		assert.NotEmpty(t, f.Token, "family %v has no tshark token", f.Proto)
		assert.NotEmpty(t, f.Label, "family %s has no label", f.Token)
		assert.NotEmpty(t, f.IndexField, "family %s has no index field", f.Token)
		assert.NotEmpty(t, f.Layer, "family %s has no PDML layer", f.Token)
		assert.Contains(t, []string{"tcp", "udp"}, f.Transport,
			"family %s has no transport whose length field says a packet carried payload", f.Token)

		// The label comes from Protocol.String(), which panics on a value it
		// does not know - and the stream view's title is built from it, so the
		// panic is reachable from drawing the screen.
		assert.Equal(t, f.Proto.String(), f.Label,
			"the enum and the table disagree about what to call %s", f.Token)

		assert.Equal(t, f, FamilyOf(f.Proto))

		byToken, ok := FamilyByToken(f.Token)
		require.True(t, ok, "%s is in the table but not findable by its token", f.Token)
		assert.Equal(t, f, byToken)
	}

	_, ok := FamilyByToken("nosuchthing")
	assert.False(t, ok)
}

// The filter is the same expression tshark prints in its own Filter: line, and
// this program puts it in the packet list. If the two disagree, the list is
// showing a different set of packets than the pane beside it.
func TestTheFilterIsSpelledTheWayTsharkSpellsIt(t *testing.T) {
	for _, tc := range []struct {
		token string
		idx   int
		want  string
	}{
		{"tcp", 0, "tcp.stream eq 0"},
		{"udp", 3, "udp.stream eq 3"},
		// The spelling from Wireshark 4.4 onwards; Resolve handles the builds
		// before it, which have no tls.stream at all.
		{"tls", 1, "tls.stream eq 1"},
		// Not websocket.stream: Wireshark has no such field, and tshark's own
		// filter for follow,websocket is the TCP stream underneath.
		{"websocket", 2, "tcp.stream eq 2"},
	} {
		f, ok := FamilyByToken(tc.token)
		require.True(t, ok, tc.token)
		assert.Equal(t, tc.want, f.Filter(tc.idx))
	}
}

// tls.stream arrived in Wireshark 4.4. In 4.2 the TLS follow tap exists and
// reports tcp.stream as its filter, and a display filter naming a field the
// build does not know is rejected rather than empty - so composing
// `tls.stream eq 0` there breaks the index pass instead of degrading.
func TestAnIndexFieldThisTsharkLacksFallsBackToTheTransport(t *testing.T) {
	without := func(missing string) func(string) bool {
		return func(name string) bool { return name != missing }
	}

	tls, ok := FamilyByToken("tls")
	require.True(t, ok)

	assert.Equal(t, "tls.stream", tls.Resolve(without("nothing")).IndexField,
		"a build that has the field keeps it")
	assert.Equal(t, "tcp.stream", tls.Resolve(without("tls.stream")).IndexField,
		"a build without it is told the transport's field, which is what its own tap reports")
	assert.Equal(t, "tls.stream", tls.Resolve(nil).IndexField,
		"nil means the field list has not loaded yet - assume the field is there")

	// The families whose index field is already the transport's cannot move.
	for _, token := range []string{"tcp", "udp", "websocket"} {
		f, ok := FamilyByToken(token)
		require.True(t, ok, token)
		assert.Equal(t, f.IndexField, f.Resolve(without(f.IndexField)).IndexField,
			"%s has nothing to fall back to", token)
	}
}

// The families layered over a transport are opt-in, and the reason is written
// into the table: one of them can answer "nothing" for a stream that exists.
func TestALayeredFamilyIsNeverChosenForTheUser(t *testing.T) {
	for _, tc := range []struct {
		token    string
		explicit bool
	}{
		{"tcp", false},
		{"udp", false},
		{"tls", true},
		{"websocket", true},
	} {
		f, ok := FamilyByToken(tc.token)
		require.True(t, ok, tc.token)
		assert.Equal(t, tc.explicit, f.Explicit, "%s", tc.token)
	}

	tls, _ := FamilyByToken("tls")
	assert.NotEmpty(t, tls.EmptyReason,
		"TLS can come back empty for a stream that exists; the user has to be told why")
	assert.Contains(t, tls.EmptyReason, "keylog")
}

// Every family's token has to survive the header parser, because tshark echoes
// it in the Follow: line of its own output. This is not hypothetical: the
// grammar's FollowExpr is [a-zA-Z,]+, so `Follow: http2,raw` fails to parse -
// which is one of the two reasons http2 is not in the table.
func TestTheParserAcceptsEveryFamilysHeader(t *testing.T) {
	for _, f := range Families() {
		inp := fmt.Sprintf(`

===================================================================
Follow: %s,raw
Filter: %s
Node 0: 10.0.0.5:51000
Node 1: 10.0.0.80:80
48656c6c6f
===================================================================
`, f.Token, f.Filter(0))

		_, err := ParseReader("", strings.NewReader(inp))
		assert.NoError(t, err, "the parser cannot read a %s stream's own header", f.Token)
	}
}

func TestAFamilyTheParserCannotReadIsNotInTheTable(t *testing.T) {
	// The guard rail for the test above: prove the parser really does refuse a
	// name with a digit in it, so that "every family parses" means something.
	inp := `

===================================================================
Follow: http2,raw
Filter: tcp.stream eq 0 and http2.streamid eq 1
Node 0: 10.0.0.5:51000
Node 1: 10.0.0.80:80
48656c6c6f
===================================================================
`
	_, err := ParseReader("", strings.NewReader(inp))
	require.Error(t, err, "if this passes, http2 can be added to the table")

	_, ok := FamilyByToken("http2")
	assert.False(t, ok)
}

// tshark writes CRLF on Windows, and a header field captured up to the newline
// keeps the carriage return. The node addresses are drawn into the
// conversation menu and into the "client → server (N bytes)" line, so a stray
// CR there sends the cursor back to the start of the row mid-text. Written as
// a fixed string rather than by running tshark, so the case is covered on the
// platforms that do not produce it.
func TestTheHeaderLosesItsLineEndings(t *testing.T) {
	inp := "\r\n===================================================================\r\n" +
		"Follow: tcp,raw\r\n" +
		"Filter: tcp.stream eq 0\r\n" +
		"Node 0: 10.0.0.5:51000\r\n" +
		"Node 1: 10.0.0.80:80\r\n" +
		"48656c6c6f\r\n" +
		"===================================================================\r\n"

	got := &chunkCollector{}
	_, err := ParseReader("", strings.NewReader(inp), GlobalStore("callbacks", got))
	require.NoError(t, err)

	assert.Equal(t, "tcp,raw", got.header.Follow)
	assert.Equal(t, "tcp.stream eq 0", got.header.Filter)
	assert.Equal(t, "10.0.0.5:51000", got.header.Node0)
	assert.Equal(t, "10.0.0.80:80", got.header.Node1)
}

//======================================================================

type chunkCollector struct {
	header FollowHeader
	chunks []IChunk
}

var _ IOnStreamChunk = (*chunkCollector)(nil)
var _ IOnStreamHeader = (*chunkCollector)(nil)

func (c *chunkCollector) OnStreamChunk(chunk IChunk) { c.chunks = append(c.chunks, chunk) }

// Clean, because tshark writes CRLF on Windows and the grammar keeps the
// carriage return - the same call the interface makes before the addresses
// reach the screen. See TestTheHeaderLosesItsLineEndings.
func (c *chunkCollector) OnStreamHeader(header FollowHeader) { c.header = header.Clean() }

// fieldExists asks this tshark whether it knows a field name, which is what
// decides the spelling of a TLS stream's filter. -T fields with an unknown
// field is an error naming it, so running it is the whole test.
func fieldExists(t *testing.T) func(string) bool {
	t.Helper()
	return func(name string) bool {
		cmd := exec.Command("tshark", "-r", "../../scripts/pcaps/tls.pcap",
			"-T", "fields", "-e", name)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return cmd.Run() == nil
	}
}

func followWith(t *testing.T, pcap string, f Family, idx int) *chunkCollector {
	t.Helper()
	tsharktest.Need(t)

	cmd := MakeCommands().Stream(pcap, f.Token, idx)
	out, err := cmd.StdoutReader()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	cmd.Wait()

	res := &chunkCollector{}
	_, perr := ParseReader("", bytes.NewReader(buf.Bytes()), GlobalStore("callbacks", res))
	require.NoError(t, perr, "tshark's own output could not be parsed:\n%s", buf.String())

	return res
}

// The argument can be spelled the way this package intends and still not be
// the spelling tshark accepts - and its output can still be a shape the parser
// cannot read. Only tshark can settle either.
func TestAWebSocketStreamIsTheMessagesNotTheFraming(t *testing.T) {
	ws, ok := FamilyByToken("websocket")
	require.True(t, ok)

	got := followWith(t, "../../scripts/pcaps/websocket.pcap", ws, 0)

	require.Len(t, got.chunks, 2, "one message each way")
	assert.Equal(t, "Hello", string(got.chunks[0].StreamData()))
	assert.Equal(t, Client, got.chunks[0].Direction())
	assert.Equal(t, "Hi there", string(got.chunks[1].StreamData()))
	assert.Equal(t, Server, got.chunks[1].Direction())

	// The client frame is masked on the wire; following the TCP stream instead
	// gives the handshake and the mask, which is the whole point of asking for
	// the WebSocket family.
	assert.Equal(t, ws.Resolve(fieldExists(t)).Filter(0), got.header.Filter,
		"the filter this program composes is not the one tshark reports")

	tcp, _ := FamilyByToken("tcp")
	raw := followWith(t, "../../scripts/pcaps/websocket.pcap", tcp, 0)
	require.NotEmpty(t, raw.chunks)
	assert.Contains(t, string(raw.chunks[0].StreamData()), "GET /chat",
		"the TCP stream carries the handshake")
	assert.NotContains(t, string(raw.chunks[0].StreamData()), "Hello",
		"and the client's message masked, not in the clear")
}

// This is the measurement :streams tls exists the way it does because of. If a
// future Wireshark starts handing back something for a stream it has no keys
// for, this test fails and says the design can change.
func TestFollowingTLSWithoutKeysIsEmptyAndSaysNothingAboutIt(t *testing.T) {
	tls, ok := FamilyByToken("tls")
	require.True(t, ok)

	got := followWith(t, "../../scripts/pcaps/tls.pcap", tls, 0)

	assert.Empty(t, got.chunks,
		"tshark returned TLS payload for a capture with no key log")

	// Not a hard-coded field name: this is tls.stream from Wireshark 4.4 and
	// tcp.stream before it, and the contract is that the two agree.
	assert.Equal(t, tls.Resolve(fieldExists(t)).Filter(0), got.header.Filter,
		"the filter this program composes is not the one tshark reports")
	// The banner is complete and the addresses are empty - the same answer
	// tshark gives for a stream that does not exist at all.
	assert.Equal(t, ":0", got.header.Node0)
	assert.Equal(t, ":0", got.header.Node1)

	// The same stream, followed as TCP, has the bytes.
	tcp, _ := FamilyByToken("tcp")
	raw := followWith(t, "../../scripts/pcaps/tls.pcap", tcp, 0)
	require.NotEmpty(t, raw.chunks, "the TCP stream is empty too - the fixture is broken")
	assert.NotEqual(t, ":0", raw.header.Node0)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
