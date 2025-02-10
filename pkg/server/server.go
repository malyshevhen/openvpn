package server

import (
	"net"
	"time"
)

// MgmtListener accepts incoming connections from OpenVPN.
//
// The primary way to instantiate this type is via the function Listen.
// See its documentation for more information.
type MgmtListener interface {
	// Accept waits for and returns the next connection.
	Accept() (*IncomingConn, error)

	// Close closes the listener. Any blocked Accept operations
	// will be blocked and each will return an error.
	Close() error

	// Serve will await new connections and call the given handler
	// for each.
	//
	// Serve does not return unless the listen port is closed; a non-nil
	// error is always returned.
	Serve(handler IncomingConnHandler) error

	// Addr returns the listener's network address.
	Addr() net.Addr
}

// NewMgmtListener constructs a MgmtListener from an already-established
// net.Listener. In most cases it will be more convenient to use
// the function Listen.
func NewMgmtListener(l net.Listener) MgmtListener {
	return &mgmtListener{l}
}

// Listen opens a listen port and awaits incoming connections from OpenVPN
// processes.
//
// OpenVPN will behave in this manner when launched with the following options:
//
//	--management ipaddr port --management-client
//
// Note that in this case the terminology is slightly confusing, since from
// the standpoint of TCP/IP it is OpenVPN that is the client and our program
// that is the server, but once the connection is established the channel
// is indistinguishable from the situation where OpenVPN exposed a management
// *server* and we connected to it. Thus we still refer to our program as
// the "client" and OpenVPN as the "server" once the connection is established.
//
// When running on Unix systems it's possible to instead listen on a Unix
// domain socket. To do this, pass an absolute path to the socket as
// the listen address, and then run OpenVPN with the following options:
//
//	--management /path/to/socket unix --management-client
func Listen(laddr string) (MgmtListener, error) {
	proto := "tcp"
	if len(laddr) > 0 && laddr[0] == '/' {
		proto = "unix"
	}
	listener, err := net.Listen(proto, laddr)
	if err != nil {
		return nil, err
	}

	return NewMgmtListener(listener), nil
}

type mgmtListener struct {
	l net.Listener
}

func (l *mgmtListener) Accept() (*IncomingConn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	return &IncomingConn{conn}, nil
}

func (l *mgmtListener) Close() error {
	return l.l.Close()
}

func (l *mgmtListener) Addr() net.Addr {
	return l.l.Addr()
}

func (l *mgmtListener) Serve(handler IncomingConnHandler) error {
	defer l.Close()

	var tempDelay time.Duration

	for {
		incoming, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				// Wait a while before we try again.
				time.Sleep(tempDelay)
				continue
			} else {
				// Listen socket is permanently closed or errored,
				// so it's time for us to exit.
				return err
			}
		}

		// always reset our retry delay once we successfully read
		tempDelay = 0

		go handler.ServeOpenVPNMgmt(*incoming)
	}
}

type IncomingConnHandler interface {
	ServeOpenVPNMgmt(IncomingConn)
}

// IncomingConnHandlerFunc is an adapter to allow the use of ordinary
// functions as connection handlers.
//
// Given a function with the appropriate signature, IncomingConnHandlerFunc(f)
// is an IncomingConnHandler that calls f.
type IncomingConnHandlerFunc func(IncomingConn)

func (f IncomingConnHandlerFunc) ServeOpenVPNMgmt(i IncomingConn) {
	f(i)
}

// ListenAndServe creates a MgmtListener for the given listen address
// and then calls AcceptAndServe on it.
//
// This is just a convenience wrapper. See the AcceptAndServe method for
// more details. Just as with AcceptAndServe, this function does not return
// except on error; in addition to the error cases handled by AcceptAndServe,
// this function may also fail if the listen socket cannot be established
// in the first place.
func ListenAndServe(laddr string, handler IncomingConnHandler) error {
	listener, err := Listen(laddr)
	if err != nil {
		return err
	}

	return listener.Serve(handler)
}
