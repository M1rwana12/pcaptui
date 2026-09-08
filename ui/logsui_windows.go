// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"os"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui"
	"github.com/m1rwana12/pcaptui/configs/profiles"
)

//======================================================================

// The Unix versions of these run the user's pager inside a terminal widget.
// There is no pty to run it in on Windows, so the file is read and shown in a
// scrollable dialog instead.
//
// They used to be empty, and the commands and menu items that reach them were
// compiled out - so `:config` was not a command on Windows at all, while the
// README, the FAQ and the User Guide all told the reader to run it. A dialog
// that names the path and shows the file answers the question those documents
// send people here to ask.

func openLogsUi(app gowid.IApp) {
	OpenScrollableText(
		fileForDialog("Log file", pcaptui.CacheFile("pcaptui.log"), tailLimit),
		appView, ratio(0.9), ratio(0.8), app,
	)
}

func openConfigUi(app gowid.IApp) {
	// The config on disk holds only what has been changed from the defaults,
	// which is usually nothing, and that is the least useful thing to show
	// somebody asking what their configuration is. WriteConfigAs writes what
	// is actually in force.
	tmp, err := os.CreateTemp("", "pcaptui-*.toml")
	if err != nil {
		OpenError(fmt.Sprintf("Could not create temp file: %v", err), app)
		return
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := profiles.WriteConfigAs(tmp.Name()); err != nil {
		OpenError(fmt.Sprintf("Could not read the configuration\n\n%v", err), app)
		return
	}

	// Headed with the real path, not the temporary file it was written to -
	// the path is the thing the user came for.
	OpenScrollableText(
		fmt.Sprintf("Config file: %s\n\n%s",
			pcaptui.ConfFile("pcaptui.toml"),
			fileBody(tmp.Name(), tailLimit),
		),
		appView, ratio(0.9), ratio(0.8), app,
	)
}

//======================================================================

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
