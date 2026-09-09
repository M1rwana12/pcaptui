// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcaptui

import (
	"regexp"
	"strconv"
	"strings"
)

//======================================================================

// The numbers tshark prints are formatted for a person, and the formatting
// depends on the machine. Three differences have shown up in this project so
// far, each found by one CI runner disagreeing with another on the same
// commit:
//
//   - digit grouping: "7,748" on macOS where this machine wrote "7748",
//     and only in some of the tables;
//   - the decimal separator: this machine writes "39,5713" and "0,000000000"
//     where the runners write "39.5713";
//   - which of the two a lone comma is, which no amount of looking at one
//     number can settle.
//
// So both readers below take a side deliberately and say which.

// ParseCount reads a whole number that tshark may have grouped for
// readability.
//
// Groups are stripped only when they are groups of exactly three. Stripping
// unconditionally would read "1,5" - one and a half, on a machine like this
// one - as fifteen, and a wrong number is worse than no number.
//
// A space separator, which some locales use, would have split the number into
// two fields before it ever reached here; if that appears it needs a different
// fix, and it announces itself as an empty table rather than as a wrong count.
func ParseCount(s string) (int, bool) {
	if groupedDigits.MatchString(s) {
		s = strings.NewReplacer(",", "", "'", "").Replace(s)
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// ParseDecimal reads a number tshark printed with a fractional part - a
// duration, a relative start time, a rate.
//
// The decimal separator is whichever of "." and "," comes last, because a
// number carrying both has its groups first. A lone separator is a decimal
// point unless it looks exactly like one group of three, which is the same
// call ParseCount makes and for the same reason: "1,234" is far more often
// twelve hundred than one and a fifth.
//
// Times and durations do not hit that ambiguity - tshark prints nine decimals
// for a start and four for a duration, neither of which is three - so the
// columns this was written for are read correctly either way.
func ParseDecimal(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	dot, comma := strings.LastIndex(s, "."), strings.LastIndex(s, ",")

	switch {
	case dot >= 0 && comma >= 0:
		// Both present: the later one is the decimal separator and the other
		// is grouping.
		if comma > dot {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.Replace(s, ",", ".", 1)
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}

	case comma >= 0:
		if groupedDigits.MatchString(s) {
			s = strings.ReplaceAll(s, ",", "")
		} else {
			s = strings.Replace(s, ",", ".", 1)
		}

	case dot >= 0:
		if groupedDigits.MatchString(strings.ReplaceAll(s, ".", ",")) {
			s = strings.ReplaceAll(s, ".", "")
		}
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

var groupedDigits = regexp.MustCompile(`^\d{1,3}([,']\d{3})+$`)

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
