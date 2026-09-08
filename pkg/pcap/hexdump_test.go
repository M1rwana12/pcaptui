// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
// Local Variables:
// mode: Go
// fill-column: 78
// End:
