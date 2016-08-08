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
	pipes  map[int32]*io.PipeWriter
	accept chan *SCTPConn
}

// Accept implements the Accept method in the Listener interface; it waits for the next call and returns a generic Conn.
func (l *SCTPListener) Accept() (net.Conn, error) {
	return l.AcceptSCTP()
}

// AcceptSCTP accepts the next incoming call and returns the new connection.
func (l *SCTPListener) AcceptSCTP() (c *SCTPConn, e error) {
	c = <-l.accept
	if c == nil {
		e = &net.OpError{Op: "accept", Net: "sctp", Addr: l.Addr(), Err: errors.New("socket is closed")}
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
		return nil, &net.OpError{Op: "listen", Net: "sctp", Addr: nil, Err: e}
	}

	// create listener
	l := &SCTPListener{}
	l.sock = sock
	l.addr = laddr
	l.pipes = make(map[int32]*io.PipeWriter)
	l.accept = make(chan *SCTPConn, ListenBufferSize)

	// start reading buffer
	go read(l)

	return l, nil
}

// Connect create new connection of this listener
func (l *SCTPListener) Connect(addr net.Addr) error {
	a, ok := addr.(*SCTPAddr)
	if ok {
		return l.ConnectSCTP(a)
	}
	return errors.New("invalid Addr, not SCTPAddr")
}

// ConnectSCTP create new connection of this listener
func (l *SCTPListener) ConnectSCTP(raddr *SCTPAddr) error {
	if l.sock == -1 {
		return errors.New("socket is closed")
	}

	// connect SCTP connection to raddr
	_, e := sctpConnectx(l.sock, raddr.rawAddr())
	return e
}

type sctpTlv struct {
	snType   uint16
	snFlags  uint16
	snLength uint32
}

// read data from buffer
func read(l *SCTPListener) {
	buf := make([]byte, RxBufferSize)
	info := sndrcvinfo{}
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
		if flag&MSG_NOTIFICATION == MSG_NOTIFICATION {
			tlv := (*sctpTlv)(unsafe.Pointer(&buf[0]))

			e = error(nil)
			switch tlv.snType {
			case SCTP_ASSOC_CHANGE:
				e = l.assocChangeNotify(buf[:n])
			case SCTP_PEER_ADDR_CHANGE:
				log.Println("[peer addr change:", int(info.assocID))
			case SCTP_SEND_FAILED:
				log.Println("[send failed:", int(info.assocID))
			case SCTP_REMOTE_ERROR:
				log.Println("[remote error:", int(info.assocID))
			case SCTP_SHUTDOWN_EVENT:
				log.Println("[shutdown event:", int(info.assocID))
			case SCTP_PARTIAL_DELIVERY_EVENT:
				log.Println("[partial delivery:", int(info.assocID))
			case SCTP_ADAPTATION_INDICATION:
				log.Println("[adaptation indication:", int(info.assocID))
			case SCTP_SENDER_DRY_EVENT:
				log.Println("[sender dry:", int(info.assocID))
			}
			if e != nil {
				log.Println(e)
			}
		} else {
			log.Println("[data:", int(info.assocID))
			// matching exist connection
			p, ok := l.pipes[info.assocID]
			if ok {
				p.Write(buf[:n])
			}
		}
	}
}

type sctpAssocChange struct {
	chtype          uint16
	flags           uint16
	length          uint32
	state           uint16
	sacError        uint16
	outboundStreams uint16
	inboundStreams  uint16
	assocID         int32
}

func (l *SCTPListener) assocChangeNotify(buf []byte) error {
	change := (*sctpAssocChange)(unsafe.Pointer(&buf[0]))

	switch change.state {
	case SCTP_COMM_UP:
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
	case SCTP_COMM_LOST, SCTP_SHUTDOWN_COMP:
		if change.state == SCTP_COMM_LOST {
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

		if change.state == SCTP_COMM_LOST {
			return errors.New("association lost")
		}
	case SCTP_RESTART:
		log.Println("[assoc change (reset):", change.assocID)
	case SCTP_CANT_STR_ASSOC:
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

/*

func (c *SCTPConn) SetRtoinfo(initial, max, min uint32) error {
	attr := C.struct_sctp_rtoinfo{}
	l := C.socklen_t(unsafe.Sizeof(attr))

	attr.srto_assoc_id = c.id
	attr.srto_initial = C.__u32(initial)
	attr.srto_max = C.__u32(max)
	attr.srto_min = C.__u32(min)

	p := unsafe.Pointer(&attr)
	i, e := C.setsockopt(c.sock, C.SOL_SCTP, C.SCTP_RTOINFO, p, l)
	if int(i) < 0 {
		return e
	}
	return nil
}

func (c *SCTPConn) SetAssocinfo(
	asocmaxrxt, number_peer_destinations uint16, peer_rwnd, local_rwnd, cookie_life uint32) error {
	attr := C.struct_sctp_assocparams{}
	l := C.socklen_t(unsafe.Sizeof(attr))

	attr.sasoc_assoc_id = c.id
	attr.sasoc_asocmaxrxt = C.__u16(asocmaxrxt)
	attr.sasoc_number_peer_destinations = C.__u16(number_peer_destinations)
	attr.sasoc_peer_rwnd = C.__u32(peer_rwnd)
	attr.sasoc_local_rwnd = C.__u32(local_rwnd)
	attr.sasoc_cookie_life = C.__u32(cookie_life)

	p := unsafe.Pointer(&attr)
	i, e := C.setsockopt(c.sock, C.SOL_SCTP, C.SCTP_ASSOCINFO, p, l)
	if int(i) < 0 {
		return e
	}
	return nil
}

func (lnr *SCTPListener) SetInitmsg(num_ostreams, instreams, attempts, init_timeo uint16) error {
	attr := C.struct_sctp_initmsg{}
	l := C.socklen_t(unsafe.Sizeof(attr))

	attr.sinit_num_ostreams = C.__u16(num_ostreams)
	attr.sinit_max_instreams = C.__u16(instreams)
	attr.sinit_max_attempts = C.__u16(attempts)
	attr.sinit_max_init_timeo = C.__u16(init_timeo)

	p := unsafe.Pointer(&attr)
	i, e := C.setsockopt(lnr.sock, C.SOL_SCTP, C.SCTP_INITMSG, p, l)
	if int(i) < 0 {
		return e
	}
	return nil
}

func (c *SCTPConn) SetNodelay(nodelay bool) error {
	attr := nodelay
	l := C.socklen_t(unsafe.Sizeof(attr))
	p := unsafe.Pointer(&attr)
	i, e := C.setsockopt(c.sock, C.SOL_SCTP, C.SCTP_NODELAY, p, l)
	if int(i) < 0 {
		return e
	}
	return nil
}

const (
	HB_ENABLE         = uint32(C.SPP_HB_ENABLE)
	HB_DISABLE        = uint32(C.SPP_HB_DISABLE)
	HB_DEMAND         = uint32(C.SPP_HB_DEMAND)
	PMTUD_ENABLE      = uint32(C.SPP_PMTUD_ENABLE)
	PMTUD_DISABLE     = uint32(C.SPP_PMTUD_DISABLE)
	SACKDELAY_ENABLE  = uint32(C.SPP_SACKDELAY_ENABLE)
	SACKDELAY_DISABLE = uint32(C.SPP_SACKDELAY_DISABLE)
	HB_TIME_IS_ZERO   = uint32(C.SPP_HB_TIME_IS_ZERO)
)

func (c *SCTPConn) SetPeerAddrParams(
	hbinterval uint32, pathmaxrxt uint16, pathmtu, sackdelay, flags uint32) error {
	attr := C.struct_sctp_paddrparams{}
	l := C.socklen_t(unsafe.Sizeof(attr))

	attr.spp_assoc_id = c.id
	attr.spp_hbinterval = C.__u32(hbinterval)
	attr.spp_pathmaxrxt = C.__u16(pathmaxrxt)
	//attr.spp_pathmtu = C.__u32(pathmtu)
	//attr.spp_sackdelay = C.__u32(sackdelay)
	//attr.spp_flags = C.__u32(flags)

	p := unsafe.Pointer(&attr)
	i, e := C.setsockopt(c.sock, C.SOL_SCTP, C.SCTP_PEER_ADDR_PARAMS, p, l)
	if int(i) < 0 {
		return e
	}
	return nil
}
*/
