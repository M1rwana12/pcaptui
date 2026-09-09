// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/dialog"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/spinner"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

// statsWait is the please-wait dialog the statistics use, and unlike the one
// shared with the other loaders it can be got out of.
//
// A statistic is one pass over the whole capture. Measured on this machine:
// 2.8 MB took 1.07 s, 11 MB 1.68 s, 22 MB 3.51 s and 44 MB 11.6 s - roughly
// 4-5 MB/s past ten megabytes, and the README's own pitch is a two-gigabyte
// capture. The overview opens by itself when a file finishes loading, so that
// wait is one the program starts on its own.
//
// The shared dialog has no buttons at all. Escape closed it, which looked like
// cancelling and was not: tshark went on reading and the result opened over
// whatever the user had moved on to. This one says Cancel, means it, and
// treats Escape the same way.
type statsWait struct {
	dlg     *dialog.Widget
	spin    *spinner.Widget
	stop    func()
	closing bool
}

func newStatsWait(name string, stop func()) *statsWait {
	w := &statsWait{
		spin: spinner.New(spinner.Options{
			Styler: gowid.MakePaletteRef("progress-spinner"),
		}),
		stop: stop,
	}

	cancel := dialog.Button{
		Msg: "Cancel",
		Action: gowid.MakeWidgetCallback("cb",
			gowid.WidgetChangedFunction(func(app gowid.IApp, _ gowid.IWidget) {
				w.dlg.Close(app)
			})),
	}

	w.dlg = dialog.New(
		framed.NewSpace(
			pile.NewFlow(
				&gowid.ContainerWidget{
					IWidget: text.New(" " + name + "... "),
					D:       gowid.RenderFixed{},
				},
				// A filler of its own rather than the package-level one, which is
				// built when the interface is and is therefore nil until then.
				fill.New(' '),
				w.spin,
			)),
		dialog.Options{
			Buttons:         []dialog.Button{cancel},
			NoShadow:        true,
			BackgroundStyle: gowid.MakePaletteRef("dialog"),
			BorderStyle:     gowid.MakePaletteRef("dialog"),
			ButtonStyle:     gowid.MakePaletteRef("dialog-button"),
		},
	)

	// Fires on open as well as on close, so the open is skipped by asking the
	// dialog which it just did. Escape and the button both arrive here, which
	// is the point: one way out, one thing it means.
	w.dlg.OnOpenClose(gowid.MakeWidgetCallback("cb",
		gowid.WidgetChangedFunction(func(app gowid.IApp, _ gowid.IWidget) {
			if w.dlg.IsOpen() {
				return
			}
			w.onClose()
		})))

	return w
}

// onClose is the decision the dialog's closing turns on: the user pressing
// Cancel or Escape means stop, and the load taking its own dialog down does
// not.
func (w *statsWait) onClose() {
	if w.closing {
		return
	}
	w.stop()
}

// open shows it, and starts the spinner turning.
func (w *statsWait) open(app gowid.IApp) {
	w.dlg.Open(appView, fixed, app)
}

// close takes it down without cancelling, for the load that finished.
func (w *statsWait) close(app gowid.IApp) {
	if !w.dlg.IsOpen() {
		return
	}
	w.closing = true
	w.dlg.Close(app)
	w.closing = false
}

func (w *statsWait) update() {
	w.spin.Update()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
