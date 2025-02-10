package server

import (
	"net"

	"github.com/malyshevhen/openvpn/pkg/client"
	"github.com/malyshevhen/openvpn/pkg/events"
)

type IncomingConn struct {
	conn net.Conn
}

// Open initiates communication with the connected OpenVPN process,
// and establishes the channel on which events will be delivered.
//
// See the documentation for NewClient for discussion about the requirements
// for eventCh.
func (ic IncomingConn) Open(eventCh chan<- events.Event) client.MgmtClient {
	return client.NewClient(ic.conn, eventCh)
}

// Close abruptly closes the socket connected to the OpenVPN process.
//
// This is a rather abrasive way to close the channel, intended for rejecting
// unwanted incoming clients that may or may not speak the OpenVPN protocol.
//
// Once communication is accepted and established, it is generally better
// to close the connection gracefully using commands on the client returned
// from Open.
func (ic IncomingConn) Close() error {
	return ic.conn.Close()
}
