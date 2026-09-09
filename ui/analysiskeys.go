// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/m1rwana12/pcaptui/pkg/stats"
)

//======================================================================

// analysisView is one of the views reached from the Analysis menu.
//
// Each has a key, and it is the same key in the menu and from the main view.
// Before this, none of them had one: the menu was reachable by mouse or by
// Escape (which lands on Misc, not Analysis), and the only other way in was to
// type the command's full name at `:`. The help text did not name the commands
// either, so a view like Expert Information - the thing the README calls the
// fastest way to find a problem - could be used for a whole session without
// ever being found.
type analysisView struct {
	Key rune
	// Command is the name typed at `:`. It is registered from this list too,
	// so a view cannot be reachable by key and not by name, and the command
	// help cannot list a name that opens nothing - or miss one that works,
	// which is what happened to http, dns and export.
	Command string
	Name    string
	Summary string
	Open    func(gowid.IApp)
}

// analysisViews is the single list the key handler, the Analysis menu and the
// help dialog are all built from, so a key cannot be listed in one place and
// bound in another.
func analysisViews() []analysisView {
	res := []analysisView{
		{
			Key:     'p',
			Command: "capinfo",
			Name:    "Capture file properties",
			Summary: "Size, duration, encapsulation, hashes",
			Open:    startCapinfo,
		},
		{
			Key:     's',
			Command: "streams",
			Name:    "Reassemble stream",
			Summary: "Follow the stream the selected packet belongs to",
			Open:    startStreamReassembly,
		},
		{
			Key:     'v',
			Command: "convs",
			Name:    "Conversations",
			Summary: "Who talked to whom, by packets and bytes",
			Open:    openConvsUi,
		},
		{
			Key:     'x',
			Command: "export",
			Name:    "Export objects",
			Summary: "Write the files this capture carried out to disk",
			Open:    openExportObjects,
		},
	}

	for _, stat := range stats.All {
		stat := stat
		res = append(res, analysisView{
			Key:     stat.Key,
			Command: stat.Command,
			Name:    stat.Name,
			Summary: stat.Summary,
			Open:    func(app gowid.IApp) { startStats(stat, app) },
		})
	}

	return res
}

// analysisKeyPress opens the view bound to r, if any.
func analysisKeyPress(r rune, app gowid.IApp) bool {
	for _, v := range analysisViews() {
		if v.Key == r {
			v.Open(app)
			return true
		}
	}
	return false
}

// analysisKeyHelp is the block of lines the help dialog shows. It is generated
// rather than written out, because a help text maintained separately from the
// bindings is a help text that goes stale - this one already had, listing
// `:config` and `:logs` as Unix-only after they stopped being so.
func analysisKeyHelp() string {
	var b strings.Builder
	for _, v := range analysisViews() {
		fmt.Fprintf(&b, "%c__ - %s\n", v.Key, v.Name)
	}
	return b.String()
}

// analysisCommandHelp is the block of `:` commands in the command-line help,
// generated from the same list so that adding a view cannot leave it out.
//
// It was left out three times: http, dns and export all worked at `:` and none
// of them appeared in `:help cmdline`, so a third of the analysis views were
// undiscoverable from the help that exists to list them.
func analysisCommandHelp() string {
	var b strings.Builder
	for _, v := range analysisViews() {
		fmt.Fprintf(&b, "%s - %s\n", padCommand(v.Command), v.Summary)
	}
	return strings.TrimRight(b.String(), "\n")
}

// padCommand pads a name out to the width the hand-written lines around it
// use, with the underscores they use rather than spaces.
func padCommand(name string) string {
	const width = 13
	if len(name) >= width {
		return name
	}
	return name + strings.Repeat("_", width-len(name))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
