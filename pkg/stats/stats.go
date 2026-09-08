// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package stats exposes tshark's -z statistics as panels pcaptui can show.
//
// tshark carries a large collection of analyses behind -z; pcaptui used only
// conv,* for the conversations view and follow,* for stream reassembly. The
// two here answer the questions asked most often of a capture: what is in it,
// and what is wrong with it.
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
)

// All is every statistic pcaptui offers, in the order they are presented.
var All = []Stat{Expert, ProtoHierarchy}

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
	filter := strings.TrimSpace(displayFilter)
	if filter == "" {
		return s.zbase
	}
	return s.zbase + "," + filter
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
