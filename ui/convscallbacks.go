// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"strings"
	"sync"
	"time"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
)

//======================================================================

type IOnDataSync interface {
	OnData(data string, app gowid.IApp)
	OnCancel(gowid.IApp)
}

type convsParseHandler struct {
	app              gowid.IApp
	tick             *time.Ticker // for updating the spinner
	stop             chan struct{}
	stopOnce         sync.Once
	ondata           IOnDataSync
	pleaseWaitClosed bool
}

var _ pcap.IBeforeBegin = (*convsParseHandler)(nil)
var _ pcap.IAfterEnd = (*convsParseHandler)(nil)

func (t *convsParseHandler) OnData(data string) {
	data = strings.Replace(data, "\r\n", "\n", -1) // For windows...

	if t.ondata != nil {
		t.app.Run(gowid.RunFunction(func(app gowid.IApp) {
			t.ondata.OnData(data, app)
		}))
	}
}

func (t *convsParseHandler) AfterDataEnd(success bool) {
	// Reached through a plain defer in the loader, so it runs even after
	// app.Quit() has stopped gowid dispatching callbacks.
	t.stopSpinner()
	if t.ondata != nil && !success {
		t.app.Run(gowid.RunFunction(func(app gowid.IApp) {
			t.ondata.OnCancel(app)
		}))
	}
}

func (t *convsParseHandler) BeforeBegin(code pcap.HandlerCode, app gowid.IApp) {
	if code&pcap.ConvCode == 0 {
		return
	}
	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		OpenPleaseWait(appView, t.app)
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

func (t *convsParseHandler) AfterEnd(code pcap.HandlerCode, app gowid.IApp) {
	if code&pcap.ConvCode == 0 {
		return
	}
	t.app.Run(gowid.RunFunction(func(app gowid.IApp) {
		if !t.pleaseWaitClosed {
			t.pleaseWaitClosed = true
			ClosePleaseWait(t.app)
		}
	}))
	t.stopSpinner()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:

// stopSpinner ends the please-wait spinner goroutine.
//
// Closing the channel from AfterEnd alone is not enough: that callback is
// dispatched from inside app.Run, and after app.Quit() gowid drops those, so
// the goroutine - registered with Goroutinewg - would keep the program from
// finishing its exit.
func (t *convsParseHandler) stopSpinner() {
	t.stopOnce.Do(func() {
		if t.stop != nil {
			close(t.stop)
		}
	})
}
