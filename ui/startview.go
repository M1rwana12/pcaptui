// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui/configs/profiles"
	"github.com/m1rwana12/pcaptui/pkg/stats"
)

//======================================================================

// StartViewOverview and StartViewPackets are the values main.start-view takes.
const (
	StartViewOverview = "overview"
	StartViewPackets  = "packets"
)

// overviewShown stops the summary reappearing on every load.
//
// A new display filter reloads the capture and reaches the same AfterEnd, and
// a summary that reopened each time somebody filtered would be an obstacle
// rather than an introduction.
var overviewShown bool

// ResetStartView lets the summary be shown again for the next capture.
func ResetStartView() {
	overviewShown = false
}

// maybeShowOverview opens the summary over the packet list once, when a
// capture has finished loading.
//
// This is the plan's triage screen, and it is a dialog over the packet list
// rather than a different view. A capture opened at packet number one answers
// none of the questions the person opening it has, and the evidence that this
// matters is babyshark: 566 stars in four days for not opening on packet one.
// But the packet list is what somebody who already knows this program expects
// to see, so the summary sits on top of it and one keypress dismisses it -
// rather than replacing the view and having to be navigated back out of.
//
// main.start-view = packets turns it off.
func maybeShowOverview(app gowid.IApp) {
	if overviewShown {
		return
	}

	// Not while drawing screenshots. The picture has to be the same every run,
	// and a dialog that opens by itself would put one in every image and
	// change what the gate compares.
	if ScreenOverride != nil {
		return
	}

	if profiles.ConfString("main.start-view", StartViewOverview) != StartViewOverview {
		return
	}

	// Only for a capture that has an end. A live capture's summary would be a
	// snapshot of the first second, obsolete before it was read.
	if Loader == nil || Loader.PcapPdml == "" || Loader.PsmlLoader.ReadingFromFifo() {
		return
	}

	overviewShown = true
	startStats(stats.Overview, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
