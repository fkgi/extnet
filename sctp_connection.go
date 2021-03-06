package extnet

import (
	"fmt"
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
	buf := make([]byte, len(b))
	copy(buf, b)
	n, e := c.send(buf, c.l.ppid, 0, 0)
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

// WriteToStream write data with specified stream and ppid.
func (c *SCTPConn) WriteToStream(b []byte, s uint16, i uint32) (int, error) {
	buf := make([]byte, len(b))
	copy(buf, b)
	n, e := c.send(buf, i, s, 0)
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
	_, e := c.send([]byte{}, 0, 0, sctpEoF)
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
	buf := make([]byte, len([]byte(reason)))
	copy(buf, []byte(reason))
	_, e := c.send(buf, 0, 0, sctpAbort)
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

func (c *SCTPConn) send(b []byte, p uint32, s, f uint16) (int, error) {
	info := sndrcvInfo{}
	if n := time.Now(); !c.wd.IsZero() && n.Before(c.wd) {
		info.timetolive = uint32(c.wd.Sub(n))
	}
	info.stream = s
	info.flags = f | c.l.uo
	info.assocID = c.id
	info.ppid = p

	i, e := sctpSend(c.l.sock, b, &info, 0)
	if Notificator != nil {
		if i < 0 {
			i = 0
		}
		Notificator(&SctpSendData{
			ID:        int(info.assocID),
			Stream:    int(info.stream),
			PPID:      int(info.ppid),
			Unordered: info.flags&sctpUnordered == sctpUnordered,
			Data:      b[:i],
			Err:       e})
	}
	return i, e
}

// SctpSendData is the error type that indicate
// send data to the association.
type SctpSendData struct {
	ID        int
	Stream    int
	PPID      int
	Unordered bool
	Data      []byte
	Err       error
}

func (e *SctpSendData) Error() string {
	if e == nil {
		return "<nil>"
	}
	s, ok := ppidStr[e.PPID]
	if !ok {
		s = "Unassigned"
	}
	uo := ""
	if e.Unordered {
		uo = ", unorderd"
	}

	if e.Err != nil {
		return fmt.Sprintf(
			"send data to association(id=%d, stream=%d, ppid=%s%s) failed, %s: % x",
			e.ID, e.Stream, s, uo, e.Err, e.Data)
	}
	return fmt.Sprintf(
		"send data to association(id=%d, stream=%d, ppid=%s%s): % x",
		e.ID, e.Stream, s, uo, e.Data)
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
