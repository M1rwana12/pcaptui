// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"os"
	"strings"
)

//======================================================================

// tailLimit is how much of a log file is worth putting in a dialog. A log that
// has been accumulating for months is megabytes, and a dialog cannot page
// through it usefully; what a bug report needs is the end.
const tailLimit = 400

// tail returns the last n lines of s, saying how much was left out.
func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return s
	}

	dropped := len(lines) - n
	return fmt.Sprintf("... %d earlier lines are in the file itself\n\n%s",
		dropped, strings.Join(lines[dropped:], "\n"))
}

// fileBody reads the last n lines of a file for display, or says why it could
// not. A file that is missing or unreadable is itself the thing worth showing:
// that is what somebody with a broken install has.
func fileBody(path string, n int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Could not read it: %v", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return "The file is empty."
	}

	return tail(string(data), n)
}

// fileForDialog is fileBody headed by the path.
//
// The path is part of the answer, not decoration: someone told to "check the
// log" needs to know which file that is before they can attach it to a bug
// report.
func fileForDialog(what string, path string, n int) string {
	return fmt.Sprintf("%s: %s\n\n%s", what, path, fileBody(path, n))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
