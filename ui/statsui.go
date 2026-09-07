// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
	"github.com/m1rwana12/pcaptui/pkg/stats"
)

var StatsLoader *stats.Loader

//======================================================================

// startStats runs one tshark statistic over the loaded capture and shows the
// result.
//
// Deliberately not cached, unlike capinfo: the result depends on the display
// filter as well as the file, and the user asks for it explicitly each time.
// A cache here would mean showing a statistic for a filter that is no longer
// in the box.
func startStats(stat stats.Stat, app gowid.IApp) {
	if Loader.PcapPdml == "" {
		OpenError("No pcap loaded.", app)
		return
	}

	filter := Loader.DisplayFilter()

	StatsLoader = stats.NewLoader(stats.MakeCommands(), Loader.Context())

	StatsLoader.StartLoad(
		Loader.PcapPdml,
		stat.ZArg(filter),
		app,
		&statsParseHandler{stat: stat, filter: filter},
	)
}

//======================================================================

type statsParseHandler struct {
	stat   stats.Stat
	filter string

	data string

	tick             *time.Ticker // for updating the spinner
	stop             chan struct{}
	pleaseWaitClosed bool
}

var _ stats.IStatsCallbacks = (*statsParseHandler)(nil)
var _ pcap.IBeforeBegin = (*statsParseHandler)(nil)
var _ pcap.IAfterEnd = (*statsParseHandler)(nil)

func (t *statsParseHandler) OnStatsData(data string) {
	t.data = strings.Replace(data, "\r\n", "\n", -1) // For windows...
}

func (t *statsParseHandler) AfterStatsEnd(success bool) {
}

func (t *statsParseHandler) BeforeBegin(code pcap.HandlerCode, app gowid.IApp) {
	if code&pcap.StatsCode == 0 {
		return
	}
	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		OpenPleaseWait(appView, app)
	}))

	t.tick = time.NewTicker(time.Duration(200) * time.Millisecond)
	t.stop = make(chan struct{})

	pcaptui.TrackedGo(func() {
	Loop:
		for {
			select {
			case <-t.tick.C:
				app.Run(gowid.RunFunction(func(app gowid.IApp) {
					pleaseWaitSpinner.Update()
				}))
			case <-t.stop:
				break Loop
			}
		}
	}, Goroutinewg)
}

func (t *statsParseHandler) AfterEnd(code pcap.HandlerCode, app gowid.IApp) {
	if code&pcap.StatsCode == 0 {
		return
	}
	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		if !t.pleaseWaitClosed {
			t.pleaseWaitClosed = true
			ClosePleaseWait(app)
		}

		OpenMessageForCopy(t.report(), appView, app)
	}))
	close(t.stop)
}

// report titles the output and, when tshark found nothing, says so.
//
// A statistic narrowed by a filter that matches no packets prints absolutely
// nothing - no header, no empty table. Handing that to the dialog would show
// the user a blank box and no reason for it.
func (t *statsParseHandler) report() string {
	var b strings.Builder

	b.WriteString(t.stat.Name)
	if t.filter != "" {
		b.WriteString(fmt.Sprintf("\nDisplay filter: %s", t.filter))
	}
	b.WriteString("\n\n")

	body := strings.TrimSpace(t.data)
	if body == "" {
		if t.filter != "" {
			b.WriteString("Nothing to report for this display filter.")
		} else {
			b.WriteString("Nothing to report for this capture.")
		}
		return b.String()
	}

	b.WriteString(body)
	return b.String()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
