// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package profiles

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//======================================================================

// OnWriteError is called the first time saving the configuration fails, and
// not again.
//
// Every setting the program remembers goes through one write: the files you
// have opened recently, the theme, the column layout, the profile in use.
// Those writes discarded their error, so a configuration directory that is
// read-only, full, or owned by somebody else produced a program that appeared
// to accept every preference and forgot all of them - with nothing said
// anywhere.
//
// Once, not per write, because the cause does not go away: fifty settings
// against a read-only directory would otherwise be fifty dialogs. The path is
// passed because it is the thing the user has to go and fix.
//
// Set by the UI at startup. Nil is fine - the failure is still logged.
var OnWriteError func(path string, err error)

var writeErrorOnce sync.Once

// writeConfig saves v, and reports a failure to save rather than dropping it.
func writeConfig(v *viper.Viper) error {
	err := v.WriteConfig()
	if err == nil {
		return nil
	}

	path := v.ConfigFileUsed()
	log.Warnf("Could not write the configuration to %s: %v", path, err)

	writeErrorOnce.Do(func() {
		if OnWriteError != nil {
			OnWriteError(path, err)
		}
	})

	return err
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
