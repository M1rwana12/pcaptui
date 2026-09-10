// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import "fmt"

//======================================================================

// displayFilterFor resolves the display filter from the two ways of giving
// one: the -Y flag, and the positional argument when the source is a file.
//
// It exists as one function so that there is one resolved filter to check.
// There used to be two: the flag was validated and the positional argument was
// not, so `pcaptui -Y 'tcp.prot == 80' -r capture.pcap` stopped and named the
// field that does not exist, while `pcaptui -r capture.pcap 'tcp.prot == 80'`
// - the form this program's own README uses - started, drew a title with no
// filename, an empty filter box and three empty panes, and exited zero. A
// typo looked exactly like a capture with nothing in it.
//
// The positional argument is a display filter only when reading a file. On an
// interface it is a capture filter, and the caller has already taken it.
func displayFilterFor(flagFilter string, argsFilter string, readingFile bool) (string, error) {
	if !readingFile || argsFilter == "" {
		return flagFilter, nil
	}

	// Both given. Preferring one silently would apply a filter the user did
	// not mean and hide the one they did.
	if flagFilter != "" {
		return "", fmt.Errorf(
			"Two display filters provided - '%s' and '%s' - please supply one only.",
			flagFilter, argsFilter)
	}

	return argsFilter, nil
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
