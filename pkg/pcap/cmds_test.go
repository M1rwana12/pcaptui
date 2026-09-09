// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// These cover how tshark is invoked. That is where this program has broken
// before and where it will break again: everything it knows about a capture
// arrives through these argument lists, and a wrong one fails quietly - tshark
// starts, exits zero, and returns something subtly different from what was
// asked for.
//
// The file that used to live here tested the live capture loader behind a
// `tshark` build tag nobody set. It had not compiled for years, referring to
// helpers that no longer existed. A test that cannot build is worse than none,
// because it looks like coverage.

func args(t *testing.T, c interface{}) []string {
	t.Helper()

	cmd, ok := c.(*Command)
	assert.True(t, ok, "expected a *Command")

	// Args[0] is the program itself.
	return cmd.Cmd.Args[1:]
}

func joined(t *testing.T, c interface{}) string {
	t.Helper()
	return strings.Join(args(t, c), " ")
}

// index returns the position of the first occurrence of v, or -1.
func index(as []string, v string) int {
	for i, a := range as {
		if a == v {
			return i
		}
	}
	return -1
}

//======================================================================

func TestPsmlReadsTheNamedFile(t *testing.T) {
	got := joined(t, MakeCommands(nil, nil, nil, nil, false).Psml("capture.pcap", ""))

	assert.Contains(t, got, "-T psml")
	assert.Contains(t, got, "-r capture.pcap")
	assert.NotContains(t, got, " -l ", "-l is for streaming, not for a file")
}

// The loader maps a row in the table back to a packet number through the first
// column, so the No. column is added whatever the user's configuration says.
// Losing it does not fail loudly - it makes every packet number wrong once a
// display filter is in effect.
func TestPsmlAlwaysAsksForTheNumberColumnFirst(t *testing.T) {
	got := joined(t, MakeCommands(nil, nil, nil, nil, false).Psml("capture.pcap", ""))

	assert.Contains(t, got, `gui.column.format:"No.","%m"`)
}

func TestPsmlPassesTheDisplayFilter(t *testing.T) {
	as := args(t, MakeCommands(nil, nil, nil, nil, false).Psml("capture.pcap", "tcp.port == 443"))

	i := index(as, "-Y")
	assert.NotEqual(t, -1, i, "no -Y in %v", as)
	assert.Equal(t, "tcp.port == 443", as[i+1])
}

func TestPsmlColoursOnlyWhenAsked(t *testing.T) {
	assert.NotContains(t, args(t, MakeCommands(nil, nil, nil, nil, false).Psml("c.pcap", "")), "--color")
	assert.Contains(t, args(t, MakeCommands(nil, nil, nil, nil, true).Psml("c.pcap", "")), "--color")
}

//======================================================================

func TestPdmlAsksForTheProtocolTree(t *testing.T) {
	got := joined(t, MakeCommands(nil, nil, nil, nil, false).Pdml("capture.pcap", ""))

	assert.Contains(t, got, "-T pdml")
	assert.Contains(t, got, "-r capture.pcap")
}

func TestPcapAsksForBytes(t *testing.T) {
	got := joined(t, MakeCommands(nil, nil, nil, nil, false).Pcap("capture.pcap", ""))

	assert.Contains(t, got, "-r capture.pcap")
	assert.Contains(t, got, "-x", "without -x tshark writes one line of text per packet")
}

//======================================================================

// tshark-args is documented as applying to every invocation, and TLS
// decryption rides on it: --tls-keylog becomes -o tls.keylog_file:... in this
// list. If it reached only some of the commands, the packet list would decrypt
// and the protocol tree would not, or the reverse.
func TestExtraArgsReachEveryCommand(t *testing.T) {
	extra := []string{"-o", "tls.keylog_file:/keys.log"}
	c := MakeCommands(nil, extra, nil, nil, false)

	for name, cmd := range map[string]interface{}{
		"psml": c.Psml("c.pcap", ""),
		"pdml": c.Pdml("c.pcap", ""),
		"pcap": c.Pcap("c.pcap", ""),
	} {
		assert.Contains(t, joined(t, cmd), "tls.keylog_file:/keys.log",
			"%s did not get the extra arguments", name)
	}
}

func TestPsmlAndPdmlArgsGoToTheirOwnCommandOnly(t *testing.T) {
	c := MakeCommands(nil, nil, []string{"--pdml-only"}, []string{"--psml-only"}, false)

	assert.Contains(t, joined(t, c.Psml("c.pcap", "")), "--psml-only")
	assert.NotContains(t, joined(t, c.Psml("c.pcap", "")), "--pdml-only")

	assert.Contains(t, joined(t, c.Pdml("c.pcap", "")), "--pdml-only")
	assert.NotContains(t, joined(t, c.Pdml("c.pcap", "")), "--psml-only")
}

func TestDecodeAsIsRepeatedForEachRule(t *testing.T) {
	c := MakeCommands([]string{"udp.port==2055,cflow", "tcp.port==8080,http"}, nil, nil, nil, false)

	got := joined(t, c.Psml("c.pcap", ""))

	assert.Contains(t, got, "-d udp.port==2055,cflow")
	assert.Contains(t, got, "-d tcp.port==8080,http")
	assert.Equal(t, 2, strings.Count(got, "-d "))
}

//======================================================================

func TestIfaceCapturesFromEveryInterfaceGiven(t *testing.T) {
	cmd, ok := MakeCommands(nil, nil, nil, nil, false).
		Iface([]string{"eth0", "wlan0"}, "", "/tmp/out.pcap").(*Command)
	assert.True(t, ok)

	got := strings.Join(cmd.Cmd.Args[1:], " ")

	assert.Contains(t, got, "-i eth0")
	assert.Contains(t, got, "-i wlan0")
	assert.Contains(t, got, "-w /tmp/out.pcap")
}

func TestIfacePassesTheCaptureFilter(t *testing.T) {
	cmd := MakeCommands(nil, nil, nil, nil, false).
		Iface([]string{"eth0"}, "port 443", "/tmp/out.pcap").(*Command)

	as := cmd.Cmd.Args[1:]
	i := index(as, "-f")
	assert.NotEqual(t, -1, i, "no capture filter in %v", as)
	assert.Equal(t, "port 443", as[i+1])
}

// Capture re-runs this program, which then runs dumpcap and falls back to
// tshark. The environment variable is how the child knows to do that instead
// of drawing a user interface.
func TestIfaceMarksTheChildAsACapture(t *testing.T) {
	cmd := MakeCommands(nil, nil, nil, nil, false).
		Iface([]string{"eth0"}, "", "/tmp/out.pcap").(*Command)

	assert.Contains(t, cmd.Cmd.Env, "PCAPTUI_CAPTURE_MODE=1")
}

//======================================================================

// Two-pass analysis has to reach every command that reads the capture, or the
// packet list and the packet detail would disagree about the same packet.
func TestTwoPassReachesEveryReaderOfTheFile(t *testing.T) {
	c := MakeCommands(nil, nil, nil, nil, false)
	c.TwoPass = true

	assert.Contains(t, args(t, c.Psml("capture.pcap", "")), "-2")
	assert.Contains(t, args(t, c.Pdml("capture.pcap", "")), "-2")
	assert.Contains(t, args(t, c.Pcap("capture.pcap", "")), "-2")
}

func TestOnePassIsTheDefault(t *testing.T) {
	c := MakeCommands(nil, nil, nil, nil, false)

	assert.NotContains(t, args(t, c.Psml("capture.pcap", "")), "-2")
	assert.NotContains(t, args(t, c.Pdml("capture.pcap", "")), "-2")
	assert.NotContains(t, args(t, c.Pcap("capture.pcap", "")), "-2")
}

// tshark refuses two-pass mode on a pipe, and the refusal would arrive as a
// capture that failed to load rather than as an explanation. The caller is
// supposed to have decided already; this is the second guard.
func TestTwoPassIsNotAskedForOnAStream(t *testing.T) {
	c := MakeCommands(nil, nil, nil, nil, false)
	c.TwoPass = true

	got := args(t, c.Psml(strings.NewReader(""), ""))

	assert.NotContains(t, got, "-2")
	assert.Contains(t, got, "-l", "a stream is still read as it arrives")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
