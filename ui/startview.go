// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/configs/profiles"
	"github.com/m1rwana12/pcaptui/pkg/capinfo"
	"github.com/m1rwana12/pcaptui/pkg/stats"
	log "github.com/sirupsen/logrus"
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

// overviewAutoLimit is the size above which the statistics are not started on
// the user's behalf.
//
// Measured with scripts/bench-overview.sh, on the three platforms this is
// tested on. Two things came out of it that guesswork had wrong:
//
//	            21 MB     85 MB
//	Linux        1.9 s    19.4 s
//	macOS        5.0 s    47.1 s
//	Windows      5.8 s        -
//
// The pass does not run at a constant rate - it slows down as the file grows,
// so a rate taken from a small capture understates a large one - and the
// slowest of the three is not the machine this is written on. A capture the
// size the README says this program handles is therefore minutes of a spinner
// that nobody asked for.
//
// capinfos, on the other hand, stays cheap at every size: 0.52 s for that same
// 85 MB on the slowest platform, against 47. That is the split this threshold
// makes. The file's own facts - when it was captured, how long it covers, how
// many packets - are what answer "what am I looking at", and they are still
// fetched and shown. The pass over every packet is offered instead of started.
//
// Ten megabytes, because that is about two seconds on the slowest platform
// measured, and two seconds is a screen appearing rather than a wait.
const overviewAutoLimit = 10 * 1024 * 1024

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

	// A file that cannot be stat'ed is treated as small: it is about to be read
	// anyway, and refusing to summarise it would be a worse guess.
	if info, err := os.Stat(Loader.PcapPdml); err == nil && !overviewAutoStart(info.Size()) {
		showOverviewFacts(app)
		return
	}

	startStats(stats.Overview, app)
}

// overviewAutoStart says whether a capture of this size gets its statistics
// started for it, rather than offered.
func overviewAutoStart(size int64) bool {
	return size <= overviewAutoLimit
}

// overviewFactsView is the dialog for a capture too large to summarise
// unasked: what the file says about itself, and what running the rest costs.
//
// There is no traffic histogram in it, because that row comes from the pass
// this is avoiding.
func overviewFactsView(info capinfo.Info, filter string) statsView {
	v := statsView{Heading: statsHeading(stats.Overview.Name, filter, false)}

	v.Rows = fileFactsLines(info, nil)
	v.Rows = append(v.Rows,
		statsLine{},
		statsLine{Text: fmt.Sprintf(
			"The statistics are one pass over all %s - press o to run them.",
			strings.TrimSpace(info.Size))})

	return v
}

// showOverviewFacts opens what the file says about itself, and offers the rest
// rather than starting it.
//
// Pressing `o` still runs the whole thing, at any size: that is somebody
// asking for it, which is a different matter from a program deciding to spend
// a minute of their time.
func showOverviewFacts(app gowid.IApp) {
	pcapfile := Loader.PcapPdml
	filter := Loader.DisplayFilter()

	pcaptui.TrackedGo(func() {
		info, err := capinfo.Read(pcapfile)
		if err != nil {
			// Nothing to show and nothing to offer that the user cannot
			// already do with `o`, so this stays in the log.
			log.Warnf("Could not read the file's own properties: %v", err)
			return
		}

		app.Run(gowid.RunFunction(func(app gowid.IApp) {
			openStatsDialog(overviewFactsView(info, filter), app)
		}))
	}, Goroutinewg)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
