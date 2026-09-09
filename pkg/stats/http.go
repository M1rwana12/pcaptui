// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"regexp"
	"strconv"
)

//======================================================================

// HTTPFilter is the display filter that shows the packets a row counted, or
// "" when there is no honest one.
//
// The rows are tshark's, and their names are the whole of what it says about
// them, so the filter is derived from the name. Only the rows whose meaning is
// unambiguous get one; "Other HTTP Packets" is tshark's own catch-all and no
// filter expresses it, so that row is shown and not offered.
func HTTPFilter(name string) string {
	switch name {
	case "Total HTTP Packets":
		return "http"
	case "HTTP Response Packets":
		return "http.response"
	case "HTTP Request Packets":
		return "http.request"
	}

	// "404 Not Found", "500 Internal Server Error".
	if m := httpCode.FindStringSubmatch(name); m != nil {
		return fmt.Sprintf("http.response.code == %s", m[1])
	}

	// "4xx: Client Error" - the class, which is a range rather than a value.
	if m := httpClass.FindStringSubmatch(name); m != nil {
		lo, err := strconv.Atoi(m[1])
		if err != nil {
			return ""
		}
		lo *= 100
		return fmt.Sprintf("http.response.code >= %d && http.response.code < %d", lo, lo+100)
	}

	// A request method, if tshark ever lists them here.
	if httpMethod.MatchString(name) {
		return fmt.Sprintf("http.request.method == %q", name)
	}

	return ""
}

var (
	httpCode  = regexp.MustCompile(`^([1-5][0-9]{2})\b`)
	httpClass = regexp.MustCompile(`^([1-5])xx\b`)
	// Upper-case throughout, which is what separates a method from a
	// description like "Broken" or a class like "5xx".
	httpMethod = regexp.MustCompile(`^[A-Z]{3,}$`)
)

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
