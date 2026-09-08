// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"
	"sync"
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
	stopOnce         sync.Once
	pleaseWaitClosed bool
}

// stopSpinner ends the goroutine that animates the please-wait spinner.
//
// It has to be reachable from a path that runs even after app.Quit(), because
// gowid stops dispatching callbacks then: closing the channel from AfterEnd
// alone would leave that goroutine parked, and it is registered with
// Goroutinewg, so the program would never finish exiting.
func (t *statsParseHandler) stopSpinner() {
	t.stopOnce.Do(func() {
		if t.stop != nil {
			close(t.stop)
		}
	})
}

var _ stats.IStatsCallbacks = (*statsParseHandler)(nil)
var _ pcap.IBeforeBegin = (*statsParseHandler)(nil)
var _ pcap.IAfterEnd = (*statsParseHandler)(nil)

func (t *statsParseHandler) OnStatsData(data string) {
	t.data = strings.Replace(data, "\r\n", "\n", -1) // For windows...
}

func (t *statsParseHandler) AfterStatsEnd(success bool) {
	// Reached through a plain defer in the loader, so it runs whether or not
	// the app is still accepting callbacks.
	t.stopSpinner()
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

		v := t.view()
		if v.empty() {
			OpenMessage(fmt.Sprintf("%s\n\n%s",
				v.Heading, statsEmptyMessage(t.filter)), appView, app)
			return
		}

		openStatsDialog(v, app)
	}))
	t.stopSpinner()
}

// view turns tshark's output into the dialog.
//
// A statistic narrowed by a filter that matches no packets prints absolutely
// nothing - no header, no empty table - so no rows is a real answer and the
// caller says so in words rather than opening an empty box.
func (t *statsParseHandler) view() statsView {
	v := statsView{Heading: statsHeading(t.stat.Name, t.filter)}

	if strings.TrimSpace(t.data) == "" {
		return v
	}

	switch t.stat.Command {
	case stats.Expert.Command:
		body := expertLines(stats.ParseExpert(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.ProtoHierarchy.Command:
		body := hierarchyLines(stats.ParseHierarchy(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	default:
		// A statistic this package does not know how to lay out is still worth
		// showing; it just cannot offer a filter for any of its rows.
		for _, line := range strings.Split(strings.TrimRight(t.data, "\n"), "\n") {
			v.Rows = append(v.Rows, statsLine{Text: line})
		}
	}

	return v
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
