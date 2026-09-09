// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
)

//======================================================================

// splitHostPort separates the address from the port in what tshark prints for
// one end of a conversation.
//
// tshark joins them with a colon and does not bracket IPv6, so one end of a
// TCP conversation looks like this:
//
//	2001:db8::1:50000          <-> 2001:db8::2:80
//
// which is an address containing four colons and a port after the fifth. This
// used to split on every colon and require exactly two pieces, so **every IPv6
// conversation was dropped from the table** - not shown as unparsed, not
// counted, simply absent, on a screen that gives no sign it is showing less
// than everything.
//
// The port is therefore taken from the last colon, and only when what follows
// is digits. An address this cannot read keeps its whole text and shows an
// empty port: a row with one blank cell says what it knows, and a row that is
// not there says nothing at all.
func splitHostPort(s string) (host string, port string) {
	// tshark does not bracket, but Wireshark's own display filter syntax does,
	// so accept it in case a future release starts writing it that way.
	if strings.HasPrefix(s, "[") {
		if end := strings.LastIndex(s, "]"); end > 0 {
			host = s[1:end]
			if rest := s[end+1:]; strings.HasPrefix(rest, ":") && isPort(rest[1:]) {
				return host, rest[1:]
			}
			return host, ""
		}
	}

	i := strings.LastIndex(s, ":")
	if i < 0 || !isPort(s[i+1:]) {
		return s, ""
	}

	return s[:i], s[i+1:]
}

// isPort is whether s is what tshark writes for a port: digits, at least one.
// A hex group of an IPv6 address can be all digits too, which is why the split
// is anchored at the last colon rather than chosen by looking at the pieces.
func isPort(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
