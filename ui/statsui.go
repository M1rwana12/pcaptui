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
	"github.com/m1rwana12/pcaptui/pkg/capinfo"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
	"github.com/m1rwana12/pcaptui/pkg/stats"
	log "github.com/sirupsen/logrus"
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

	loader := stats.NewLoader(stats.MakeCommands(), Loader.Context())
	StatsLoader = loader

	h := &statsParseHandler{stat: stat, filter: filter}
	h.wait = newStatsWait(stat.Name, func() {
		// On the goroutine that draws, because it arrives from the dialog.
		h.cancelled = true
		loader.StopLoad()
		log.Infof("Cancelled %s", stat.Name)
	})

	pcapfile := Loader.PcapPdml
	zargs := stat.ZArgs(filter)

	if stat.Command != stats.Overview.Command {
		loader.StartLoad(pcapfile, zargs, app, h)
		return
	}

	// The overview also says what the file is, and capinfos answers that for a
	// fraction of what the -z pass costs - 0.27 s against 9 to 11 s on a 44 MB
	// capture. Asked first, off the goroutine that draws, so that the dialog
	// opens with all four sections rather than growing one later.
	//
	// A capinfos that fails is not worth stopping for: the three statistics
	// are still the answer to three of the four questions, and the section is
	// simply not drawn.
	pcaptui.TrackedGo(func() {
		info, err := capinfo.Read(pcapfile)
		if err != nil {
			log.Warnf("Could not read the file's own properties: %v", err)
		}

		// The traffic over time needs a bucket size, and the bucket size needs
		// the length of the capture - which capinfos has just said. It rides in
		// the same pass as the rest.
		// Guarded twice over, because capinfos failing is a case that reaches
		// here: "39,571274 seconds" has a number in front of a word, and an
		// empty Duration has neither.
		if fields := strings.Fields(info.Duration); len(fields) > 0 {
			if seconds, ok := pcaptui.ParseDecimal(fields[0]); ok {
				zargs = append(zargs, stats.IOStatArg(seconds, filter))
			}
		}

		app.Run(gowid.RunFunction(func(app gowid.IApp) {
			h.facts = info
			loader.StartLoad(pcapfile, zargs, app, h)
		}))
	}, Goroutinewg)
}

//======================================================================

type statsParseHandler struct {
	stat   stats.Stat
	filter string

	data string

	// facts is what capinfos said about the file, fetched before the pass
	// starts because it costs a fraction of what the pass does and answers the
	// one question the three statistics do not: when.
	facts capinfo.Info

	wait *statsWait

	// failed is set when tshark reported a problem, and cancelled when the
	// user closed the wait dialog, so that neither outcome then claims there
	// was nothing to report. Both are only ever touched inside app.Run, which
	// is to say on the one goroutine that draws.
	failed    bool
	cancelled bool

	tick     *time.Ticker // for updating the spinner
	stop     chan struct{}
	stopOnce sync.Once
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

// Without this one, pcap.HandleError finds nothing to call and returns having
// done nothing at all. tshark exits non-zero for a display filter it cannot
// parse, for a -z argument this build does not have, and for a capture that
// has been deleted or made unreadable since it was opened - and its stderr
// says which. All of that was dropped on the floor, and what the user saw
// instead was "Nothing to report for this capture." The overview now opens by
// itself on every file, so that sentence was the program's answer to a
// failure it had been told about.
var _ pcap.IOnError = (*statsParseHandler)(nil)

func (t *statsParseHandler) OnStatsData(data string) {
	t.data = strings.Replace(data, "\r\n", "\n", -1) // For windows...
}

func (t *statsParseHandler) AfterStatsEnd(success bool) {
	// Reached through a plain defer in the loader, so it runs whether or not
	// the app is still accepting callbacks.
	t.stopSpinner()
}

func (t *statsParseHandler) OnError(code pcap.HandlerCode, app gowid.IApp, err error) {
	if code&pcap.StatsCode == 0 {
		return
	}

	log.Error(err)

	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		if t.cancelled {
			// tshark was killed on purpose; its complaint about that is not
			// news to the person who did it.
			return
		}
		t.failed = true
		t.wait.close(app)
		OpenError(fmt.Sprintf("%s\n\n%v", t.stat.Name, err), app)
	}))
}

func (t *statsParseHandler) BeforeBegin(code pcap.HandlerCode, app gowid.IApp) {
	if code&pcap.StatsCode == 0 {
		return
	}
	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		t.wait.open(app)
	}))

	t.tick = time.NewTicker(time.Duration(200) * time.Millisecond)
	t.stop = make(chan struct{})

	pcaptui.TrackedGo(func() {
	Loop:
		for {
			select {
			case <-t.tick.C:
				app.Run(gowid.RunFunction(func(app gowid.IApp) {
					t.wait.update()
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
		t.wait.close(app)

		if t.cancelled {
			// The user asked for this to stop. Opening the result now would
			// land it on top of whatever they moved on to.
			return
		}

		if t.failed {
			// The error dialog has already said what went wrong; saying
			// "nothing to report" on top of it would contradict it.
			return
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

// filterAsUsed is the filter that actually narrowed the result, which for a
// tap that ignores one is none - so "nothing to report for this display
// filter" is not said about a table the filter never touched.
func (t *statsParseHandler) filterAsUsed() string {
	if t.stat.IgnoresFilter() {
		return ""
	}
	return t.filter
}

// view turns tshark's output into the dialog.
//
// A statistic narrowed by a filter that matches no packets prints absolutely
// nothing - no header, no empty table - so no rows is a real answer and the
// caller says so in words rather than opening an empty box.
func (t *statsParseHandler) view() statsView {
	v := statsView{Heading: statsHeading(t.stat.Name, t.filter, t.stat.IgnoresFilter())}

	if strings.TrimSpace(t.data) == "" {
		return v
	}

	switch t.stat.Command {
	case stats.Overview.Command:
		body := overviewLines(t.data, t.facts)
		v.Header, v.Rows = body.Header, body.Rows
	case stats.Expert.Command:
		body := expertLines(stats.ParseExpert(t.data), stats.ExpertTotals(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.ProtoHierarchy.Command:
		body := hierarchyLines(stats.ParseHierarchy(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.Endpoints.Command:
		body := endpointLines(stats.ParseEndpoints(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.HTTP.Command:
		body := httpLines(stats.ParseTree(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.DNS.Command:
		body := dnsLines(stats.ParseTree(t.data))
		v.Header, v.Rows = body.Header, body.Rows
	case stats.Credentials.Command:
		body := credentialLines(stats.ParseCredentials(t.data))
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
