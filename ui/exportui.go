// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/dialog"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/export"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// openExportObjects asks which kind of object to pull out of the capture.
//
// The kinds come from tshark rather than from a list here, because which ones
// exist is a property of the tshark on this machine.
func openExportObjects(app gowid.IApp) {
	if Loader.PcapPdml == "" {
		OpenError("No pcap loaded.", app)
		return
	}

	// Asking tshark what it can export is a subprocess, and every other tshark
	// this program runs is started from a goroutine. Run on the goroutine that
	// draws, it freezes the whole interface for as long as tshark takes to
	// start - which on this platform is not nothing.
	pcaptui.TrackedGo(func() {
		types, err := export.Types()

		app.Run(gowid.RunFunction(func(app gowid.IApp) {
			if err != nil {
				OpenError(err.Error(), app)
				return
			}
			openExportTypes(types, app)
		}))
	}, Goroutinewg)
}

// openExportTypes is the picker itself, opened once tshark has said which
// kinds of object this build knows how to write.
func openExportTypes(types []string, app gowid.IApp) {

	var d *dialog.Widget

	choose := func(typ string) func(gowid.IApp, gowid.IWidget) {
		return func(app gowid.IApp, w gowid.IWidget) {
			if d != nil {
				d.Close(app)
			}
			exportObjects(typ, app)
		}
	}

	ws := make([]gowid.IWidget, 0, len(types))
	for _, typ := range types {
		b := button.NewBare(text.New(typ))
		b.OnClick(gowid.MakeWidgetCallback("cb", choose(typ)))
		ws = append(ws,
			styled.NewInvertedFocus(selectable.New(b), gowid.MakePaletteRef("dialog")))
	}

	walker := list.NewSimpleListWalker(ws)

	content := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: text.New(exportHeading()),
			D:       flow,
		},
		&gowid.ContainerWidget{
			IWidget: text.New(""),
			D:       flow,
		},
		&gowid.ContainerWidget{
			IWidget: list.New(walker),
			D:       weight(1),
		},
	})

	d = dialog.New(
		framed.NewSpace(content),
		dialog.Options{
			Buttons:         dialog.CloseOnly,
			NoShadow:        true,
			BackgroundStyle: gowid.MakePaletteRef("dialog"),
			BorderStyle:     gowid.MakePaletteRef("dialog"),
			ButtonStyle:     gowid.MakePaletteRef("dialog-button"),
			Modal:           true,
			// Otherwise the dialog opens with Close focused, and Enter on what
			// looks like the first row closes it instead of choosing.
			FocusOnWidget: true,
			TabToButtons:  true,
		},
	)
	YesNo = d

	dialog.OpenExt(d, appView, ratio(0.5), ratio(0.6), app)
	walker.SetFocus(list.ListPos(0), app)
}

// exportHeading says what is about to happen, including the part a user would
// otherwise have to discover: the display filter is not applied.
func exportHeading() string {
	if Loader.DisplayFilter() == "" {
		return "Export the files carried by this capture"
	}
	return "Export the files carried by this capture\n" +
		"(the whole capture, not the display filter - an object is\n" +
		" reassembled from its packets, and hiding some of them would\n" +
		" write a file that is quietly short)"
}

//======================================================================

// exportObjects runs the export and says where the files went.
func exportObjects(typ string, app gowid.IApp) {
	pcapPath := Loader.PcapPdml
	dir := export.Dir(pcapPath, typ)

	OpenPleaseWait(appView, app)

	pcaptui.TrackedGo(func() {
		written, err := export.Objects(pcapPath, typ, dir)

		app.Run(gowid.RunFunction(func(app gowid.IApp) {
			ClosePleaseWait(app)

			if err != nil {
				log.Errorf("Could not export %s objects: %v", typ, err)
				OpenError(err.Error(), app)
				return
			}

			log.Infof("Exported %d %s object(s) to %s", len(written), typ, dir)
			OpenMessage(exportResult(typ, dir, written), appView, app)
		}))
	}, Goroutinewg)
}

// exportResult is what to say afterwards. Nothing exported is a real answer
// and gets a sentence of its own: tshark exits successfully either way, so
// without this the two outcomes would look the same.
func exportResult(typ string, dir string, written []string) string {
	if len(written) == 0 {
		return fmt.Sprintf("No %s objects in this capture.", typ)
	}

	shown := written
	const most = 10
	more := ""
	if len(shown) > most {
		more = fmt.Sprintf("\n... and %d more", len(shown)-most)
		shown = shown[:most]
	}

	return fmt.Sprintf("%d %s object(s) written to\n%s\n\n%s%s",
		len(written), typ, dir, strings.Join(shown, "\n"), more)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
