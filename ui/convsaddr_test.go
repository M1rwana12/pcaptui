// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestAnIPv4EndpointSplitsAtItsOnlyColon(t *testing.T) {
	host, port := splitHostPort("192.168.0.2:1550")

	assert.Equal(t, "192.168.0.2", host)
	assert.Equal(t, "1550", port)
}

// This is what the whole file is about: tshark writes IPv6 unbracketed, so the
// address has colons of its own and the port is after the last one. Splitting
// on every colon and demanding two pieces dropped every one of these rows.
func TestAnIPv6EndpointSplitsAtItsLastColon(t *testing.T) {
	host, port := splitHostPort("2001:db8::1:50000")

	assert.Equal(t, "2001:db8::1", host)
	assert.Equal(t, "50000", port)
}

func TestAFullyWrittenIPv6AddressSplitsToo(t *testing.T) {
	host, port := splitHostPort("2001:0db8:0000:0000:0000:0000:0000:0001:443")

	assert.Equal(t, "2001:0db8:0000:0000:0000:0000:0000:0001", host)
	assert.Equal(t, "443", port)
}

// A hex group can be all digits, so the split cannot be chosen by looking at
// which piece looks like a number - only the last colon is the port.
func TestAnAddressEndingInDigitsIsNotMistakenForAPort(t *testing.T) {
	host, port := splitHostPort("fe80::1234:5678:9012:80")

	assert.Equal(t, "fe80::1234:5678:9012", host)
	assert.Equal(t, "80", port)
}

// A row this cannot read keeps its text and shows an empty port. A row with a
// blank cell says what it knows; a row that was dropped says nothing at all,
// which is how the IPv6 rows disappeared in the first place.
func TestAnAddressWithNoPortKeepsItsWholeText(t *testing.T) {
	for _, s := range []string{
		"192.168.0.2",
		"00:a0:cc:3b:bf:fa",
		"",
	} {
		host, port := splitHostPort(s)

		assert.Equal(t, s, host, "the address must survive whole")
		assert.Equal(t, "", port)
	}
}

// Not what this tshark writes, but Wireshark's own filter syntax brackets IPv6
// and a future release may follow it.
func TestABracketedAddressLosesItsBrackets(t *testing.T) {
	host, port := splitHostPort("[2001:db8::1]:443")

	assert.Equal(t, "2001:db8::1", host)
	assert.Equal(t, "443", port)
}

func TestABracketedAddressWithNoPortIsStillRead(t *testing.T) {
	host, port := splitHostPort("[2001:db8::1]")

	assert.Equal(t, "2001:db8::1", host)
	assert.Equal(t, "", port)
}

// A bare IPv6 address ending in digits cannot be told from an address with a
// port - "2001:db8::1" is either the address itself or "2001:db8:" on port 1,
// and nothing in the text says which. This is only ever handed the endpoints
// of a TCP or UDP conversation, where tshark always writes a port; the other
// tabs pass the address through untouched. Recorded so that the day someone
// reuses this function, the limit is written down rather than discovered.
func TestABareIPv6AddressIsAmbiguousAndIsSplitAnyway(t *testing.T) {
	host, port := splitHostPort("2001:db8::1")

	assert.Equal(t, "2001:db8:", host)
	assert.Equal(t, "1", port)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
