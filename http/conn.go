package http

import (
	"net"

	"github.com/curol/network/url"
)

//**********************************************************************************************************************
// Connection
//**********************************************************************************************************************

type parsedConnection struct {
	remoteAddress string
	localAddress  string
	url           *url.URL
	host          string
	hostname      string
	path          string
}

func parseConnection(conn net.Conn) (*parsedConnection, error) {
	pc := new(parsedConnection)

	pc.remoteAddress = conn.RemoteAddr().String()
	pc.localAddress = conn.LocalAddr().String()
	u, err := url.Parse(pc.remoteAddress)
	if err != nil {
		return nil, err
	}
	pc.url = u
	pc.hostname = u.Hostname()
	pc.host = u.Host
	return pc, nil
}
