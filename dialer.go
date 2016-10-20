package extnet

import (
	"net"
	"time"
	"unsafe"
)

// SCTPDialer contains options for connecting to an address.
type SCTPDialer struct {
	LocalAddr *SCTPAddr

	OutStream   uint
	InStream    uint
	MaxAttempts uint
	InitTimeout time.Duration
}

// DialSCTP connects from the local address laddr to the remote address raddr.
func DialSCTP(laddr, raddr *SCTPAddr) (c *SCTPConn, e error) {
	d := &SCTPDialer{}
	d.LocalAddr = laddr

	return d.Dial(raddr)
}

// Dial connects to the addr.
func (d *SCTPDialer) Dial(addr *SCTPAddr) (*SCTPConn, error) {
	sock, e := d.bindsocket(d.LocalAddr)
	if e != nil {
		return nil, e
	}

	// create listener
	l := &SCTPListener{}
	l.sock = sock
	l.con = make(map[assocT]*SCTPConn)
	l.accept = make(chan *SCTPConn, 1)

	// start reading buffer
	go read(l)

	e = l.ConnectSCTP(addr)
	if e != nil {
		return nil, e
	}
	c := <-l.accept
	close(l.accept)
	l.accept = nil

	return c, nil
}

// ListenSCTP announces on the SCTP address laddr and returns a SCTP listener.
func ListenSCTP(laddr *SCTPAddr) (*SCTPListener, error) {
	d := &SCTPDialer{}
	d.LocalAddr = laddr

	return d.Listen()
}

// Listen start listening.
func (d *SCTPDialer) Listen() (*SCTPListener, error) {
	// bind local address
	sock, e := d.bindsocket(d.LocalAddr)
	if e != nil {
		return nil, e
	}

	// start listen
	e = sockListen(sock)
	if e != nil {
		sockClose(sock)
		return nil, &net.OpError{
			Op: "listen", Net: "sctp", Source: d.LocalAddr, Err: e}
	}

	// create listener
	l := &SCTPListener{}
	l.sock = sock
	l.con = make(map[assocT]*SCTPConn)
	l.accept = make(chan *SCTPConn, BacklogSize)

	// start reading buffer
	go read(l)

	return l, nil
}

// bind SCTP socket
func (d *SCTPDialer) bindsocket(laddr *SCTPAddr) (int, error) {
	// create SCTP connection socket
	sock, e := sockOpen()
	if e != nil {
		e = &net.OpError{Op: "makesock", Net: "sctp", Addr: nil, Err: e}
		return -1, e
	}

	// set notifycation enabled
	e = setNotify(sock)
	if e != nil {
		sockClose(sock)
		e = &net.OpError{Op: "setsockopt", Net: "sctp", Addr: nil, Err: e}
		return -1, e
	}

	// set init parameter
	type opt struct {
		o uint16
		i uint16
		a uint16
		t uint16
	}

	attr := opt{
		o: uint16(d.OutStream),
		i: uint16(d.InStream),
		a: uint16(d.MaxAttempts),
		t: uint16(d.InitTimeout)}
	l := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)
	e = setSockOpt(sock, sctpInitMsg, p, l)
	if e != nil {
		sockClose(sock)
		e = &net.OpError{Op: "setsockopt", Net: "sctp", Addr: nil, Err: e}
		return -1, e
	}

	// bind SCTP connection
	ptr, n := laddr.rawAddr()
	e = sctpBindx(sock, ptr, n)
	if e != nil {
		e = &net.OpError{Op: "bindx", Net: "sctp", Addr: nil, Err: e}
		sockClose(sock)
		return -1, e
	}
	return sock, nil
}
