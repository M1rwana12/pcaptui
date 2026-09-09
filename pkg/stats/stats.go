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
	// IPv4 rather than every endpoint type: an address is the thing people
	// look for by name, and it is the one a display filter can be built from
	// without knowing which layer to ask about.
	Endpoints = Stat{
		Name:    "Endpoints",
		Command: "endpoints",
		Key:     't',
		Summary: "Who is on the wire, by packets and bytes",
		zbase:   "endpoints,ip",
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
		extraZ:  []string{"expert", "endpoints,ip"},
	}
)

// All is every statistic pcaptui offers, in the order they are presented.
var All = []Stat{Overview, Expert, ProtoHierarchy, Endpoints, HTTP, DNS}

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
	return withFilter(s.zbase, displayFilter)
}

// ZArgs is every -z argument this statistic needs, narrowed to displayFilter.
// Most have one; Overview asks for three in a single pass.
func (s Stat) ZArgs(displayFilter string) []string {
	res := []string{withFilter(s.zbase, displayFilter)}
	for _, z := range s.extraZ {
		res = append(res, withFilter(z, displayFilter))
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
