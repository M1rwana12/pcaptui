// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package shark

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// A configuration in which every column is hidden used to stop the program
// before the interface appeared:
//
//	panic: All columns widgets were rendered Max, so there is no max height
//
// gowid's columns widget renders each child as Max, finds no maximum height
// among them, and gives up. Nothing in the interface prevents hiding the last
// column, so a user could write a configuration that made their own program
// refuse to start, with a panic that says nothing about columns being hidden.

func spec(hidden bool) PsmlColumnSpec {
	return PsmlColumnSpec{Field: PsmlField{Token: "%m"}, Name: "No.", Hidden: hidden}
}

func TestAllHiddenWhenEveryColumnIsHidden(t *testing.T) {
	assert.True(t, AllHidden([]PsmlColumnSpec{spec(true), spec(true)}))
}

func TestNotAllHiddenWhenOneRemains(t *testing.T) {
	assert.False(t, AllHidden([]PsmlColumnSpec{spec(true), spec(false), spec(true)}))
}

func TestNotAllHiddenWhenNoneAreHidden(t *testing.T) {
	assert.False(t, AllHidden([]PsmlColumnSpec{spec(false)}))
}

// An empty list is a different problem, handled separately by falling back to
// the defaults. Reporting it as "all hidden" would be true but useless, and
// would send the caller down the wrong branch.
func TestEmptyListIsNotAllHidden(t *testing.T) {
	assert.False(t, AllHidden(nil))
	assert.False(t, AllHidden([]PsmlColumnSpec{}))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
