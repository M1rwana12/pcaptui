// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

//======================================================================

// TwoPassFlag is pcaptui's spelling of tshark's -2. It has its own name
// because it is not simply forwarded: pcaptui applies it to a file and not to
// a live capture, and says so when asked for the impossible one.
const TwoPassFlag = "--two-pass"

// TwoPassArg is what tshark calls it.
const TwoPassArg = "-2"

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
