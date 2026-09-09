// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/pcap"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestTwoPassIsOffUntilAskedFor(t *testing.T) {
	on, why := twoPassFor(false, []pcap.IPacketSource{pcap.FileSource{Filename: "capture.pcap"}})

	assert.False(t, on)
	assert.Empty(t, why, "nothing to explain when nothing was asked for")
}

func TestAFileCanBeReadTwice(t *testing.T) {
	on, why := twoPassFor(true, []pcap.IPacketSource{pcap.FileSource{Filename: "capture.pcap"}})

	assert.True(t, on)
	assert.Empty(t, why)
}

// tshark refuses two-pass mode on a pipe outright, and a live capture is being
// written while it is read - reading it twice would mean reading two different
// things.
func TestAStreamCannotBeReadTwice(t *testing.T) {
	for _, src := range []pcap.IPacketSource{
		pcap.InterfaceSource{Iface: "eth0"},
		pcap.FifoSource{Filename: "/tmp/fifo"},
		pcap.PipeSource{Descriptor: "/dev/fd/0", Fd: 0},
	} {
		on, why := twoPassFor(true, []pcap.IPacketSource{src})

		assert.False(t, on, "%T should not be read twice", src)
		assert.Contains(t, why, "needs a file to read twice",
			"%T should say why, not go quiet", src)
		assert.Contains(t, why, src.Name(), "the message should name the source")
	}
}

// A file among interfaces is still a file, and it is the one being read.
func TestAFileAmongStreamsIsStillAFile(t *testing.T) {
	on, why := twoPassFor(true, []pcap.IPacketSource{
		pcap.InterfaceSource{Iface: "eth0"},
		pcap.FileSource{Filename: "capture.pcap"},
	})

	assert.True(t, on)
	assert.Empty(t, why)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
