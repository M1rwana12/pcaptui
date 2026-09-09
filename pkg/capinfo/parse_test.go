// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package capinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// Real capinfos output, from the capture this project tests against. The
// decimal comma is not a typo: it is what this machine prints, which is why
// none of these values is parsed into a number.
const telnetInfo = `File name:           scripts/pcaps/telnet-cooked.pcap
File type:           Wireshark/tcpdump/... - pcap
File encapsulation:  Ethernet
File timestamp precision:  microseconds (6)
Packet size limit:   file hdr: 1514 bytes
Number of packets:   92
File size:           9244 bytes
Data size:           7748 bytes
Capture duration:    39,571274 seconds
Earliest packet time: 1999-11-28 04:12:38,387203
Latest packet time:   1999-11-28 04:13:17,958477
Data byte rate:      195 bytes/s
Data bit rate:       1566 bits/s
Average packet size: 84,22 bytes
Average packet rate: 2 packets/s
SHA256:              ae870805f1e5f6a2621b1f6e1e0229b47cc96d917f42c215acbcfcd46f9d72fc
Strict time order:   True
Number of interfaces in file: 1
Interface #0 info:
                     Encapsulation = Ethernet (1 - ether)
                     Capture length = 1514
                     Number of packets = 92
`

func TestTheFactsAreRead(t *testing.T) {
	got := Parse(telnetInfo)

	assert.Equal(t, "92", got.Packets)
	assert.Equal(t, "9244 bytes", got.Size)
	assert.Equal(t, "7748 bytes", got.DataSize)
	assert.Equal(t, "39,571274 seconds", got.Duration)
	assert.Equal(t, "Ethernet", got.Encap)
	assert.Equal(t, "2 packets/s", got.Rate)
}

// A time has colons of its own, so cutting the line at the first one would
// leave most of the value behind.
func TestATimeKeepsItsColons(t *testing.T) {
	got := Parse(telnetInfo)

	assert.Equal(t, "1999-11-28 04:12:38,387203", got.Earliest)
	assert.Equal(t, "1999-11-28 04:13:17,958477", got.Latest)
}

// So does the snapshot length, which carries a second key inside its value.
func TestTheSnapshotLimitKeepsItsOwnKey(t *testing.T) {
	got := Parse(telnetInfo)

	assert.Equal(t, "file hdr: 1514 bytes", got.SnapLen)
}

// capinfos repeats several keys, indented, inside a per-interface block. A
// capture with two interfaces would otherwise end up reporting the last
// interface's packet count as the file's.
func TestThePerInterfaceBlockIsNotRead(t *testing.T) {
	twoInterfaces := telnetInfo + `Interface #1 info:
                     Encapsulation = Ethernet (1 - ether)
                     Number of packets = 3
`
	got := Parse(twoInterfaces)

	assert.Equal(t, "92", got.Packets, "the file's total, not one interface's")
}

// capinfos abbreviates for a reader once a file is large, which is another
// reason these are text and not numbers.
func TestAnAbbreviatedValueIsKeptAsWritten(t *testing.T) {
	got := Parse("Number of packets:   376 k\nFile size:           44 MB\n")

	assert.Equal(t, "376 k", got.Packets)
	assert.Equal(t, "44 MB", got.Size)
}

func TestNothingReadableIsEmpty(t *testing.T) {
	assert.True(t, Parse("").Empty())
	assert.True(t, Parse("capinfos: cannot open file\n").Empty())
	assert.False(t, Parse(telnetInfo).Empty())
}

// A key with nothing after it says nothing, and an empty value on screen reads
// as a fact that was looked up and came back blank.
func TestAKeyWithNoValueIsSkipped(t *testing.T) {
	got := Parse("Number of packets:   \nFile size:           1984 bytes\n")

	assert.Equal(t, "", got.Packets)
	assert.Equal(t, "1984 bytes", got.Size)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
