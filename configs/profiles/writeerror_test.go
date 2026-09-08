// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package profiles

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// unwritable returns a viper pointed at a path that cannot be written,
// because a file stands where one of its parent directories would have to be.
// That is portable; making a directory read-only is not - on Windows the
// read-only attribute does not stop a file being created inside.
func unwritable(t *testing.T) *viper.Viper {
	t.Helper()

	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))

	v := viper.New()
	v.SetConfigFile(filepath.Join(blocker, "sub", "pcaptui.toml"))
	v.Set("main.theme", "dracula")

	return v
}

func reset() {
	writeErrorOnce = sync.Once{}
	OnWriteError = nil
}

//======================================================================

func TestASuccessfulWriteIsSilent(t *testing.T) {
	reset()
	defer reset()

	called := 0
	OnWriteError = func(string, error) { called++ }

	v := viper.New()
	v.SetConfigFile(filepath.Join(t.TempDir(), "pcaptui.toml"))
	v.Set("main.theme", "dracula")

	assert.NoError(t, writeConfig(v))
	assert.Zero(t, called)
}

// The whole point: the write used to fail and say nothing, so a read-only or
// full configuration directory gave a program that accepted every setting and
// remembered none of them.
func TestAFailedWriteIsReported(t *testing.T) {
	reset()
	defer reset()

	var gotPath string
	var gotErr error
	OnWriteError = func(p string, e error) { gotPath, gotErr = p, e }

	err := writeConfig(unwritable(t))

	require.Error(t, err, "writing there should not have succeeded")
	assert.Error(t, gotErr, "the failure was not reported")
	assert.Contains(t, gotPath, "pcaptui.toml",
		"the report has to name the file the user must go and fix")
}

// The cause does not go away, and fifty settings against a read-only directory
// would otherwise be fifty dialogs.
func TestTheUserIsToldOnceNotEveryTime(t *testing.T) {
	reset()
	defer reset()

	called := 0
	OnWriteError = func(string, error) { called++ }

	v := unwritable(t)
	for i := 0; i < 5; i++ {
		assert.Error(t, writeConfig(v))
	}

	assert.Equal(t, 1, called)
}

// The UI sets the hook during Build; writes happen before that, and during
// tests there is no UI at all.
func TestNoHookIsNotACrash(t *testing.T) {
	reset()
	defer reset()

	assert.Error(t, writeConfig(unwritable(t)))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
