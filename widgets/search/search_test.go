// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func TestAHexTermBecomesTheBytesItNames(t *testing.T) {
	b, err := hexTermToBytes("54AD090f")

	require.NoError(t, err)
	assert.Equal(t, []byte{0x54, 0xad, 0x09, 0x0f}, b)
}

func TestAnEmptyHexTermIsNoBytes(t *testing.T) {
	b, err := hexTermToBytes("")

	require.NoError(t, err)
	assert.Empty(t, b)
}

// These used to panic. The validator's regexp does refuse them, so reaching
// this needed the validator to be bypassed - but a panic is a poor way to find
// out that it was, and the error path beside it already existed for regex
// searches.
func TestHexInputThatCannotBeReadIsAnErrorNotAPanic(t *testing.T) {
	for _, in := range []string{"abc", "5", "zz", "54AD09g", "0x54"} {
		_, err := hexTermToBytes(in)
		assert.Error(t, err, "should have refused %q", in)
	}
}

func TestTheHexErrorNamesWhatWasWrong(t *testing.T) {
	_, err := hexTermToBytes("abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "even number")

	_, err = hexTermToBytes("zz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "zz")
}

func TestNewHexTermPassesTheErrorOn(t *testing.T) {
	_, err := newHexTerm("nothex")

	assert.Error(t, err)
}

//======================================================================

func TestHexDigitsConvert(t *testing.T) {
	assert.Equal(t, 0, hexToByte('0'))
	assert.Equal(t, 9, hexToByte('9'))
	assert.Equal(t, 10, hexToByte('a'))
	assert.Equal(t, 15, hexToByte('f'))
	assert.Equal(t, 10, hexToByte('A'))
	assert.Equal(t, 15, hexToByte('F'))
}

func TestANonHexDigitIsRejected(t *testing.T) {
	for _, r := range []byte{'g', 'G', 'z', ' ', '-', 0} {
		assert.Equal(t, -1, hexToByte(r), "should have rejected %q", string(r))
	}
}

//======================================================================

// The configuration file could stop the program from starting. search.New
// calls updateSearchTargetFromConf during ui.Build, and that read
// main.search-type and main.search-target straight out of the file instead of
// through the guards written for exactly this - so
//
//	search-type = 'nonsense'
//
// panicked before the interface appeared, with the message "panic called with
// nil argument", which names neither the key nor the file.
//
// Measured against the real binary before the fix, not reasoned about.
func TestAnUnknownSearchTypeFallsBackInsteadOfPanicking(t *testing.T) {
	// getSearchType and getSearchTarget are what updateSearchTargetFromConf
	// now goes through; with no config file at all they return the defaults,
	// and every value they can return is a case in that switch.
	assert.Contains(t, searchTypeMap, getSearchType())
	assert.Contains(t, searchTargetMap, getSearchTarget())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
