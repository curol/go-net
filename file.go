// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"net"
)

// BUG(mikio): On JS and Windows, the FileConn, FileListener and
// FilePacketConn functions are not implemented.

type fileAddr string

func (fileAddr) Network() string  { return "file+net" }
func (f fileAddr) String() string { return string(f) }

// FileConn returns a copy of the network connection
// corresponding to the open file f. It is the caller's
// responsibility to close f when finished. Closing c does
// not affect f, and closing f does not affect c.
var FileConn = net.FileConn

// FileListener returns a copy of the network listener corresponding
// to the open file f.
// It is the caller's responsibility to close ln when finished.
// Closing ln does not affect f, and closing f does not affect ln.
var FileListener = net.FileListener

// FilePacketConn returns a copy of the packet network connection
// corresponding to the open file f.
// It is the caller's responsibility to close f when finished.
// Closing c does not affect f, and closing f does not affect c.
var FilePacketConn = net.FilePacketConn
