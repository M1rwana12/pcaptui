// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func TestHexDumpLineReadsAFullLine(t *testing.T) {
	line := "0000  20 52 45 43 56 00 20 53 45 4e 44 00 08 00 45 00    RECV. SEND...E."

	got, ok := HexDumpLine(line)

	assert.True(t, ok)
	assert.Equal(t, []byte{
		0x20, 0x52, 0x45, 0x43, 0x56, 0x00, 0x20, 0x53,
		0x45, 0x4e, 0x44, 0x00, 0x08, 0x00, 0x45, 0x00,
	}, got)
}

// This is the line that made the old parser wrong. The text column starts
// "00 OK", and "00 " is two hex digits and a space, so scanning the whole line
// produced a seventeenth byte that is not in the packet - and everything after
// it in the packet shifted by one.
func TestTextColumnIsNotMistakenForBytes(t *testing.T) {
	line := "0040  30 30 20 4f 4b 0d 0a 43 6f 6e 74 65 6e 74 2d 54   00 OK..Content-T"

	got, ok := HexDumpLine(line)

	assert.True(t, ok)
	assert.Len(t, got, 16, "the line holds sixteen bytes, not seventeen")
	assert.Equal(t, byte(0x54), got[15], "the last byte is 0x54, not a stray 0x00")
}

// Any text at all can look like hex. "de ad be ef" in a payload, a hostname
// ending in "ab ", a hex string a program logged.
func TestTextThatLooksEntirelyLikeHexIsStillIgnored(t *testing.T) {
	line := "0010  61 62 63 64 65 66 30 31 32 33 34 35 36 37 38 39   de ad be ef 0123"

	got, _ := HexDumpLine(line)

	assert.Len(t, got, 16)
}

// The last line of a packet is short and the hex column is space-padded.
func TestShortFinalLine(t *testing.T) {
	line := "00b0  73 3e                                             s>"

	got, ok := HexDumpLine(line)

	assert.True(t, ok)
	assert.Equal(t, []byte{0x73, 0x3e}, got)
}

func TestLinesThatAreNotHexDumpAreRejected(t *testing.T) {
	for _, line := range []string{
		"",
		"    ",
		"Frame 1: 137 bytes on wire",
		"zz  61 62",     // no offset
		"0000",          // offset with nothing after it
		"0000 ",         // offset, one space, no bytes
		"  0000  61 62", // leading space: not how tshark writes it
	} {
		_, ok := HexDumpLine(line)
		assert.False(t, ok, "should have rejected %q", line)
	}
}

// tshark widens the offset once a packet is longer than 0xffff bytes.
func TestWideOffsets(t *testing.T) {
	got, ok := HexDumpLine("012340  ab cd                                           ..")

	assert.True(t, ok)
	assert.Equal(t, []byte{0xab, 0xcd}, got)
}

func TestUppercaseHexIsAccepted(t *testing.T) {
	got, ok := HexDumpLine("0000  AB CD EF                                        ...")

	assert.True(t, ok)
	assert.Equal(t, []byte{0xAB, 0xCD, 0xEF}, got)
}

//======================================================================

// feed runs a whole -x dump through the reader and returns one slice per
// packet.
func feed(dump string) [][]byte {
	var res [][]byte
	h := &HexDumpReader{}

	// Split the way the loader reads: one line at a time with its newline
	// kept, and no empty line invented after the last one.
	lines := strings.SplitAfter(dump, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for _, line := range lines {
		if packet, done := h.Line(line); done {
			res = append(res, packet)
		}
	}

	return res
}

// A packet with one data source has no heading at all - which is every packet
// in a capture with no reassembly.
func TestAPacketWithOneSourceIsReadWhole(t *testing.T) {
	packets := feed("0000  01 02 03 04\n\n")

	require.Len(t, packets, 1)
	assert.Equal(t, []byte{1, 2, 3, 4}, packets[0])
}

// Real output of `tshark -x` for the second packet of a two-segment HTTP
// response. Note the heading above the *first* source: tshark labels every
// source once there is more than one, and reading the first label as "stop"
// left the hex pane empty for every reassembled packet.
const twoSources = `Packet (20 bytes):
0000  aa bb cc dd
Reassembled TCP (49 bytes):
0000  48 54 54 50

`

func TestTheFrameIsReadAndTheReassemblyIsNot(t *testing.T) {
	packets := feed(twoSources)

	require.Len(t, packets, 1, "the second source is not a packet of its own")
	assert.Equal(t, []byte{0xaa, 0xbb, 0xcc, 0xdd}, packets[0],
		"the reassembled bytes are the same packet seen another way, not more of it")
}

// Three sources happens with decrypted TLS inside a reassembled stream.
func TestOnlyTheFirstOfSeveralSourcesIsRead(t *testing.T) {
	packets := feed(`Frame (8 bytes):
0000  11 22
Reassembled TCP (4 bytes):
0000  33 44
Decrypted TLS (2 bytes):
0000  55 66

`)

	require.Len(t, packets, 1)
	assert.Equal(t, []byte{0x11, 0x22}, packets[0])
}

// The blank line is the only thing that ends a packet, so a heading in the
// middle must not shift every row after it by one.
func TestASourceHeadingDoesNotEndAPacket(t *testing.T) {
	packets := feed(twoSources + "0000  09 08 07\n\n")

	require.Len(t, packets, 2)
	assert.Equal(t, []byte{0xaa, 0xbb, 0xcc, 0xdd}, packets[0])
	assert.Equal(t, []byte{9, 8, 7}, packets[1])
}

// The reader is reused across a whole load, so what one packet saw must not
// leak into the next.
func TestASecondPacketStartsClean(t *testing.T) {
	packets := feed(twoSources + "Packet (3 bytes):\n0000  de ad\nReassembled TCP (1 bytes):\n0000  ff\n\n")

	require.Len(t, packets, 2)
	assert.Equal(t, []byte{0xde, 0xad}, packets[1],
		"the second packet's own heading is its first source, not a leftover count")
}

// A capture read to the end without a trailing blank line loses nothing that
// was already delivered.
func TestPacketsAreDeliveredOnTheirBlankLine(t *testing.T) {
	packets := feed("0000  01\n\n0000  02\n")

	require.Len(t, packets, 1, "the last packet has no blank line to end it")
	assert.Equal(t, []byte{1}, packets[0])
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
