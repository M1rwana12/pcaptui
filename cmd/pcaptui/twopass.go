// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/m1rwana12/pcaptui/pkg/pcap"
)

//======================================================================

// twoPassFor decides whether tshark can be asked to read the capture twice,
// and returns what to tell the user when it cannot.
//
// Two passes fill in the references that point forwards, which nothing reading
// in order can know: measured on a request and its answer, one pass leaves the
// request with no http.response_in and two passes name the frame that answered
// it. The ones pointing backwards - how long the response took, which request
// it belongs to - are there either way.
//
// It is not the default because tshark prints nothing at all until it has read
// the whole file that way. Measured on 376,000 packets, the first packet took
// 2.7 seconds to appear instead of 0.36, and that gap grows with the file
// while the total cost is only 8% more. Packets on screen immediately is the
// thing this program is for.
//
// It needs a file. tshark refuses a pipe outright - "TShark can't read pipe or
// FIFO files in two-pass mode" - and a live capture is being written while it
// is read, so reading it twice would mean reading two different things.
func twoPassFor(want bool, srcs []pcap.IPacketSource) (bool, string) {
	if !want {
		return false, ""
	}

	if len(pcap.FileSystemSources(srcs)) == 0 {
		return false, fmt.Sprintf(
			"Two-pass analysis needs a file to read twice; ignoring it for %s.",
			pcap.SourcesString(srcs))
	}

	return true, ""
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
