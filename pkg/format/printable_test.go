// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// This runs on every column of every packet in the list - tshark writes
// unprintable bytes as "\x41" in PSML, and the list shows what it produces. A
// mistake here is a wrong Info column on a million rows, and nothing to say
// so.

func TestAHexCodeBecomesItsByte(t *testing.T) {
	assert.Equal(t, "A", string(TranslateHexCodes([]byte(`\x41`))))
	assert.Equal(t, "AB", string(TranslateHexCodes([]byte(`\x41\x42`))))
}

func TestOrdinaryTextIsLeftAlone(t *testing.T) {
	for _, s := range []string{
		"GET /index.html HTTP/1.1",
		"192.168.0.2",
		"",
		`C:\Users\senja`,
	} {
		assert.Equal(t, s, string(TranslateHexCodes([]byte(s))), s)
	}
}

func TestACodeInTheMiddleOfTextIsTranslated(t *testing.T) {
	assert.Equal(t, "before\tafter",
		string(TranslateHexCodes([]byte(`before\x09after`))))
}

// A trailing or truncated escape is not an escape. Swallowing what follows it
// would eat the rest of the cell, and panicking on it would take the program
// down for a packet whose Info column happens to end in a backslash.
func TestAnIncompleteEscapeIsNotSwallowed(t *testing.T) {
	for _, s := range []string{
		`ends in \x`,
		`ends in \x4`,
		`\xZZ is not hex`,
		`\`,
	} {
		assert.Equal(t, s, string(TranslateHexCodes([]byte(s))), s)
	}
}

// Upper and lower case both, because tshark's own output is not consistent
// about it across dissectors.
func TestBothCasesOfHexAreRead(t *testing.T) {
	assert.Equal(t, "\r\n", string(TranslateHexCodes([]byte(`\x0d\x0A`))))
}

// The one this is really for: a chat message ending in a carriage return and a
// newline, which is what "HTTP/1.1 200 OK\r\n" arrives as.
func TestALineEndingIsTranslated(t *testing.T) {
	got := string(TranslateHexCodes([]byte(`HTTP/1.1 200 OK\x0d\x0a`)))

	assert.Equal(t, "HTTP/1.1 200 OK\r\n", got)
}

func TestAZeroByteIsAByte(t *testing.T) {
	got := TranslateHexCodes([]byte(`a\x00b`))

	assert.Equal(t, []byte{'a', 0, 'b'}, got)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
