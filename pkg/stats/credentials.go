// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"strings"
)

//======================================================================

// CredentialRow is one login `tshark -z credentials` found in the clear.
//
// Unlike every other statistic here, this one names the packet. The expert tap
// reports how often something happened and never where, so its rows have to be
// turned back into a display filter on their own text; a credential arrives
// with a frame number, so its filter is exact.
type CredentialRow struct {
	Packet   int
	Protocol string // "HTTP basic auth", "FTP", "POP", "IMAP", "SMTP", telnet
	Username string
	Info     string
}

// DisplayFilter is the packet this was found in.
func (r CredentialRow) DisplayFilter() string {
	return fmt.Sprintf("frame.number == %d", r.Packet)
}

//======================================================================

// ParseCredentials reads the output of `tshark -z credentials`:
//
//	Packet     Protocol         Username         Info
//	------     --------         --------         --------
//	1          HTTP basic auth  user
//
// By column position, taken off the header, because the protocol is several
// words - "HTTP basic auth" - and the username is whatever the user chose,
// which can be several words too. Splitting on whitespace would put half of a
// protocol into the username column.
func ParseCredentials(out string) []CredentialRow {
	var res []CredentialRow

	var cols []int

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		if cols == nil {
			cols = credentialColumns(line)
			continue
		}

		row, ok := credentialRow(line, cols)
		if ok {
			res = append(res, row)
		}
	}

	return res
}

// credentialColumns returns where each column starts, or nil for a line that
// is not the header.
func credentialColumns(header string) []int {
	if !strings.HasPrefix(strings.TrimSpace(header), "Packet") {
		return nil
	}

	var res []int
	for _, title := range []string{"Packet", "Protocol", "Username", "Info"} {
		i := strings.Index(header, title)
		if i < 0 {
			return nil
		}
		res = append(res, i)
	}

	return res
}

func credentialRow(line string, cols []int) (CredentialRow, bool) {
	// The rule under the header is drawn with dashes in the same columns, so
	// it is skipped by its first field failing to be a packet number rather
	// than by recognising the rule itself.
	packet, ok := parseCount(strings.TrimSpace(slice(line, cols[0], cols[1])))
	if !ok {
		return CredentialRow{}, false
	}

	return CredentialRow{
		Packet:   packet,
		Protocol: strings.TrimSpace(slice(line, cols[1], cols[2])),
		Username: strings.TrimSpace(slice(line, cols[2], cols[3])),
		Info:     strings.TrimSpace(slice(line, cols[3], len(line))),
	}, true
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
