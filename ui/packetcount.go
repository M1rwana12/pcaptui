// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
)

//======================================================================

// updatePacketCount refreshes the title bar counter. It is called from
// updatePacketListWithData, which already runs on every tick of a load as well
// as at the end of one, so the number climbs while a capture is being read.
func updatePacketCount(psml iPsmlInfo, app gowid.IApp) {
	setPacketCount(psml, Loader != nil && Loader.PsmlLoader.IsLoading(), app)
}

// packetCountLoadFinished restates the count without "loading".
//
// The loader assigns state = NotLoading *after* running the AfterEnd handlers,
// so asking it at that moment gives the wrong answer - and nothing runs
// afterwards to correct it, which left the title saying "loading" for the rest
// of the session. AfterEnd for the PSML load is itself the signal that the
// load is over, so it is used rather than the flag.
func packetCountLoadFinished(psml iPsmlInfo, app gowid.IApp) {
	setPacketCount(psml, false, app)
}

func setPacketCount(psml iPsmlInfo, loading bool, app gowid.IApp) {
	if packetCount == nil || Loader == nil {
		return
	}

	packetCount.SetText(packetCountText(
		len(psml.PsmlData()),
		Loader.DisplayFilter() != "",
		loading,
	), app)
}

//======================================================================

// packetCountText is what the title bar says about how many packets there are.
//
// The count was missing from the interface entirely: nothing anywhere said
// whether a capture held seven packets or seven million, and nothing said that
// a display filter was the reason the list looked short. Wireshark's status bar
// says both, and it is the most-read part of its window.
//
// filtered and loading are both stated rather than left to be inferred, for the
// same reason the packet list marks a value it had to cut short: a number that
// is not the whole story has to say so.
func packetCountText(n int, filtered bool, loading bool) string {
	var parts []string

	switch {
	case n == 0 && loading:
		parts = append(parts, "loading")
	case n == 0 && filtered:
		return "no packets match"
	case n == 0:
		return "no packets"
	case n == 1:
		parts = append(parts, "1 packet")
	default:
		parts = append(parts, fmt.Sprintf("%s packets", groupDigits(n)))
	}

	if filtered {
		parts = append(parts, "filtered")
	}
	if loading && n != 0 {
		parts = append(parts, "loading")
	}

	return strings.Join(parts, " · ")
}

// groupDigits puts a comma every three digits. A million-packet capture is a
// stated goal of this program, and 1048576 is not a number anyone reads at a
// glance.
func groupDigits(n int) string {
	s := fmt.Sprintf("%d", n)

	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}

	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}

	return sign + b.String()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
