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
	"github.com/gcla/gowid/vim"
	"github.com/gdamore/tcell/v2"
	"github.com/m1rwana12/pcaptui/internal/screenshot"
	"github.com/m1rwana12/pcaptui/ui"
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

// stillLoading reports whether either half of a capture load is still running.
//
// A screen that has stopped changing is not the same as a capture that has
// finished loading. The packet list arrives first, from PSML, and the protocol
// tree and the bytes arrive second, from PDML - and between the two the screen
// can sit perfectly still for longer than the stability window. On a busy CI
// runner it did: the same commit produced a screenshot with both lower panes
// blank, and passed when the job was re-run. A gate that fails at random is a
// gate people learn to ignore.
//
// So the load itself is asked, and the picture only has to be still after it
// says it is done.
// And a load that says it is done is not a screen that is complete: the two
// halves arrive separately and the detail lands in a cache the panes read on
// their next render, so between them the packet list is drawn over two empty
// panes. Asking whether the focused packet's detail is there closes that
// window - the gate produced exactly that picture at random three times in one
// day before this.
func stillLoading() bool {
	if ui.Loader == nil {
		return false
	}
	if ui.Loader.PsmlLoader.IsLoading() || ui.Loader.PdmlLoader.IsLoading() {
		return true
	}
	return !ui.SelectedPacketDetailLoaded()
}

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

// castHold is how long each recorded frame is shown for, in seconds. Slow
// enough to read a screen of packet data, which is what the animation is for.
const castHold = 1.6

// captureCast records a frame per keystroke and writes them as one looping
// animation.
//
// A still picture cannot show what this program is for - moving through a
// capture. The keys are typed one at a time and the screen photographed after
// each has settled, so what the animation shows is a real session rather than
// a reconstruction, produced by the same renderer as the still screenshots and
// checkable the same way.
func captureCast(app *gowid.App, screen tcell.SimulationScreen, prefix string, keys string) {
	var frames [][][]screenshot.Cell

	// A keypress that changed nothing - the cursor already at the end of the
	// list, a key the focused pane ignores - would otherwise hold the same
	// picture for two turns and read as the animation being stuck.
	shot := func() {
		rows, _, _ := screenshot.Grab(screen)
		if len(frames) > 0 && screenshot.Text(rows) == screenshot.Text(frames[len(frames)-1]) {
			return
		}
		frames = append(frames, rows)
	}

	settle(app, screen, "")
	shot()

	for _, k := range parseKeys(keys) {
		before := screenshot.Text(frames[len(frames)-1])

		screen.InjectKey(k.Key(), k.Rune(), k.Modifiers())
		app.Run(gowid.RunFunction(func(gowid.IApp) {}))

		settle(app, screen, before)
		shot()
	}

	_, w, h := screenshot.Grab(screen)
	if err := writeCast(prefix, frames, w, h); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing screencast: %v\n", err)
	}

	app.Quit()
}

func writeCast(prefix string, frames [][][]screenshot.Cell, w, h int) error {
	if dir := filepath.Dir(prefix); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	svg := prefix + ".svg"
	if err := os.WriteFile(svg, []byte(screenshot.AnimatedSVG(frames, w, h, castHold)), 0o644); err != nil {
		return err
	}

	// The frames as text too, so CI can check the animation still shows what
	// it is meant to without comparing a picture.
	var b strings.Builder
	for i, f := range frames {
		fmt.Fprintf(&b, "=== frame %d ===\n%s", i+1, screenshot.Text(f))
	}
	txt := prefix + ".txt"
	if err := os.WriteFile(txt, []byte(b.String()), 0o644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Wrote %s and %s (%d frames, %dx%d)\n", svg, txt, len(frames), w, h)
	return nil
}

// typeKeys feeds keystrokes to the interface one at a time.
//
// The string is written in the same syntax the :map command documents:
// printable characters stand for themselves, and compound keys are named -
// <esc>, <enter>, <tab>, <down>, <pgdn>, <C-f>. Named keys are what let a
// screenshot reach a pane other than the one the program opens focused on;
// without them the only reachable states were the first screen and whatever a
// typed command could produce.
//
// Explicit key events rather than InjectKeyBytes: the byte form has to be
// parsed back into events by tcell's terminal decoder, which a simulation
// screen is not driving. One key per poll interval, because the command line
// builds its completion list between keystrokes and a burst arrives before it
// is ready for it.
func typeKeys(app *gowid.App, screen tcell.SimulationScreen, keys string) {
	for _, k := range parseKeys(keys) {
		screen.InjectKey(k.Key(), k.Rune(), k.Modifiers())

		app.Run(gowid.RunFunction(func(gowid.IApp) {}))
		time.Sleep(shotPollEvery)
	}
}

// parseKeys turns the flag's value into keypresses.
//
// A literal newline, tab or space in the argument means the named key, because
// that is how a shell writes one and how this flag was used before it
// understood names. The vim parser drops all three silently - which is how a
// screenshot came out showing ":expert" typed into the command line but never
// run, and, later, how ":streams websocket" was typed as ":streamswebsocket"
// and did nothing at all. A space is a separator in a vim mapping and not a
// printable character, so it has to be named; the failure is the quiet kind,
// because a command line with a word missing is still a command line.
func parseKeys(keys string) []gowid.Key {
	keys = strings.NewReplacer(
		"\r\n", "<enter>",
		"\n", "<enter>",
		"\r", "<enter>",
		"\t", "<tab>",
		" ", "<space>",
	).Replace(keys)

	res := make([]gowid.Key, 0, len(keys))
	for _, kp := range vim.VimStringToKeys(keys) {
		// vim.KeyPress is a defined type over gowid.Key, so it carries none of
		// its methods; the conversion is how the fields are read.
		res = append(res, gowid.Key(kp))
	}
	return res
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

		if now == last && strings.TrimSpace(now) != "" && !stillLoading() {
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
