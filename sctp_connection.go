package extnet

import (
	"io"
	"net"
	"time"
)

// SCTPConn is an implementation of the Conn interface for SCTP network connections.
type SCTPConn struct {
	l    *SCTPListener
	id   int32
	r    *io.PipeReader
	addr net.Addr
}

func (c *SCTPConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *SCTPConn) Write(b []byte) (int, error) {
	return c.send(b, 0)
}

// Close closes the connection.
func (c *SCTPConn) Close() error {
	_, e := c.send([]byte{}, SCTP_EOF)
	return e
}

// Abort closes the connection with abort message.
func (c *SCTPConn) Abort(reason string) error {
	_, e := c.send([]byte(reason), SCTP_ABORT)
	return e
}

func (c *SCTPConn) send(b []byte, flag uint16) (int, error) {
	info := sndrcvinfo{}
	info.flags = flag
	info.assocID = c.id
	return sctpSend(c.l.sock, b, &info, 0)
}

// LocalAddr returns the local network address.
func (c *SCTPConn) LocalAddr() net.Addr {
	return c.l.addr
}

// RemoteAddr returns the remote network address.
func (c *SCTPConn) RemoteAddr() net.Addr {
	return c.addr
}

// SetDeadline implements the Conn SetDeadline method. *not implemented yet
func (c *SCTPConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements the Conn SetReadDeadline method. *not implemented yet
func (c *SCTPConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements the Conn SetWriteDeadline method. *not implemented yet
func (c *SCTPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// DialSCTP connects from the local address laddr to the remote address raddr.
func DialSCTP(laddr, raddr *SCTPAddr) (c *SCTPConn, e error) {
	sock, e := bindsocket(laddr)
	if e != nil {
		return nil, e
	}

	// create listener
	l := &SCTPListener{}
	l.sock = sock
	l.addr = laddr
	l.pipes = make(map[int32]*io.PipeWriter)
	l.accept = make(chan *SCTPConn, 1)

	// start reading buffer
	go read(l)

	e = l.ConnectSCTP(raddr)
	if e != nil {
		return nil, e
	}
	c = <-l.accept
	close(l.accept)
	l.accept = nil

	return
}

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
