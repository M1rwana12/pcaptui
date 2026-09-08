// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package screenshot

import (
	"fmt"
	"strings"
)

//======================================================================

// AnimatedSVG draws a sequence of screens as one looping picture.
//
// SVG rather than a GIF because it is the same renderer the still screenshots
// already use, so the animation cannot drift away from them; because it stays
// sharp at any size, where a GIF of a terminal is either huge or blurry; and
// because it needs no encoder - there is no ffmpeg or ImageMagick to depend on
// and nothing to install to regenerate it.
//
// Frames are drawn on top of each other and shown one at a time by a discrete
// SMIL animation. Discrete rather than interpolated: a terminal does not fade
// between states, and interpolation would show two frames at once.
//
// A viewer that ignores SMIL - some renderers do - shows the first frame,
// which is why the first frame is the one worth seeing on its own.
func AnimatedSVG(frames [][][]Cell, w, h int, secondsPerFrame float64) string {
	if len(frames) == 0 {
		return ""
	}
	if len(frames) == 1 {
		return SVG(frames[0], w, h)
	}

	width := float64(w)*cellW + padding*2
	height := float64(h)*cellH + padding*2
	total := secondsPerFrame * float64(len(frames))

	var b strings.Builder

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" role="img">`,
		width, height, width, height)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <title>pcaptui</title>`+"\n")
	fmt.Fprintf(&b, `  <rect width="%.0f" height="%.0f" rx="10" fill="%s"/>`+"\n", width, height, defaultBg)
	fmt.Fprintf(&b, `  <g font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace" font-size="14">`)
	b.WriteByte('\n')

	for i, rows := range frames {
		fmt.Fprintf(&b, `    <g opacity="%d">`+"\n", boolToInt(i == 0))
		writeFrameAnimation(&b, i, len(frames), total)
		writeFrameBody(&b, rows, "      ")
		b.WriteString("    </g>\n")
	}

	b.WriteString("  </g>\n</svg>\n")
	return b.String()
}

// writeFrameAnimation makes frame i visible for its own slice of the loop.
func writeFrameAnimation(b *strings.Builder, i, n int, total float64) {
	from := float64(i) / float64(n)
	to := float64(i+1) / float64(n)

	// The first frame starts visible and is hidden when its turn ends; every
	// other starts hidden, appears, and hides again. The last frame's turn
	// runs to the end of the loop.
	switch {
	case i == 0:
		fmt.Fprintf(b, `      <animate attributeName="opacity" values="1;0" keyTimes="0;%.4f" dur="%.2fs" calcMode="discrete" repeatCount="indefinite"/>`+"\n",
			to, total)
	case i == n-1:
		fmt.Fprintf(b, `      <animate attributeName="opacity" values="0;1" keyTimes="0;%.4f" dur="%.2fs" calcMode="discrete" repeatCount="indefinite"/>`+"\n",
			from, total)
	default:
		fmt.Fprintf(b, `      <animate attributeName="opacity" values="0;1;0" keyTimes="0;%.4f;%.4f" dur="%.2fs" calcMode="discrete" repeatCount="indefinite"/>`+"\n",
			from, to, total)
	}
}

// writeFrameBody is the drawing half of SVG, without the document around it.
func writeFrameBody(b *strings.Builder, rows [][]Cell, indent string) {
	// Backgrounds first, so text is never painted under a later rectangle.
	for y, row := range rows {
		for _, run := range runs(row) {
			bg := background(run.style)
			if bg == defaultBg {
				continue
			}
			fmt.Fprintf(b, `%s<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s"/>`+"\n",
				indent,
				padding+float64(run.from)*cellW, padding+float64(y)*cellH,
				float64(run.to-run.from+1)*cellW, cellH, bg)
		}
	}

	for y, row := range rows {
		for _, run := range runs(row) {
			text := strings.TrimRight(run.text, " ")
			if text == "" {
				continue
			}
			fmt.Fprintf(b, `%s<text x="%.2f" y="%.2f" fill="%s"%s xml:space="preserve">%s</text>`+"\n",
				indent,
				padding+float64(run.from)*cellW, padding+float64(y)*cellH+cellH*0.72,
				foreground(run.style), weight(run.style), escape(text))
		}
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
