// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package stats exposes tshark's -z statistics as panels pcaptui can show.
//
// tshark carries a large collection of analyses behind -z; pcaptui used only
// conv,* for the conversations view and follow,* for stream reassembly. The
// ones here answer the questions asked most often of a capture: what is in it,
// what is wrong with it, who is on the wire, and how its HTTP and DNS turned
// out.
package stats

import "strings"

//======================================================================

// Stat describes one tshark statistic and how to ask for it.
type Stat struct {
	// Name titles the dialog.
	Name string
	// Command is the minibuffer command that opens it.
	Command string
	// Key is the single keypress that opens it, from the main view and from
	// the Analysis menu alike. One letter per view is how k9s and lazygit are
	// navigated, and a view reachable only by typing its name is a view most
	// people never find.
	Key rune
	// Summary is the one-line description shown in the menu and help.
	Summary string
	// zbase is the -z argument without any display filter.
	zbase string
	// extraZ are further -z arguments asked for in the same pass. tshark
	// accepts several, and one pass over a capture beats three.
	extraZ []string
	// ignoresFilter marks a tap that takes a display filter and does not
	// apply it. Measured on credentials: `-z credentials,frame.number == 999`
	// is accepted, exits zero, and reports the login in packet 1 anyway, while
	// `-z expert` with the same filter correctly reports nothing. Passing the
	// filter would make the dialog's heading say the result was narrowed when
	// it was not.
	ignoresFilter bool
}

var (
	// Expert is tshark's expert information: the warnings the dissectors
	// raise - retransmissions, malformed packets, protocol violations.
	// Wireshark shows it under Analyze > Expert Information.
	Expert = Stat{
		Name:    "Expert Information",
		Command: "expert",
		Key:     'e',
		Summary: "Problems the dissectors found in this capture",
		zbase:   "expert",
	}

	// ProtoHierarchy is the protocol hierarchy: every protocol present, by
	// packet and byte count. Wireshark shows it under Statistics > Protocol
	// Hierarchy.
	ProtoHierarchy = Stat{
		Name:    "Protocol Hierarchy",
		Command: "hierarchy",
		Key:     'y',
		Summary: "Every protocol in this capture, by packets and bytes",
		zbase:   "io,phs",
	}

	// Endpoints is who is on the wire and how much each of them sent and
	// received. Wireshark shows it under Statistics > Endpoints.
	//
	// Addresses rather than every endpoint type: an address is the thing
	// people look for by name, and it is the one a display filter can be built
	// from without knowing which layer to ask about.
	//
	// Both families, in one pass. Asking only for IPv4 meant a capture of IPv6
	// traffic answered "Nothing to report" to the question "who is on the
	// wire" while the protocol hierarchy in the same dialog listed ipv6 six
	// lines above. Measured on a 22 MB capture: adding the second table cost
	// nothing outside the noise between runs.
	Endpoints = Stat{
		Name:    "Endpoints",
		Command: "endpoints",
		Key:     't',
		Summary: "Who is on the wire, by packets and bytes",
		zbase:   "endpoints,ip",
		extraZ:  []string{"endpoints,ipv6"},
	}

	// HTTP counts the responses by status: how many succeeded, how many were
	// not found, how many the server failed on. Wireshark shows it under
	// Statistics > HTTP > Packet Counter.
	//
	// Key `w` for web: `h` is taken by the vim-style movement keys.
	HTTP = Stat{
		Name:    "HTTP",
		Command: "http",
		Key:     'w',
		Summary: "How the HTTP responses in this capture turned out",
		zbase:   "http,tree",
	}

	// DNS is what was asked for and how the answers turned out: how many
	// lookups failed, with which code, and how long the server took.
	// Wireshark shows it under Statistics > DNS.
	DNS = Stat{
		Name:    "DNS",
		Command: "dns",
		Key:     'd',
		Summary: "What was asked of DNS and how the answers turned out",
		zbase:   "dns,tree",
	}

	// Credentials is the logins tshark could read in the clear: HTTP basic
	// auth, FTP, POP, IMAP, SMTP and telnet. Wireshark shows it under
	// Tools > Credentials.
	//
	// The one statistic here that names a packet rather than counting
	// occurrences, so its rows filter exactly: `frame.number == 1`.
	//
	// It ignores a display filter, measured rather than assumed - see
	// ignoresFilter - so it is never given one.
	//
	// Key `a` for auth: `c` is taken by copy-mode.
	Credentials = Stat{
		Name:          "Credentials",
		Command:       "credentials",
		Key:           'a',
		Summary:       "Logins this capture carries in the clear",
		zbase:         "credentials",
		ignoresFilter: true,
	}

	// Overview is the three of them at once: what is in the capture, what is
	// wrong with it, and who is on the wire.
	//
	// It is the question somebody actually has when handed a capture, and
	// answering it needed knowing three commands. tshark accepts several -z
	// arguments in one invocation, so this is one pass over the file rather
	// than three.
	Overview = Stat{
		Name:    "Overview",
		Command: "overview",
		Key:     'o',
		Summary: "What is in this capture, what is wrong with it, who is on the wire",
		zbase:   "io,phs",
		extraZ:  []string{"expert", "endpoints,ip", "endpoints,ipv6"},
	}
)

// All is every statistic pcaptui offers, in the order they are presented.
var All = []Stat{Overview, Expert, ProtoHierarchy, Endpoints, Credentials, HTTP, DNS}

// Lookup finds a statistic by its minibuffer command.
func Lookup(command string) (Stat, bool) {
	for _, s := range All {
		if s.Command == command {
			return s, true
		}
	}
	return Stat{}, false
}

// ZArg builds the -z argument, narrowing the statistic to displayFilter when
// one is set.
//
// tshark reads everything after the statistic's own arguments as the filter,
// so a filter containing commas - "tcp.port in {80,443}" is ordinary Wireshark
// syntax - is passed through whole and needs no escaping.
func (s Stat) ZArg(displayFilter string) string {
	return withFilter(s.zbase, s.filterFor(displayFilter))
}

// IgnoresFilter is whether this statistic is computed over the whole capture
// whatever the display filter says, so that the caller can say so rather than
// claiming a narrowing that did not happen.
func (s Stat) IgnoresFilter() bool {
	return s.ignoresFilter
}

// filterFor is the filter to pass to tshark: none, for a tap that would take
// it and ignore it.
func (s Stat) filterFor(displayFilter string) string {
	if s.ignoresFilter {
		return ""
	}
	return displayFilter
}

// ZArgs is every -z argument this statistic needs, narrowed to displayFilter.
// Most have one; Overview asks for three in a single pass.
func (s Stat) ZArgs(displayFilter string) []string {
	filter := s.filterFor(displayFilter)

	res := []string{withFilter(s.zbase, filter)}
	for _, z := range s.extraZ {
		res = append(res, withFilter(z, filter))
	}
	return res
}

func withFilter(zbase string, displayFilter string) string {
	filter := strings.TrimSpace(displayFilter)
	if filter == "" {
		return zbase
	}
	return zbase + "," + filter
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
