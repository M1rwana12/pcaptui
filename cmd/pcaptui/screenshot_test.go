// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func TestParseSizeAcceptsOrdinaryTerminals(t *testing.T) {
	for _, s := range []string{"80x24", "130x34", "200x60", "4096x4096"} {
		_, _, err := parseSize(s)
		assert.NoError(t, err, "should have accepted %q", s)
	}
}

// Both of these were panics before the bounds existed: gowid refuses to build
// a line wider than 4096 cells, and tcell's simulation screen allocates
// width*height cells up front, so a large enough pair is makeslice: len out of
// range. A flag should not be able to crash the program it belongs to.
func TestParseSizeRefusesWhatTheLibrariesCannotDraw(t *testing.T) {
	for _, s := range []string{"4097x34", "130x4097", "99999999x99999999"} {
		_, _, err := parseSize(s)
		assert.Error(t, err, "should have rejected %q", s)
	}
}

func TestParseSizeRefusesTerminalsTooSmallToShowAnything(t *testing.T) {
	for _, s := range []string{"19x24", "80x9", "1x1"} {
		_, _, err := parseSize(s)
		assert.Error(t, err, "should have rejected %q", s)
	}
}

func TestParseSizeRefusesWhatIsNotASize(t *testing.T) {
	for _, s := range []string{"", "130", "130x", "x34", "wide", "130x34x2", "-130x34"} {
		_, _, err := parseSize(s)
		assert.Error(t, err, "should have rejected %q", s)
	}
}

// The message has to name the flag's value, because the person reading it is
// looking at a command line, not at this source.
func TestParseSizeSaysWhatWasWrong(t *testing.T) {
	_, _, err := parseSize("4097x34")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "4097")
	assert.Contains(t, err.Error(), "4096")
}

//======================================================================

// A newline in the flag value has to mean Enter. The shell writes one when a
// script quotes a multi-line argument, which is how scripts/screenshots.sh
// asked for :expert - and gowid's vim parser drops it, so the screenshot came
// out showing the command typed into the command line but never run.
func TestARawNewlineMeansEnter(t *testing.T) {
	keys := parseKeys(":expert\n")

	require.Len(t, keys, 8)
	assert.Equal(t, tcell.KeyEnter, keys[7].Key())
}

func TestARawTabMeansTab(t *testing.T) {
	keys := parseKeys("\t")

	require.Len(t, keys, 1)
	assert.Equal(t, tcell.KeyTab, keys[0].Key())
}

func TestNamedKeysAreUnderstood(t *testing.T) {
	keys := parseKeys("<esc><down>q")

	require.Len(t, keys, 3)
	assert.Equal(t, tcell.KeyEscape, keys[0].Key())
	assert.Equal(t, tcell.KeyDown, keys[1].Key())
	assert.Equal(t, tcell.KeyRune, keys[2].Key())
	assert.Equal(t, 'q', keys[2].Rune())
}

func TestPrintableCharactersStandForThemselves(t *testing.T) {
	keys := parseKeys("http")

	require.Len(t, keys, 4)
	for i, r := range "http" {
		assert.Equal(t, r, keys[i].Rune())
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
