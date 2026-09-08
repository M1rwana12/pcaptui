// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/dialog"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

// statsLine is one line of a statistics dialog.
//
// Filter is what the line is *about*, expressed as a display filter, and empty
// for a line that is only text - a heading, a rule, a blank. A line that has
// one can be selected, and selecting it narrows the packet list to those
// packets.
type statsLine struct {
	Text   string
	Filter string
}

// statsView is a whole statistics dialog: a heading, column titles that stay
// put, and the rows that scroll under them.
type statsView struct {
	Heading string
	Header  []string
	Rows    []statsLine
}

func (v statsView) empty() bool {
	return len(v.Rows) == 0
}

func (l statsLine) actionable() bool {
	return l.Filter != ""
}

//======================================================================

// openStatsDialog shows the lines of a statistic, with the ones that stand for
// packets selectable.
//
// Expert Information used to be handed to OpenMessageForCopy, which grows to
// fit its text and cannot scroll: a capture with more distinct expert items
// than the terminal has rows produced a dialog whose bottom could not be
// reached. This is a list, so it scrolls, and its rows do something.
func openStatsDialog(v statsView, app gowid.IApp) {
	var d *dialog.Widget

	apply := func(filter string) func(gowid.IApp, gowid.IWidget) {
		return func(app gowid.IApp, w gowid.IWidget) {
			if d != nil {
				d.Close(app)
			}
			FilterWidget.SetValue(filter, app)
			RequestNewFilter(filter, app)
		}
	}

	ws := make([]gowid.IWidget, 0, len(v.Rows))
	for _, l := range v.Rows {
		ws = append(ws, statsLineWidget(l, apply))
	}

	walker := list.NewSimpleListWalker(ws)
	body := list.New(walker)

	// The heading and the column titles stay put while the rows scroll under
	// them. They are also the only way to keep them on screen at all: a list
	// opens positioned on its first selectable child, so anything above that
	// scrolls out of sight before the dialog is even drawn.
	above := []gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: text.New(v.Heading),
			D:       flow,
		},
		&gowid.ContainerWidget{
			IWidget: text.New(""),
			D:       flow,
		},
	}
	for _, h := range v.Header {
		above = append(above, &gowid.ContainerWidget{
			IWidget: text.New(h),
			D:       flow,
		})
	}

	content := pile.New(append(above, &gowid.ContainerWidget{
		IWidget: body,
		D:       weight(1),
	}))

	d = dialog.New(
		framed.NewSpace(content),
		dialog.Options{
			Buttons:         dialog.CloseOnly,
			NoShadow:        true,
			BackgroundStyle: gowid.MakePaletteRef("dialog"),
			BorderStyle:     gowid.MakePaletteRef("dialog"),
			ButtonStyle:     gowid.MakePaletteRef("dialog-button"),
			Modal:           true,
			// Without this a dialog opens with its buttons focused, and
			// measured: Enter on the first row closed the dialog and applied
			// nothing, because the keypress went to Close. The rows are the
			// reason this dialog exists.
			FocusOnWidget: true,
			// Tab still reaches Close, which is the only other thing here.
			TabToButtons: true,
		},
	)
	YesNo = d

	dialog.OpenExt(d, appView, ratio(0.9), ratio(0.8), app)
}

func statsLineWidget(l statsLine, apply func(string) func(gowid.IApp, gowid.IWidget)) gowid.IWidget {
	if !l.actionable() {
		return text.New(l.Text)
	}

	b := button.NewBare(text.New(l.Text))
	b.OnClick(gowid.MakeWidgetCallback("cb", apply(l.Filter)))

	// Inverted on focus rather than coloured: these rows sit on the dialog's
	// own background, and the accent belongs to the packet list.
	return styled.NewInvertedFocus(selectable.New(b), gowid.MakePaletteRef("dialog"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
