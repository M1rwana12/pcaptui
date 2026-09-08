// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Keys already spoken for elsewhere in the interface. appKeyPress runs only
// when nothing else consumed the keypress, so a clash here would not be a
// crash - it would be a key that works everywhere except the one view where
// the user needs it, which is worse, because it looks like a flake.
const reservedKeys = "qQ:?Zm'gGc/|\\+-<>zhjkl"

func TestNoAnalysisKeyIsAlreadyTaken(t *testing.T) {
	for _, v := range analysisViews() {
		assert.NotContains(t, reservedKeys, string(v.Key),
			"%q for %s is already bound elsewhere", v.Key, v.Name)
	}
}

func TestEveryAnalysisViewHasItsOwnKey(t *testing.T) {
	seen := map[rune]string{}
	for _, v := range analysisViews() {
		require.NotZero(t, v.Key, "%s has no key", v.Name)
		if other, dup := seen[v.Key]; dup {
			t.Errorf("%q opens both %s and %s", v.Key, other, v.Name)
		}
		seen[v.Key] = v.Name
	}
}

func TestEveryAnalysisViewCanBeOpened(t *testing.T) {
	for _, v := range analysisViews() {
		assert.NotNil(t, v.Open, "%s has no action", v.Name)
	}
}

// The keys are five of the program's fourteen or so global bindings; a stray
// upper-case or punctuation key here would collide with the vim commands.
func TestAnalysisKeysAreLowerCaseLetters(t *testing.T) {
	for _, v := range analysisViews() {
		assert.True(t, v.Key >= 'a' && v.Key <= 'z',
			"%q for %s is not a lower-case letter", v.Key, v.Name)
	}
}

func TestAnUnboundKeyOpensNothing(t *testing.T) {
	// nil app is safe precisely because nothing is opened.
	assert.False(t, analysisKeyPress('!', nil))
}

//======================================================================

// The help dialog is generated from the bindings so the two cannot disagree.
// It had already drifted once: it described :config and :logs as Unix-only
// after they stopped being so.
func TestTheHelpListsEveryKeyAndName(t *testing.T) {
	help := analysisKeyHelp()

	for _, v := range analysisViews() {
		assert.Contains(t, help, string(v.Key))
		assert.Contains(t, help, v.Name)
	}
	assert.Equal(t, len(analysisViews()), strings.Count(help, "\n"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
