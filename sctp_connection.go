package extnet

import (
	"io"
	"net"
	"sync"
	"time"
	"unsafe"
)

// SCTPConn is an implementation of the Conn interface for SCTP network connections.
type SCTPConn struct {
	l  *SCTPListener
	id assocT

	buf, win []byte
	err      error

	m, rm, wm sync.Mutex
	wc        sync.Cond

	wd, rd time.Time
}

func (c *SCTPConn) Read(b []byte) (n int, e error) {
	c.rm.Lock()
	defer c.rm.Unlock()
	c.m.Lock()
	defer c.m.Unlock()

	var t *time.Timer
	if now := time.Now(); !c.rd.IsZero() && now.Before(c.rd) {
		t = time.AfterFunc(c.rd.Sub(now), func() {
			er := &net.OpError{
				Op:     "read",
				Net:    "sctp",
				Source: c.LocalAddr(),
				Addr:   c.RemoteAddr(),
				Err:    &timeoutError{}}
			c.queue(nil, er)
		})
	}

	for {
		if len(c.win) != 0 {
			n = copy(b, c.win)
			c.win = c.win[n:]
			break
		}
		if c.err != nil {
			e = c.err
			if e != io.EOF {
				c.err = nil
			}
			break
		}
		c.wc.Wait()
	}

	if t != nil {
		t.Stop()
	}
	return
}

func (c *SCTPConn) queue(b []byte, e error) error {
	if c.err == io.EOF {
		return c.err
	}

	c.wm.Lock()
	defer c.wm.Unlock()

	if b != nil {
		c.m.Lock()
		defer c.m.Unlock()

		if len(c.win) == 0 {
			c.win = append(c.buf, b...)
		} else {
			c.win = append(c.win, b...)
		}
		c.wc.Signal()
	}

	if e != nil {
		c.m.Lock()
		defer c.m.Unlock()

		c.err = e
		c.wc.Signal()
	}
	return nil
}

func (c *SCTPConn) Write(b []byte) (int, error) {
	n, e := c.send(b, 0)
	if e != nil {
		e = &net.OpError{
			Op:     "write",
			Net:    "sctp",
			Source: c.LocalAddr(),
			Addr:   c.RemoteAddr(),
			Err:    e}
	}
	return n, e
}

// Close closes the connection.
func (c *SCTPConn) Close() error {
	_, e := c.send([]byte{}, sctpEoF)
	if e != nil {
		e = &net.OpError{
			Op:     "close",
			Net:    "sctp",
			Source: c.LocalAddr(),
			Addr:   c.RemoteAddr(),
			Err:    e}
	}
	b := make([]byte, 1024)
	for {
		_, eof := c.Read(b)
		if eof == io.EOF {
			break
		}
	}
	return e
}

// Abort closes the connection with abort message.
func (c *SCTPConn) Abort(reason string) error {
	_, e := c.send([]byte(reason), sctpAbort)
	if e != nil {
		e = &net.OpError{
			Op:     "abort",
			Net:    "sctp",
			Source: c.LocalAddr(),
			Addr:   c.RemoteAddr(),
			Err:    e}
	}
	return e
}

func (c *SCTPConn) send(b []byte, flag uint16) (int, error) {
	info := sndrcvInfo{}
	if n := time.Now(); !c.wd.IsZero() && n.Before(c.wd) {
		info.timetolive = uint32(c.wd.Sub(n))
	}
	info.flags = flag
	info.assocID = c.id
	return sctpSend(c.l.sock, b, &info, 0)
}

// LocalAddr returns the local network address.
func (c *SCTPConn) LocalAddr() net.Addr {
	ptr, n, e := sctpGetladdrs(c.l.sock, c.id)
	if e != nil {
		return nil
	}
	defer sctpFreeladdrs(ptr)
	return resolveFromRawAddr(ptr, n)
}

// RemoteAddr returns the remote network address.
func (c *SCTPConn) RemoteAddr() net.Addr {
	ptr, n, e := sctpGetpaddrs(c.l.sock, c.id)
	if e != nil {
		return nil
	}
	defer sctpFreepaddrs(ptr)
	return resolveFromRawAddr(ptr, n)
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
	c.rd = t
	return nil
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *SCTPConn) SetWriteDeadline(t time.Time) error {
	c.wd = t
	return nil
}

// SetRtoInfo set retransmit timer options
func (c *SCTPConn) SetRtoInfo(ini, min, max int) error {
	type opt struct {
		assocID assocT
		ini     uint32
		max     uint32
		min     uint32
	}
	attr := opt{
		assocID: c.id,
		ini:     uint32(ini),
		max:     uint32(max),
		min:     uint32(min)}
	l := unsafe.Sizeof(attr)
	p := unsafe.Pointer(&attr)

	return setSockOpt(c.l.sock, sctpRtoInfo, p, l)
}

// SetAssocinfo set association parameter
func (c *SCTPConn) SetAssocinfo(pRwnd, lRwnd, cLife, assocMaxRxt, numPeerDest int) error {
	type opt struct {
		assocID     assocT
		pRwnd       uint32
		lRwnd       uint32
		cLife       uint32
		assocMaxRxt uint16
		numPeerDest uint16
	}
	attr := opt{
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
