package extnet

import (
	"fmt"
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

// DialSCTP connects from the local address laddr
// to the remote address raddr.
func DialSCTP(laddr, raddr *SCTPAddr) (c *SCTPConn, e error) {
	if raddr == nil {
		return nil, &net.OpError{
			Op:     "dial",
			Net:    "sctp",
			Source: laddr,
			Addr:   raddr,
			Err:    fmt.Errorf("no remote address")}
	}
	return dial(&SCTPDialer{LocalAddr: laddr}, raddr)
}

// Dial connects to the addr.
func (d *SCTPDialer) Dial(n, addr string) (net.Conn, error) {
	switch n {
	case "sctp", "sctp4", "sctp6":
	default:
		return nil, &net.OpError{
			Op:     "dial",
			Net:    n,
			Source: d.LocalAddr,
			Addr:   nil,
			Err:    net.UnknownNetworkError(n)}
	}

	ra, e := ResolveSCTPAddr(n, addr)
	if e != nil {
		return nil, e
	}
	return dial(d, ra)
}

func dial(d *SCTPDialer, addr *SCTPAddr) (*SCTPConn, error) {
	l, e := listen(d)

	// connect SCTP connection to raddr
	ptr, n := addr.rawAddr()
	i, e := sctpConnectx(l.sock, ptr, n)
	if e != nil {
		e = &net.OpError{
			Op:     "connect",
			Net:    "sctp",
			Source: l.Addr(),
			Addr:   addr,
			Err:    e}
		l.Close()
		return nil, e
	}

	for {
		c, e := l.AcceptSCTP()
		if e != nil {
			l.Close()
			return nil, e
		}
		if c.id == i {
			l.close = make(chan bool)
			return c, nil
		}
		c.Abort("close")
	}
}

// ListenSCTP announces on the SCTP address laddr
// and returns a SCTP listener.
func ListenSCTP(n string, laddr *SCTPAddr) (*SCTPListener, error) {
	switch n {
	case "sctp", "sctp4", "sctp6":
	default:
		return nil, &net.OpError{
			Op:   "listen",
			Net:  n,
			Addr: laddr,
			Err:  net.UnknownNetworkError(n)}
	}

	return listen(&SCTPDialer{LocalAddr: laddr})
}

// Listen start listening.
func (d *SCTPDialer) Listen() (net.Listener, error) {
	return listen(d)
}

func listen(d *SCTPDialer) (*SCTPListener, error) {
	if d.LocalAddr == nil {
		return nil, &net.OpError{
			Op:   "listen",
			Net:  "sctp",
			Addr: nil,
			Err:  fmt.Errorf("no local address")}
	}

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
			Op:     "listen",
			Net:    "sctp",
			Source: d.LocalAddr,
			Addr:   nil,
			Err:    e}
	}

	// create listener
	l := &SCTPListener{
		sock:   sock,
		con:    make(map[assocT]*SCTPConn),
		accept: make(chan *SCTPConn, BacklogSize)}

	// start reading buffer
	r := make(chan bool)
	go read(l, r)
	<-r

	return l, nil
}

// bind SCTP socket
func (d *SCTPDialer) bindsocket(laddr *SCTPAddr) (int, error) {

	// create SCTP connection socket
	sock, e := sockOpen()
	if e != nil {
		e = &net.OpError{
			Op:   "makesock",
			Net:  "sctp",
			Addr: laddr,
			Err:  e}
		return -1, e
	}

	// set notifycation enabled
	e = setNotify(sock)
	if e != nil {
		sockClose(sock)
		e = &net.OpError{
			Op:   "setsockopt",
			Net:  "sctp",
			Addr: laddr,
			Err:  e}
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
		e = &net.OpError{
			Op:   "setsockopt",
			Net:  "sctp",
			Addr: laddr,
			Err:  e}
		return -1, e
	}

	// bind SCTP connection
	ptr, n := laddr.rawAddr()
	e = sctpBindx(sock, ptr, n)
	if e != nil {
		e = &net.OpError{
			Op:   "bindx",
			Net:  "sctp",
			Addr: laddr,
			Err:  e}
		sockClose(sock)
		return -1, e
	}
	return sock, nil
}
