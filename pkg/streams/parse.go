// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package streams

import (
	"encoding/hex"
	"fmt"
	"strings"
)

//======================================================================

type IChunk interface {
	Direction() Direction
	StreamData() []byte
}

type IOnStreamChunk interface {
	OnStreamChunk(chunk IChunk)
}

type IOnStreamHeader interface {
	OnStreamHeader(header FollowHeader)
}

//======================================================================

type parseContext interface {
	Err() error
}

type StreamParseError struct{}

func (e StreamParseError) Error() string {
	return "Stream reassembly parse error"
}

var _ error = StreamParseError{}

//======================================================================

type Protocol int

const (
	Unspecified Protocol = 0
	TCP         Protocol = iota
	UDP         Protocol = iota
	TLS         Protocol = iota
	WebSocket   Protocol = iota
	HTTP2       Protocol = iota
)

var _ fmt.Stringer = Protocol(0)

func (p Protocol) String() string {
	switch p {
	case Unspecified:
		return "Unspecified"
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	case TLS:
		return "TLS"
	case WebSocket:
		return "WebSocket"
	case HTTP2:
		return "HTTP/2"
	default:
		panic(fmt.Sprintf("unknown stream protocol %d", int(p)))
	}
}

//======================================================================

type Direction int

const (
	Client Direction = 0
	Server Direction = iota
)

func (d Direction) String() string {
	switch d {
	case Client:
		return "Client"
	case Server:
		return "Server"
	default:
		return "Unknown!"
	}
}

//======================================================================

type Bytes struct {
	Dirn Direction
	Data []byte
}

var _ fmt.Stringer = Bytes{}
var _ IChunk = Bytes{}

func (b Bytes) Direction() Direction {
	return b.Dirn
}

func (b Bytes) StreamData() []byte {
	return b.Data
}

func (b Bytes) String() string {
	return fmt.Sprintf("Direction: %v\n%s", b.Dirn, hex.Dump(b.Data))
}

//======================================================================

// Clean returns the header with the line endings taken off its four fields.
//
// tshark writes CRLF on Windows and the grammar captures a header line up to
// the newline, so every field arrives with a carriage return on the end there
// and without one everywhere else. Node0 and Node1 are drawn into the
// conversation menu and into the "client → server (N bytes)" line, where a
// stray CR lands in the middle of the text and sends the cursor back to the
// start of the row.
func (f FollowHeader) Clean() FollowHeader {
	trim := func(s string) string {
		return strings.TrimRight(s, "\r\n \t")
	}
	return FollowHeader{
		Follow: trim(f.Follow),
		Filter: trim(f.Filter),
		Node0:  trim(f.Node0),
		Node1:  trim(f.Node1),
	}
}

type FollowHeader struct {
	Follow string
	Filter string
	Node0  string
	Node1  string
}

func (h FollowHeader) String() string {
	return fmt.Sprintf("[client:%s server:%s follow:%s filter:%s]", h.Node0, h.Node1, h.Follow, h.Filter)
}

type FollowStream struct {
	FollowHeader
	Bytes []Bytes
}

var _ fmt.Stringer = FollowStream{}

func (f FollowStream) String() string {
	datastrs := make([]string, 0, len(f.Bytes))
	for _, b := range f.Bytes {
		datastrs = append(datastrs, b.String())
	}
	data := strings.Join(datastrs, "\n")
	return fmt.Sprintf("Follow: %s\nFilter: %s\nNode0: %s\nNode1: %s\nData:\n%s", f.Follow, f.Filter, f.Node0, f.Node1, data)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
