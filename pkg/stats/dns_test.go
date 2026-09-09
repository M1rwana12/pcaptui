// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestTheDNSTotalFiltersOnDNS(t *testing.T) {
	assert.Equal(t, "dns", DNSFilter("", "Total Packets"))
}

// tshark names the section after one of the two things in it, so "Response"
// appears at two depths and means two different things.
func TestTheQueryResponseSplitFiltersOnTheFlag(t *testing.T) {
	assert.Equal(t, "dns.flags.response == 0", DNSFilter("Response", "Query"))
	assert.Equal(t, "dns.flags.response == 1", DNSFilter("Response", "Response"))
	assert.Equal(t, "", DNSFilter("", "Response"), "the section itself is every packet twice over")
}

func TestAnRcodeRowFiltersOnThatCode(t *testing.T) {
	assert.Equal(t, "dns.flags.rcode == 0", DNSFilter("rcode", "No error"))
	assert.Equal(t, "dns.flags.rcode == 2", DNSFilter("rcode", "Server failure"))
	assert.Equal(t, "dns.flags.rcode == 3", DNSFilter("rcode", "No such name"))
	assert.Equal(t, "dns.flags.rcode == 5", DNSFilter("rcode", "Refused"))
}

func TestAQueryTypeRowFiltersOnThatType(t *testing.T) {
	assert.Equal(t, "dns.qry.type == 1", DNSFilter("Query Type", "A"))
	assert.Equal(t, "dns.qry.type == 28", DNSFilter("Query Type", "AAAA"))
	assert.Equal(t, "dns.qry.type == 15", DNSFilter("Query Type", "MX"))
}

// The same name in another section is another thing entirely: "A" under Answer
// Type is what came back, not what was asked for, and that is a different
// field. Rather than a second mapping to keep right, it is not offered.
func TestTheSameNameInAnotherSectionIsNotTheSameRow(t *testing.T) {
	assert.Equal(t, "dns.qry.type == 1", DNSFilter("Query Type", "A"))
	assert.Equal(t, "", DNSFilter("Answer Type", "A"))
}

// Most of the table is headings and averages. No display filter selects "the
// packets that contributed to this average", so those rows get none.
func TestARowThatNoFilterCanSelectGetsNone(t *testing.T) {
	assert.Equal(t, "", DNSFilter("Query Stats", "Qname Len"))
	assert.Equal(t, "", DNSFilter("Service Stats", "request-response time (msec)"))
	assert.Equal(t, "", DNSFilter("Response Stats", "no. of answers"))
	assert.Equal(t, "", DNSFilter("", "Payload size"))
}

// A name this does not know is not guessed at. Wireshark can add response
// codes and record types, and a wrong filter is worse than none.
func TestAnUnknownNameGetsNoFilter(t *testing.T) {
	assert.Equal(t, "", DNSFilter("rcode", "Some future code"))
	assert.Equal(t, "", DNSFilter("Query Type", "HTTPS"))
	assert.Equal(t, "", DNSFilter("", ""))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
