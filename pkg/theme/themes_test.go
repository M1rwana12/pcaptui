// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package theme

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rakyll/statik/fs"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

//======================================================================

var assetsDir = filepath.Join("..", "..", "assets")

//======================================================================

// The hex pane draws four things at once: the line-number gutter, the
// protocol layer the cursor is in, the field within it, and the cursor byte
// itself. Each has a colour for when the pane has focus and one for when it
// does not.
//
// Two rules hold across every built-in theme, and they are what these tests
// are for - the 8-colour theme once broke both, and nothing said so.

// A pane that looks the same focused and unfocused cannot tell you where your
// keystrokes will go. Something in the hex pane has to change.
func TestEveryThemeShowsWhetherTheHexPaneHasFocus(t *testing.T) {
	forEachTheme(t, func(t *testing.T, name string, v *viper.Viper, mode string) {
		for _, level := range []string{"byte", "field", "layer", "interval"} {
			sel := colorPair(v, mode, "hex-"+level+"-selected")
			unsel := colorPair(v, mode, "hex-"+level+"-unselected")
			require.NotEqual(t, sel, unsel,
				"%s [%s]: hex-%s looks identical focused and unfocused",
				name, mode, level)
		}
	})
}

// A field is drawn inside its layer and the cursor inside the field. If two
// of them share a colour, the pane stops saying where one ends.
func TestAFocusedHexPaneSeparatesCursorFieldAndLayer(t *testing.T) {
	forEachTheme(t, func(t *testing.T, name string, v *viper.Viper, mode string) {
		seen := map[string]string{}
		for _, level := range []string{"byte", "field", "layer"} {
			pair := colorPair(v, mode, "hex-"+level+"-selected")
			if other, dup := seen[pair]; dup {
				t.Errorf("%s [%s]: hex-%s-selected and hex-%s-selected are both %s",
					name, mode, other, level, pair)
			}
			seen[pair] = level
		}
	})
}

//======================================================================

// The themes the program ships are the ones statik compiled into the binary,
// not the files under assets/. Editing a theme and forgetting to regenerate
// leaves the old colours in every build, and nothing else would say so - the
// tests above read the files, and the program reads the binary.
func TestTheBuiltInAssetsAreTheFilesInTheTree(t *testing.T) {
	statikFS, err := fs.New()
	require.NoError(t, err)

	found := 0
	err = filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.HasPrefix(path, filepath.Join(assetsDir, "statik")) {
			return nil
		}
		found++

		rel := filepath.ToSlash(strings.TrimPrefix(path, assetsDir))

		onDisk, err := os.ReadFile(path)
		require.NoError(t, err)

		embedded, err := readEmbedded(statikFS, rel)
		require.NoError(t, err,
			"%s is not compiled into the binary; run statik -src=. -f in assets/", rel)

		require.Equal(t, string(onDisk), embedded,
			"%s differs from the copy compiled into the binary; run statik -src=. -f in assets/", rel)

		return nil
	})
	require.NoError(t, err)
	require.NotZero(t, found, "no assets found to compare")
}

func readEmbedded(statikFS http.FileSystem, name string) (string, error) {
	file, err := statikFS.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	return string(b), err
}

//======================================================================

func forEachTheme(t *testing.T, f func(*testing.T, string, *viper.Viper, string)) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(assetsDir, "themes", "*.toml"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "no built-in themes found to check")

	for _, path := range files {
		name := filepath.Base(path)

		file, err := os.Open(path)
		require.NoError(t, err)

		v := viper.New()
		v.SetConfigType("toml")
		err = v.ReadConfig(file)
		file.Close()
		require.NoError(t, err, "%s does not parse", name)

		for _, mode := range []string{"dark", "light"} {
			f(t, name, v, mode)
		}
	}
}

// colorPair renders one theme entry as a string, so that two of them can be
// compared and named in a failure.
func colorPair(v *viper.Viper, mode string, key string) string {
	got := v.GetStringSlice(fmt.Sprintf("%s.%s", mode, key))
	if len(got) == 0 {
		return "<missing>"
	}
	return fmt.Sprintf("%v", got)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
