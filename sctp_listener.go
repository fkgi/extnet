package extnet

import (
	"errors"
	"io"
	"net"
	"unsafe"
)

// SCTPListener is a SCTP network listener.
type SCTPListener struct {
	sock   int
	con    map[assocT]*SCTPConn
	accept chan *SCTPConn
}

// Accept implements the Accept method in the Listener interface;
// it waits for the next call and returns a generic Conn.
func (l *SCTPListener) Accept() (net.Conn, error) {
	return l.AcceptSCTP()
}

// AcceptSCTP accepts the next incoming call and returns the new connection.
func (l *SCTPListener) AcceptSCTP() (c *SCTPConn, e error) {
	c = <-l.accept
	if c == nil {
		e = &net.OpError{
			Op: "accept", Net: "sctp", Addr: l.Addr(),
			Err: errors.New("socket is closed")}
	}
	return
}

// Close stops listening on the SCTP address.
func (l *SCTPListener) Close() (e error) {
	s := l.sock
	if s == -1 {
		return
	}
	l.sock = -1

	t := l.accept
	l.accept = nil
	close(t)

	for _, c := range l.con {
		c.queue(nil, io.EOF)
	}

	return sockClose(s)
}

// Addr returns the listener's network address, a *SCTPAddr.
func (l *SCTPListener) Addr() net.Addr {
	ptr, n, e := sctpGetladdrs(l.sock, 0)
	if e != nil {
		return nil
	}
	defer sctpFreeladdrs(ptr)

	return resolveFromRawAddr(ptr, n)
}

// ListenSCTP announces on the SCTP address laddr and returns a SCTP listener.
func ListenSCTP(laddr *SCTPAddr) (*SCTPListener, error) {
	// bind local address
	sock, e := bindsocket(laddr)
	if e != nil {
		return nil, e
	}

	// start listen
	e = sockListen(sock)
	if e != nil {
		sockClose(sock)
		return nil, &net.OpError{
			Op: "listen", Net: "sctp", Source: laddr, Err: e}
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

// Connect create new connection of this listener
func (l *SCTPListener) Connect(addr net.Addr) error {
	if a, ok := addr.(*SCTPAddr); ok {
		return l.ConnectSCTP(a)
	}
	laddr := l.Addr()
	return &net.OpError{
		Op: "connect", Net: "sctp", Source: laddr, Addr: addr,
		Err: errors.New("invalid Addr, not SCTPAddr")}
}

// ConnectSCTP create new connection of this listener
func (l *SCTPListener) ConnectSCTP(raddr *SCTPAddr) error {
	if l.sock == -1 {
		return &net.OpError{
			Op: "connect", Net: "sctp", Source: l.Addr(), Addr: raddr,
			Err: errors.New("socket is closed")}
	}

	// connect SCTP connection to raddr
	ptr, n := raddr.rawAddr()
	_, e := sctpConnectx(l.sock, ptr, n)
	if e != nil {
		return &net.OpError{
			Op: "connect", Net: "sctp", Source: l.Addr(), Addr: raddr, Err: e}
	}
	return nil
}

type sctpTlv struct {
	snType   uint16
	snFlags  uint16
	snLength uint32
}

// read data from buffer
func read(l *SCTPListener) {
	buf := make([]byte, RxBufferSize)
	info := sndrcvInfo{}
	flag := 0

	for {
		// receive message
		n, e := sctpRecvmsg(l.sock, buf, &info, &flag)
		if e != nil {
			if l.accept != nil {
				l.Close()
			}
			break
		}

		// check message type is notify
		if flag&msgNotification == msgNotification {
			tlv := (*sctpTlv)(unsafe.Pointer(&buf[0]))

			switch tlv.snType {
			case sctpAssocChange:
				l.assocChangeNotify(buf[:n])
			case sctpPeerAddrChange:
				l.paddrChangeNotify(buf[:n])
			case sctpSendFailed:
				tracePrint("send failed", info.assocID)
			case sctpRemoteError:
				tracePrint("remote error", info.assocID)
			case sctpShutdownEvent:
				tracePrint("shutdown event", info.assocID)
			case sctpPartialDeliveryEvent:
				tracePrint("partial delivery", info.assocID)
			case sctpAdaptationIndication:
				tracePrint("adaptation indication", info.assocID)
			case sctpSenderDryEvent:
				tracePrint("sender dry", info.assocID)
			}
		} else {
			tracePrint("data", info.assocID)
			// matching exist connection
			if p, ok := l.con[info.assocID]; ok {
				p.queue(buf[:n], nil)
			} else {
				errorPrint("available connection not found", info.assocID)
			}
		}
	}
}

type assocChange struct {
	chtype          uint16
	flags           uint16
	length          uint32
	state           uint16
	sacError        uint16
	outboundStreams uint16
	inboundStreams  uint16
	assocID         assocT
}

func (l *SCTPListener) assocChangeNotify(buf []byte) {
	c := (*assocChange)(unsafe.Pointer(&buf[0]))

	switch c.state {
	case sctpCommUp:
		tracePrint("assoc change (comm up)", c.assocID)

		if _, ok := l.con[c.assocID]; ok {
			errorPrint("overwrite duplicate assoc id", c.assocID)
		}

		// create new connection
		con := &(SCTPConn{})
		con.l = l
		con.id = c.assocID

		con.buf = make([]byte, 0, RxBufferSize)
		con.win = con.buf
		con.wc.L = &con.m

		l.con[c.assocID] = con

		if l.accept != nil {
			l.accept <- con
		}
	case sctpCommLost, sctpShutdownComp:
		if c.state == sctpCommLost {
			tracePrint("assoc change (comm lost)", c.assocID)
		} else {
			tracePrint("assoc change (shutdown comp)", c.assocID)
		}

		if con, ok := l.con[c.assocID]; ok {
			delete(l.con, c.assocID)
			con.queue(nil, io.EOF)
		} else {
			errorPrint("assoc id not found", c.assocID)
			return
		}
	case sctpRestart:
		tracePrint("assoc change (reset)", c.assocID)
	case sctpCantStrAssoc:
		tracePrint("assoc change (cant str assoc)", c.assocID)
	}
	return
}

type paddrChange struct {
	chtype   uint16
	flags    uint16
	length   uint32
	addr     sockaddrStorage
	state    int
	spcError int
	assocID  assocT
}

type sockaddrStorage struct {
	length uint8
	family uint8
	pad1   [6]byte
	align  int64
	pad2   [112]byte
}

func (l *SCTPListener) paddrChangeNotify(buf []byte) {
	c := (*paddrChange)(unsafe.Pointer(&buf[0]))
	switch c.state {
	case sctpAddrAvailable:
		tracePrint("peer addr change(addr available):", c.assocID)
	case sctpAddrUnreachable:
		tracePrint("peer addr change(addr unreachable):", c.assocID)
	case sctpAddrRemoved:
		tracePrint("peer addr change(addr removed):", c.assocID)
	case sctpAddrAdded:
		tracePrint("peer addr change(addr added):", c.assocID)
	case sctpAddrMadePrim:
		tracePrint("peer addr change(addr made prim):", c.assocID)
	case sctpAddrConfirmed:
		tracePrint("peer addr change(addr confirmed):", c.assocID)
	}
}

/*
	change := (*C.struct_sctp_send_failed)(bufp)
	info.sinfo_assoc_id = change.ssf_assoc_id
	switch change.ssf_flags {
	case C.SCTP_DATA_UNSENT:
		// log.Println("[send failed(data unset):", int(change.ssf_assoc_id))
	case C.SCTP_DATA_SENT:
		// log.Println("[send failed(data sent):", int(change.ssf_assoc_id))
	}

	change := (*C.struct_sctp_remote_error)(bufp)
	info.sinfo_assoc_id = change.sre_assoc_id
	// log.Println("[remote error(", change.sre_error, ")", change.sre_assoc_id)
*/

type initmsg struct {
	oStr      uint16
	iStr      uint16
	attempts  uint16
	initTimeo uint16
}

// SetInitmsg set init message parameter.
// oStr is number of output stream.
// iStr is maximum numnner of input stream.
// attempts is maximum number of attempt count.
// initTimeo is time of timeout.
func (l *SCTPListener) SetInitmsg(oStr, iStr, attempts, initTimeo int) error {
	attr := initmsg{
		oStr:      uint16(oStr),
		iStr:      uint16(iStr),
		attempts:  uint16(attempts),
		initTimeo: uint16(initTimeo)}
	lng := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)
	return setSockOpt(l.sock, sctpInitMsg, p, lng)
}
