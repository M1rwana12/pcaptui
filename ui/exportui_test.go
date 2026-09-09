// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// tshark exits successfully whether it wrote a hundred files or none, so the
// two outcomes have to be told apart here or they look the same.
func TestNoObjectsIsSaidInWords(t *testing.T) {
	got := exportResult("tftp", "/tmp/somewhere", nil)

	assert.Equal(t, "No tftp objects in this capture.", got)
	assert.NotContains(t, got, "/tmp/somewhere",
		"a directory nothing was written to is not where to look")
}

func TestTheMessageSaysWhereTheFilesWent(t *testing.T) {
	got := exportResult("http", "/tmp/objects", []string{"a.html", "b.css"})

	assert.Contains(t, got, "2 http object(s)")
	assert.Contains(t, got, "/tmp/objects")
	assert.Contains(t, got, "a.html")
	assert.Contains(t, got, "b.css")
}

// A capture can carry hundreds. The dialog is for telling you it worked and
// where to look, not for listing them all.
func TestALongListIsCutShortAndSaysSo(t *testing.T) {
	var many []string
	for i := 0; i < 25; i++ {
		many = append(many, "object")
	}

	got := exportResult("http", "/tmp/objects", many)

	assert.Contains(t, got, "25 http object(s)")
	assert.Contains(t, got, "and 15 more")
	assert.Equal(t, 10, strings.Count(got, "object\n"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
