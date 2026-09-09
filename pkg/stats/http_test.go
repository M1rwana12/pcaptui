// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestAStatusCodeRowFiltersOnThatCode(t *testing.T) {
	assert.Equal(t, "http.response.code == 404", HTTPFilter("404 Not Found"))
	assert.Equal(t, "http.response.code == 500", HTTPFilter("500 Internal Server Error"))
	assert.Equal(t, "http.response.code == 200", HTTPFilter("200 OK"))
}

// A class is a range, not a value: 4xx is every code from 400 to 499.
func TestAStatusClassRowFiltersOnTheRange(t *testing.T) {
	assert.Equal(t,
		"http.response.code >= 400 && http.response.code < 500",
		HTTPFilter("4xx: Client Error"))
	assert.Equal(t,
		"http.response.code >= 100 && http.response.code < 200",
		HTTPFilter("1xx: Informational"))
}

func TestTheTotalsRowsFilterOnWhatTheyCount(t *testing.T) {
	assert.Equal(t, "http", HTTPFilter("Total HTTP Packets"))
	assert.Equal(t, "http.response", HTTPFilter("HTTP Response Packets"))
	assert.Equal(t, "http.request", HTTPFilter("HTTP Request Packets"))
}

// tshark's own catch-alls. Showing a row is right; pretending a filter
// expresses it is not, so these get none and the row is not selectable.
func TestARowWithNoHonestFilterGetsNone(t *testing.T) {
	assert.Equal(t, "", HTTPFilter("Other HTTP Packets"))
	assert.Equal(t, "", HTTPFilter("???: broken"))
	assert.Equal(t, "", HTTPFilter(""))
}

// 600 is not a status class, and a row that merely starts with digits is not
// a status code.
func TestOnlyRealStatusCodesBecomeCodeFilters(t *testing.T) {
	assert.Equal(t, "", HTTPFilter("600 Nonsense"))
	assert.Equal(t, "", HTTPFilter("6xx: Nonsense"))
	assert.Equal(t, "", HTTPFilter("42 packets"))
}

func TestARequestMethodFiltersOnTheMethod(t *testing.T) {
	assert.Equal(t, `http.request.method == "GET"`, HTTPFilter("GET"))
	assert.Equal(t, `http.request.method == "POST"`, HTTPFilter("POST"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
