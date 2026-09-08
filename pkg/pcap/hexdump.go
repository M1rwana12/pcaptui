// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

//======================================================================

// HexDumpLine reads one line of `tshark -x` output and returns the bytes it
// carries.
//
// A line looks like this:
//
//	0040  30 30 20 4f 4b 0d 0a 43 6f 6e 74 65 6e 74 2d 54   00 OK..Content-T
//	^^^^  ^-------------------- 48 columns ------------^    ^ the same bytes
//	offset          the hex column                            as text
//
// The bytes appear twice: once as hex and once as text. Scanning the whole
// line for "two hex digits and a space" therefore also matches the text
// column - in the line above, "00 " at the start of "00 OK" - and splices a
// byte that is not there into the middle of the packet. Everything after it
// is shifted by one, and the layer highlighting derived from the protocol
// tree stops lining up with the bytes it is meant to mark.
//
// So the hex column is read by position instead. It begins after the offset
// and its separating spaces, and is exactly 16 groups of three characters
// wide, space-padded on a short final line.
//
// ok is false for a line that is not a hexdump line at all, which is how the
// caller detects the boundary between one packet and the next.
func HexDumpLine(line string) (bytes []byte, ok bool) {
	i := 0

	// The offset.
	for i < len(line) && isHexDigit(line[i]) {
		i++
	}
	if i == 0 {
		return nil, false
	}

	// The spaces after it.
	spaces := 0
	for i < len(line) && line[i] == ' ' {
		i++
		spaces++
	}
	if spaces == 0 {
		return nil, false
	}

	bytes = make([]byte, 0, hexBytesPerLine)

	for n := 0; n < hexBytesPerLine; n++ {
		p := i + n*3
		if p+1 >= len(line) {
			break
		}

		hi, lo := line[p], line[p+1]
		if !isHexDigit(hi) || !isHexDigit(lo) {
			break // the column is space-padded once the bytes run out
		}

		bytes = append(bytes, hexValue(hi)<<4|hexValue(lo))
	}

	return bytes, len(bytes) > 0
}

// hexBytesPerLine is how many bytes tshark puts on one line of -x output. It
// also fixes the width of the hex column, which is what lets the text column
// be skipped by position.
const hexBytesPerLine = 16

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hexValue(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
