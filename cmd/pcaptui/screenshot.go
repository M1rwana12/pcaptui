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

	// gowid refuses to build a line wider than 4096 cells, and tcell's
	// simulation screen allocates width*height cells up front - so a large
	// enough number is a panic rather than an error. Both are checked here so
	// the message names the flag instead of the library.
	const maxDimension = 4096

	w, err := strconv.Atoi(parts[0])
	if err != nil || w < 20 || w > maxDimension {
		return 0, 0, fmt.Errorf("screenshot width %q must be a number between 20 and %d", parts[0], maxDimension)
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil || h < 10 || h > maxDimension {
		return 0, 0, fmt.Errorf("screenshot height %q must be a number between 10 and %d", parts[1], maxDimension)
	}

	return w, h, nil
}

//======================================================================

const (
	shotPollEvery = 150 * time.Millisecond
	shotStableFor = 3 // consecutive identical polls
	shotGiveUp    = 25 * time.Second
)

// captureWhenSettled writes the screen once it stops changing, then quits.
//
// If keys are given, they are typed once the capture has finished loading, and
// the screen is left to settle a second time before being written - that is how
// the screenshot of a view reached by a command, rather than the one the
// program opens on, is taken.
func captureWhenSettled(app *gowid.App, screen tcell.SimulationScreen, prefix string, keys string) {
	loaded := settle(app, screen, "")

	if keys != "" {
		log.Infof("Screenshot: typing %q", keys)
		typeKeys(app, screen, keys)

		// Pass what was on screen before, so a view that takes a moment to
		// appear is waited for rather than photographed before it arrives.
		settle(app, screen, loaded)
	}

	rows, w, h := screenshot.Grab(screen)
	if err := writeShot(prefix, rows, w, h); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing screenshot: %v\n", err)
	}

	app.Quit()
}

// typeKeys feeds keystrokes to the interface one at a time.
//
// Explicit key events rather than InjectKeyBytes: the byte form has to be
// parsed back into events by tcell's terminal decoder, which a simulation
// screen is not driving. One key per poll interval, because the command line
// builds its completion list between keystrokes and a burst arrives before it
// is ready for it.
//
// "\n" in the string means Enter and "\t" means Tab, which is how a
// screenshot reaches a pane that is not the one the program starts focused on.
func typeKeys(app *gowid.App, screen tcell.SimulationScreen, keys string) {
	for _, r := range keys {
		switch r {
		case '\n', '\r':
			screen.InjectKey(tcell.KeyEnter, ' ', tcell.ModNone)
		case '\t':
			screen.InjectKey(tcell.KeyTab, '\t', tcell.ModNone)
		default:
			screen.InjectKey(tcell.KeyRune, r, tcell.ModNone)
		}

		app.Run(gowid.RunFunction(func(gowid.IApp) {}))
		time.Sleep(shotPollEvery)
	}
}

// settle polls until the screen has stopped changing, and returns what it
// finally showed.
//
// It waits for the picture to stop moving rather than sleeping for a fixed
// time because loading a capture takes as long as it takes: a fixed sleep is
// either too short on a slow machine, and produces a half-drawn image, or too
// long on every machine.
//
// differentFrom, when not empty, must be left behind first: the screen has to
// change away from it before stability counts. Without that, a command whose
// result takes a moment would be photographed before it appeared.
func settle(app *gowid.App, screen tcell.SimulationScreen, differentFrom string) string {
	deadline := time.Now().Add(shotGiveUp)

	var last string
	same := 0
	moved := differentFrom == ""

	for {
		// The main loop only paints in response to a render event, and with no
		// keyboard attached nothing produces one. Post an empty function to
		// generate one, so the screen we read is the screen the program means
		// to show.
		app.Run(gowid.RunFunction(func(gowid.IApp) {}))

		time.Sleep(shotPollEvery)

		rows, _, _ := screenshot.Grab(screen)
		now := screenshot.Text(rows)

		if !moved && now != differentFrom {
			moved = true
		}

		if now == last && strings.TrimSpace(now) != "" {
			same++
		} else {
			same = 0
		}
		last = now

		if moved && same >= shotStableFor {
			return now
		}

		if time.Now().After(deadline) {
			log.Warnf("Screenshot: screen still changing after %v, taking it anyway", shotGiveUp)
			return now
		}
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
