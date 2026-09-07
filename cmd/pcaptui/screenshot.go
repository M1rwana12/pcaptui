// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell/v2"
	"github.com/m1rwana12/pcaptui/internal/screenshot"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// newShotScreen builds the simulation screen the interface is drawn into for
// --screenshot. size is "WxH" in character cells.
func newShotScreen(size string) (tcell.SimulationScreen, error) {
	w, h, err := parseSize(size)
	if err != nil {
		return nil, err
	}

	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		return nil, fmt.Errorf("could not start the simulation screen: %v", err)
	}
	s.SetSize(w, h)

	return s, nil
}

func parseSize(size string) (int, int, error) {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(size)), "x", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("screenshot size %q is not of the form WxH", size)
	}

	w, err := strconv.Atoi(parts[0])
	if err != nil || w < 20 {
		return 0, 0, fmt.Errorf("screenshot width %q must be a number of at least 20", parts[0])
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil || h < 10 {
		return 0, 0, fmt.Errorf("screenshot height %q must be a number of at least 10", parts[1])
	}

	return w, h, nil
}

//======================================================================

// How long the screen has to stop changing before it is considered finished,
// and how long to wait for that before giving up and taking what is there.
const (
	shotPollEvery = 150 * time.Millisecond
	shotStableFor = 3 // consecutive identical polls
	shotGiveUp    = 25 * time.Second
)

// captureWhenSettled writes the screen once it stops changing, then quits.
//
// It waits for the picture to settle rather than sleeping for a fixed time
// because loading a capture takes as long as it takes: a fixed sleep is either
// too short on a slow machine, and produces a half-drawn image, or too long on
// every machine. Settling is also what makes the result reproducible enough to
// commit and compare against.
func captureWhenSettled(app *gowid.App, screen tcell.SimulationScreen, prefix string) {
	deadline := time.Now().Add(shotGiveUp)

	var last string
	same := 0

	for {
		// The main loop only paints in response to a render event, and with no
		// keyboard attached nothing produces one. Post an empty function to
		// generate one, so the screen we read is the screen the program means
		// to show.
		app.Run(gowid.RunFunction(func(gowid.IApp) {}))

		time.Sleep(shotPollEvery)

		rows, w, h := screenshot.Grab(screen)
		now := screenshot.Text(rows)

		if now == last && strings.TrimSpace(now) != "" {
			same++
		} else {
			same = 0
		}
		last = now

		settled := same >= shotStableFor
		expired := time.Now().After(deadline)

		if !settled && !expired {
			continue
		}

		if expired && !settled {
			log.Warnf("Screenshot: screen still changing after %v, taking it anyway", shotGiveUp)
		}

		if err := writeShot(prefix, rows, w, h); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing screenshot: %v\n", err)
		}

		app.Quit()
		return
	}
}

func writeShot(prefix string, rows [][]screenshot.Cell, w, h int) error {
	if dir := filepath.Dir(prefix); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	txt := prefix + ".txt"
	if err := os.WriteFile(txt, []byte(screenshot.Text(rows)), 0o644); err != nil {
		return err
	}

	svg := prefix + ".svg"
	if err := os.WriteFile(svg, []byte(screenshot.SVG(rows, w, h)), 0o644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Wrote %s and %s (%dx%d)\n", txt, svg, w, h)
	return nil
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
