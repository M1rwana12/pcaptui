// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.
//
// The shape of this loader follows pkg/capinfo, which Graham Clark wrote for
// the same job: run one short-lived process over the capture and hand its
// output to a dialog.

package stats

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/pkg/pcap"
	log "github.com/sirupsen/logrus"
)

//======================================================================

var Goroutinewg *sync.WaitGroup

//======================================================================

type ILoaderCmds interface {
	Stats(pcapfile string, zargs ...string) pcap.IPcapCommand
}

type commands struct{}

func MakeCommands() commands {
	return commands{}
}

var _ ILoaderCmds = commands{}

// Stats runs tshark for one statistic. -q suppresses the per-packet output so
// only the statistic itself reaches stdout.
func (c commands) Stats(pcapfile string, zargs ...string) pcap.IPcapCommand {
	// tshark takes several -z arguments and produces all of them from a single
	// pass over the capture, which is why the overview costs one read of the
	// file rather than three.
	args := []string{"-q"}
	for _, z := range zargs {
		args = append(args, "-z", z)
	}
	args = append(args, "-r", pcapfile)

	return &pcap.Command{
		Cmd: exec.Command(pcaptui.TSharkBin(), args...),
	}
}

//======================================================================

type Loader struct {
	cmds ILoaderCmds

	SuppressErrors bool // if true, don't report process errors e.g. at shutdown

	mainCtx      context.Context // cancelling this cancels the dependent contexts
	mainCancelFn context.CancelFunc

	statsCtx      context.Context
	statsCancelFn context.CancelFunc

	statsCmd pcap.IPcapCommand
}

func NewLoader(cmds ILoaderCmds, ctx context.Context) *Loader {
	res := &Loader{
		cmds: cmds,
	}
	res.mainCtx, res.mainCancelFn = context.WithCancel(ctx)
	return res
}

func (c *Loader) StopLoad() {
	if c.statsCancelFn != nil {
		c.statsCancelFn()
	}
}

//======================================================================

type IStatsCallbacks interface {
	OnStatsData(data string)
	AfterStatsEnd(success bool)
}

func (c *Loader) StartLoad(pcapfile string, zargs []string, app gowid.IApp, cb IStatsCallbacks) {
	pcaptui.TrackedGo(func() {
		c.loadStatsAsync(pcapfile, zargs, app, cb)
	}, Goroutinewg)
}

func (c *Loader) loadStatsAsync(pcapf string, zargs []string, app gowid.IApp, cb IStatsCallbacks) {
	c.statsCtx, c.statsCancelFn = context.WithCancel(c.mainCtx)

	procChan := make(chan int)
	pid := 0

	defer func() {
		if pid == 0 {
			close(procChan)
		}
	}()

	c.statsCmd = c.cmds.Stats(pcapf, zargs...)

	termChan := make(chan error)

	pcaptui.TrackedGo(func() {
		var err error
		cmd := c.statsCmd
		cancelledChan := c.statsCtx.Done()
		procChan := procChan
		state := pcap.NotStarted

		kill := func() {
			err := pcaptui.KillIfPossible(cmd)
			if err != nil {
				log.Infof("Did not kill tshark stats process: %v", err)
			}
		}

	loop:
		for {
			select {
			case err = <-termChan:
				state = pcap.Terminated
				if !c.SuppressErrors && err != nil {
					if _, ok := err.(*exec.ExitError); ok {
						// tshark exits non-zero for an invalid display filter,
						// and its stderr says which field is wrong. That text
						// is the whole value of the error to the user.
						pcap.HandleError(pcap.StatsCode, app, pcap.MakeUsefulError(cmd, err), cb)
					}
				}

			case pid := <-procChan:
				procChan = nil
				if pid != 0 {
					state = pcap.Started
					if cancelledChan == nil {
						kill()
					}
				}

			case <-cancelledChan:
				cancelledChan = nil
				if state == pcap.Started {
					kill()
				}
			}

			if state == pcap.Terminated || (procChan == nil && state == pcap.NotStarted) {
				break loop
			}
		}
	}, Goroutinewg)

	statsOut, err := c.statsCmd.StdoutReader()
	if err != nil {
		pcap.HandleError(pcap.StatsCode, app, err, cb)
		return
	}

	defer func() {
		cb.AfterStatsEnd(true)
	}()

	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		pcap.HandleBegin(pcap.StatsCode, app, cb)
	}))
	defer func() {
		app.Run(gowid.RunFunction(func(app gowid.IApp) {
			pcap.HandleEnd(pcap.StatsCode, app, cb)
		}))
	}()

	err = c.statsCmd.Start()
	if err != nil {
		err = fmt.Errorf("Error starting statistics command %v: %v", c.statsCmd, err)
		pcap.HandleError(pcap.StatsCode, app, err, cb)
		return
	}

	log.Infof("Started stats command %v with pid %d", c.statsCmd, c.statsCmd.Pid())

	pcaptui.TrackedGo(func() {
		termChan <- c.statsCmd.Wait()
	}, Goroutinewg)

	pid = c.statsCmd.Pid()
	procChan <- pid

	buf := new(bytes.Buffer)
	buf.ReadFrom(statsOut)

	cb.OnStatsData(buf.String())

	c.statsCancelFn()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
