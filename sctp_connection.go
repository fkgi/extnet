package extnet

import (
	"io"
	"net"
	"time"
	"unsafe"
)

// SCTPConn is an implementation of the Conn interface for SCTP network connections.
type SCTPConn struct {
	l      *SCTPListener
	id     assocT
	r      *io.PipeReader
	addr   net.Addr
	wdline time.Time
	rdline *time.Timer
}

func (c *SCTPConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *SCTPConn) Write(b []byte) (int, error) {
	return c.send(b, 0)
}

// Close closes the connection.
func (c *SCTPConn) Close() error {
	_, e := c.send([]byte{}, sctpEoF)
	return e
}

// Abort closes the connection with abort message.
func (c *SCTPConn) Abort(reason string) error {
	_, e := c.send([]byte(reason), sctpAbort)
	return e
}

func (c *SCTPConn) send(b []byte, flag uint16) (int, error) {
	info := sndrcvInfo{}
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

// SetDeadline implements the Conn SetDeadline method.
func (c *SCTPConn) SetDeadline(t time.Time) (e error) {
	e = c.SetReadDeadline(t)
	if e != nil {
		return
	}
	e = c.SetWriteDeadline(t)
	return
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *SCTPConn) SetReadDeadline(t time.Time) error {
	c.rdline = time.AfterFunc(t.Sub(time.Now()), func() {
		c.l.pipes[c.id].Write(nil)
	})
	return nil
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *SCTPConn) SetWriteDeadline(t time.Time) error {
	c.wdline = t
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
	l.pipes = make(map[assocT]*io.PipeWriter)
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

type rtoinfo struct {
	assocID assocT
	ini     uint32
	max     uint32
	min     uint32
}

// SetRtoInfo set retransmit timer options
func (c *SCTPConn) SetRtoInfo(ini, min, max int) error {
	attr := rtoinfo{
		assocID: c.id,
		ini:     uint32(ini),
		max:     uint32(max),
		min:     uint32(min)}
	l := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)

	return setSockOpt(c.l.sock, sctpRtoInfo, p, l)
}

type assocparams struct {
	assocID     assocT
	pRwnd       uint32
	lRwnd       uint32
	cLife       uint32
	assocMaxRxt uint16
	numPeerDest uint16
}

// SetAssocinfo set association parameter
func (c *SCTPConn) SetAssocinfo(pRwnd, lRwnd, cLife, assocMaxRxt, numPeerDest int) error {
	attr := assocparams{
		assocID:     c.id,
		pRwnd:       uint32(pRwnd),
		lRwnd:       uint32(lRwnd),
		cLife:       uint32(cLife),
		assocMaxRxt: uint16(assocMaxRxt),
		numPeerDest: uint16(numPeerDest)}
	l := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)

	return setSockOpt(c.l.sock, sctpAssocInfo, p, l)
}

// SetNodelay set delay answer or not
func (c *SCTPConn) SetNodelay(attr bool) error {
	l := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)

	return setSockOpt(c.l.sock, sctpNodelay, p, l)
}

/*
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
