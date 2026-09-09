// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
)

//======================================================================

// DNSFilter is the display filter that shows the packets a row counted, or ""
// when there is no honest one.
//
// A row's name means nothing on its own here: "A" under Query Type is what was
// asked for and "A" under Answer Type is what came back, and half the table is
// headings like "Label Stats" that count nothing a filter could select. So the
// section the row sits in decides, and only four sections offer one:
//
//	Total Packets      every DNS packet
//	Response           the query/response split
//	rcode              how the responses turned out
//	Query Type         what was asked for
//
// The rest are shown and not offered. They are averages and lengths - "Qname
// Len", "no. of answers" - and no display filter selects "the packets that
// contributed to this average".
//
// One of the four does not agree with its own count, and cannot be made to.
// tshark counts the rcode bits of every DNS header, and a query carries a zero
// there, so a capture of two lookups reports "No error" three times: both
// queries and the one good answer. The dissector puts dns.flags.rcode on
// responses only, so the filter finds one packet where the row says three.
// The filter is still the useful half - "show me the answers that failed" is
// the question being asked - so it is offered, and the difference is Wireshark
// counting something no filter can name.
func DNSFilter(parent string, name string) string {
	switch parent {
	case "":
		if name == "Total Packets" {
			return "dns"
		}
	case "Response":
		// tshark names the section after one of the two things in it, so the
		// section and its first row share a name.
		switch name {
		case "Query":
			return "dns.flags.response == 0"
		case "Response":
			return "dns.flags.response == 1"
		}
	case "rcode":
		if code, ok := dnsRcodes[name]; ok {
			return fmt.Sprintf("dns.flags.rcode == %d", code)
		}
	case "Query Type":
		if t, ok := dnsQueryTypes[name]; ok {
			return fmt.Sprintf("dns.qry.type == %d", t)
		}
	}

	return ""
}

// dnsRcodes are the response codes by the names Wireshark prints. Only the six
// from RFC 1035, which have been spelled this way for as long as Wireshark has
// had a DNS dissector; a name not listed gets no filter rather than a guess.
var dnsRcodes = map[string]int{
	"No error":        0,
	"Format error":    1,
	"Server failure":  2,
	"No such name":    3,
	"Not implemented": 4,
	"Refused":         5,
}

// dnsQueryTypes are the record types worth clicking on, by the names Wireshark
// prints. Not every type it knows - the long tail is not what anyone is
// looking for in a capture, and each entry is a name that could be spelled
// differently in some future release.
var dnsQueryTypes = map[string]int{
	"A":     1,
	"NS":    2,
	"CNAME": 5,
	"SOA":   6,
	"PTR":   12,
	"MX":    15,
	"TXT":   16,
	"AAAA":  28,
	"SRV":   33,
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
