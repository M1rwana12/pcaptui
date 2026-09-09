// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package capinfo

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/m1rwana12/pcaptui"
)

//======================================================================

// Info is what capinfos says about a capture file, as it said it.
//
// Every field is the text capinfos printed, not a number. It abbreviates for a
// reader - "376 k" packets, "44 MB", and on this machine "39,571274 seconds"
// with a decimal comma - so these are facts to show, not quantities to
// compute with. Anything that needs arithmetic should be asking tshark.
type Info struct {
	Packets  string // "7", "376 k"
	Size     string // "1984 bytes", "44 MB"
	DataSize string // the packet bytes, without the file's own overhead
	Duration string // "6,000000000 seconds"
	Earliest string // "2026-01-01 10:00:00,000000000"
	Latest   string
	Rate     string // "2 packets/s"
	SnapLen  string // "file hdr: 1514 bytes", or "file hdr: (not set)"
	Encap    string // "Ethernet"
}

// Empty is whether nothing was read at all, which is what a failed or missing
// capinfos looks like from here.
func (i Info) Empty() bool {
	return i == Info{}
}

//======================================================================

// Parse reads the output of capinfos.
//
// The format is "Key:" then the value, one per line. Only lines whose key
// starts at the left margin are read: capinfos repeats several of them,
// indented, inside a per-interface block at the end, and a capture with two
// interfaces would otherwise overwrite the file's own totals with the last
// interface's.
func Parse(out string) Info {
	var res Info

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		switch strings.TrimSpace(key) {
		case "Number of packets":
			res.Packets = value
		case "File size":
			res.Size = value
		case "Data size":
			res.DataSize = value
		case "Capture duration":
			res.Duration = value
		case "File encapsulation":
			res.Encap = value
		case "Average packet rate":
			res.Rate = value

		// These two carry a time, so the value has colons of its own and
		// cutting at the first one leaves the rest behind.
		case "Earliest packet time":
			res.Earliest = after(line, "Earliest packet time:")
		case "Latest packet time":
			res.Latest = after(line, "Latest packet time:")

		// And this one carries a second key: "file hdr: 1514 bytes".
		case "Packet size limit":
			res.SnapLen = after(line, "Packet size limit:")
		}
	}

	return res
}

// after is everything following a prefix, trimmed - for the values that
// contain a colon themselves.
func after(line string, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}

//======================================================================

// Read runs capinfos over one file and returns what it said about it.
//
// Synchronous, unlike the loader beside it, because it is neither slow nor
// streaming: measured on a 44 MB capture, capinfos took 0.27 s where the -z
// pass over the same file took 9.1 to 11.6 s. The caller still has to keep it
// off the goroutine that draws.
func Read(pcapfile string) (Info, error) {
	cmd := exec.Command(pcaptui.CapinfosBin(), pcapfile)

	out, err := cmd.Output()
	if err != nil {
		return Info{}, fmt.Errorf("could not read the properties of %s: %v",
			pcapfile, err)
	}

	return Parse(string(out)), nil
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
