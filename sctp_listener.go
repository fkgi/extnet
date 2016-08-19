package extnet

import (
	"errors"
	"io"
	"log"
	"net"
	"unsafe"
)

// SCTPListener is a SCTP network listener.
type SCTPListener struct {
	sock   int
	addr   net.Addr
	pipes  map[assocT]*io.PipeWriter
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

	for _, w := range l.pipes {
		w.CloseWithError(nil)
	}

	return sockClose(s)
}

// Addr returns the listener's network address, a *SCTPAddr.
func (l *SCTPListener) Addr() net.Addr {
	return l.addr
}

// ListenSCTP announces on the SCTP address laddr and returns a SCTP listener.
func ListenSCTP(laddr *SCTPAddr) (*SCTPListener, error) {
	sock, e := bindsocket(laddr)
	if e != nil {
		return nil, e
	}

	// start listen
	e = sockListen(sock)
	if e != nil {
		sockClose(sock)
		return nil, &net.OpError{Op: "listen", Net: "sctp", Source: laddr, Err: e}
	}

	// create listener
	l := &SCTPListener{}
	l.sock = sock
	l.addr = laddr
	l.pipes = make(map[assocT]*io.PipeWriter)
	l.accept = make(chan *SCTPConn, ListenBufferSize)

	// start reading buffer
	go read(l)

	return l, nil
}

// Connect create new connection of this listener
func (l *SCTPListener) Connect(addr net.Addr) error {
	if a, ok := addr.(*SCTPAddr); ok {
		return l.ConnectSCTP(a)
	}
	return &net.OpError{
		Op: "connect", Net: "sctp", Source: l.addr, Addr: addr,
		Err: errors.New("invalid Addr, not SCTPAddr")}
}

// ConnectSCTP create new connection of this listener
func (l *SCTPListener) ConnectSCTP(raddr *SCTPAddr) error {
	if l.sock == -1 {
		return &net.OpError{
			Op: "connect", Net: "sctp", Source: l.addr, Addr: raddr,
			Err: errors.New("socket is closed")}
	}

	// connect SCTP connection to raddr
	_, e := sctpConnectx(l.sock, raddr.rawAddr())
	return &net.OpError{Op: "connect", Net: "sctp", Source: l.addr, Addr: raddr, Err: e}
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
				log.Println("SCTP socket reading failed, socket will close: ", e.Error())
				l.Close()
			}
			break
		}

		// check message type is notify
		if flag&msgNotification == msgNotification {
			tlv := (*sctpTlv)(unsafe.Pointer(&buf[0]))

			e = error(nil)
			switch tlv.snType {
			case sctpAssocChange:
				e = l.assocChangeNotify(buf[:n])
			case sctpPeerAddrChange:
				log.Println("[peer addr change:", int(info.assocID))
			case sctpSendFailed:
				log.Println("[send failed:", int(info.assocID))
			case sctpRemoteError:
				log.Println("[remote error:", int(info.assocID))
			case sctpShutdownEvent:
				log.Println("[shutdown event:", int(info.assocID))
			case sctpPartialDeliveryEvent:
				log.Println("[partial delivery:", int(info.assocID))
			case sctpAdaptationIndication:
				log.Println("[adaptation indication:", int(info.assocID))
			case sctpSenderDryEvent:
				log.Println("[sender dry:", int(info.assocID))
			}

			if e != nil {
				log.Println(e)
			}
		} else {
			log.Println("[data:", int(info.assocID))
			// matching exist connection
			if p, ok := l.pipes[info.assocID]; ok {
				p.Write(buf[:n])
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

func (l *SCTPListener) assocChangeNotify(buf []byte) error {
	change := (*assocChange)(unsafe.Pointer(&buf[0]))

	switch change.state {
	case sctpCommUp:
		log.Println("[assoc change (comm up):", change.assocID)

		if _, ok := l.pipes[change.assocID]; ok {
			return errors.New("add association failed, assoc_id exist")
		}

		addr, e := sctpGetpaddrs(l.sock, int(change.assocID))
		if e != nil {
			return e
		}

		// create new connection
		c := &(SCTPConn{})
		c.l = l
		c.id = change.assocID
		c.r, l.pipes[change.assocID] = io.Pipe()
		c.addr = resolveFromRawAddr(addr)

		if l.accept != nil {
			l.accept <- c
		}
	case sctpCommLost, sctpShutdownComp:
		if change.state == sctpCommLost {
			log.Println("[assoc change (comm lost):", change.assocID)
		} else {
			log.Println("[assoc change (shutdown comp):", change.assocID)
		}

		w, ok := l.pipes[change.assocID]
		if !ok {
			return errors.New("remove association failed, assoc_id not found")
		}
		delete(l.pipes, change.assocID)
		w.Close()

		if change.state == sctpCommLost {
			return errors.New("association lost")
		}
	case sctpRestart:
		log.Println("[assoc change (reset):", change.assocID)
	case sctpCantStrAssoc:
		log.Println("[assoc change (cant str assoc):", change.assocID)
	}
	return nil
}

/*
	change := (*C.struct_sctp_paddr_change)(bufp)
	info.sinfo_assoc_id = change.spc_assoc_id
	switch change.spc_state {
	case C.SCTP_ADDR_AVAILABLE:
		// log.Println("[peer addr change(addr available):", int(change.spc_assoc_id))
	case C.SCTP_ADDR_UNREACHABLE:
		// log.Println("[peer addr change(addr unreachable):", int(change.spc_assoc_id))
	case C.SCTP_ADDR_REMOVED:
		// log.Println("[peer addr change(addr removed):", int(change.spc_assoc_id))
	case C.SCTP_ADDR_ADDED:
		// log.Println("[peer addr change(addr added):", int(change.spc_assoc_id))
	case C.SCTP_ADDR_MADE_PRIM:
		// log.Println("[peer addr change(addr made prim):", int(change.spc_assoc_id))
	case C.SCTP_ADDR_CONFIRMED:
		// log.Println("[peer addr change(addr confirmed):", int(change.spc_assoc_id))
	}

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

	change := (*C.struct_sctp_shutdown_event)(bufp)
	info.sinfo_assoc_id = change.sse_assoc_id
	// log.Println("[shutdown event:", int(change.sse_assoc_id))

	change := (*C.struct_sctp_pdapi_event)(bufp)
	info.sinfo_assoc_id = change.pdapi_assoc_id
	// log.Println("[partial delivery event:", int(change.pdapi_assoc_id))

	change := (*C.struct_sctp_adaptation_event)(bufp)
	info.sinfo_assoc_id = change.sai_assoc_id
	// log.Println("[adaptation indication:", int(change.sai_assoc_id))

	change := (*C.struct_sctp_sender_dry_event)(bufp)
	info.sinfo_assoc_id = change.sender_dry_assoc_id
	// log.Println("[sender dry event:", int(change.sender_dry_assoc_id))
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
