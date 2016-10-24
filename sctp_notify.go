package extnet

import (
	"fmt"
	"io"
	"net"
	"syscall"
	"unsafe"
)

// SctpAssocUp is the error type that indicate new association is ready.
type SctpAssocUp struct {
	ID      int
	OStream int
	IStream int
}

func (e *SctpAssocUp) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"a new association(id=%d) is ready", e.ID)
}

// SctpAssocLost is the error type that indicate the association has failed.
type SctpAssocLost struct {
	ID  int
	Err uint16
}

func (e *SctpAssocLost) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"the association(id=%d) has failed", e.ID)
}

// SctpAssocShutdown is the error type that indicate
// the association has gracefully closed.
type SctpAssocShutdown struct {
	ID int
}

func (e *SctpAssocShutdown) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"the association(id=%d) has gracefully closed", e.ID)
}

// SctpAssocRestart is the error type that indicate
// SCTP has detected that the peer has restarted.
type SctpAssocRestart struct {
	ID      int
	OStream int
	IStream int
}

func (e *SctpAssocRestart) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"the association(id=%d) peer has restarted", e.ID)
}

// SctpAssocStartFail is the error type that indicate
// the association failed to setup.
type SctpAssocStartFail struct {
	ID  int
	Err uint16
}

func (e *SctpAssocStartFail) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"the association(id=%d) failed to setup", e.ID)
}

func (l *SCTPListener) assocChangeNotify(buf []byte) {
	type ntfy struct {
		chtype          uint16
		flags           uint16
		length          uint32
		state           uint16
		sacError        uint16
		outboundStreams uint16
		inboundStreams  uint16
		assocID         assocT
	}

	c := (*ntfy)(unsafe.Pointer(&buf[0]))

	switch c.state {
	case sctpCommUp:
		if Notificator != nil {
			Notificator(&SctpAssocUp{
				ID:      int(c.assocID),
				OStream: int(c.outboundStreams),
				IStream: int(c.inboundStreams)})
		}

		if _, ok := l.con[c.assocID]; ok {
			panic(fmt.Sprintf(
				"duplicate assoc id %d in new association notification",
				c.assocID))
		}

		// create new connection
		con := &SCTPConn{
			l:   l,
			id:  c.assocID,
			buf: make([]byte, 0, RxBufferSize)}
		con.win = con.buf
		con.wc.L = &con.m
		l.con[c.assocID] = con

		if l.accept != nil {
			l.accept <- con
		} else {
			con.Abort("closed")
		}
	case sctpCommLost:
		if Notificator != nil {
			Notificator(&SctpAssocLost{
				ID:  int(c.assocID),
				Err: c.sacError})
		}

		if con, ok := l.con[c.assocID]; ok {
			delete(l.con, c.assocID)
			con.queue(nil, io.EOF)
		}
	case sctpShutdownComp:
		if Notificator != nil {
			Notificator(&SctpAssocShutdown{
				ID: int(c.assocID)})
		}

		if con, ok := l.con[c.assocID]; ok {
			delete(l.con, c.assocID)
			con.queue(nil, io.EOF)
		}
	case sctpRestart:
		if Notificator != nil {
			Notificator(&SctpAssocRestart{
				ID:      int(c.assocID),
				OStream: int(c.outboundStreams),
				IStream: int(c.inboundStreams)})
		}
	case sctpCantStrAssoc:
		if Notificator != nil {
			Notificator(&SctpAssocStartFail{
				ID:  int(c.assocID),
				Err: c.sacError})
		}
	}
	return
}

// SctpPeerAddrAvailable is the error type that indicate
// this address is now reachable.
type SctpPeerAddrAvailable struct {
	ID int
	IP net.IP
}

func (e *SctpPeerAddrAvailable) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) is now reachable", e.ID)
}

// SctpPeerAddrUnreachable is the error type that indicate
// this address specified can no longer be reached.
type SctpPeerAddrUnreachable struct {
	ID  int
	IP  net.IP
	Err int
}

func (e *SctpPeerAddrUnreachable) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) can no longer be reached", e.ID)
}

// SctpPeerAddrRemoved is the error type that indicate
// this address is no longer part of the association.
type SctpPeerAddrRemoved struct {
	ID  int
	IP  net.IP
	Err int
}

func (e *SctpPeerAddrRemoved) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address is no longer part of the association(id=%d)", e.ID)
}

// SctpPeerAddrAdded is the error type that indicate
// this address is now part of the association.
type SctpPeerAddrAdded struct {
	ID int
	IP net.IP
}

func (e *SctpPeerAddrAdded) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address is now part of the association(id=%d)", e.ID)
}

// SctpPeerAddrMadePrim is the error type that indicate
// this address has now been made to be the primary destination address.
type SctpPeerAddrMadePrim struct {
	ID int
	IP net.IP
}

func (e *SctpPeerAddrMadePrim) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) is the primary destination", e.ID)
}

// SctpPeerAddrConfirmed is the error type that indicate
// this address is confirmed from peer.
type SctpPeerAddrConfirmed struct {
	ID int
	IP net.IP
}

func (e *SctpPeerAddrConfirmed) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) is confirmed from peer", e.ID)
}

func (l *SCTPListener) paddrChangeNotify(buf []byte) {
	/*
		type sockaddrStorage struct {
			length uint8     // 1
			family uint8     // 1
			pad1   [6]byte   // 6
			align  int64     // 8
			pad2   [112]byte //112 = 128
		}
	*/
	type ntfy struct {
		chtype   uint16
		flags    uint16
		length   uint32
		addr     [128]byte // sockaddrStorage
		state    int
		spcError int
		assocID  assocT
	}

	c := (*ntfy)(unsafe.Pointer(&buf[0]))
	var ip net.IP
	switch c.addr[1] {
	case syscall.AF_INET:
		a := *(*syscall.RawSockaddrInet4)(unsafe.Pointer(&c.addr))
		ip = net.IPv4(a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3])
	case syscall.AF_INET6:
		a := *(*syscall.RawSockaddrInet6)(unsafe.Pointer(&c.addr))
		ip = make([]byte, net.IPv6len)
		for j := 0; j < net.IPv6len; j++ {
			ip[j] = a.Addr[j]
		}
	default:
		panic(fmt.Sprintf(
			"invalid family of address change notification on association %d",
			c.assocID))
	}

	switch c.state {
	case sctpAddrAvailable:
		if Notificator != nil {
			Notificator(&SctpPeerAddrAvailable{
				ID: int(c.assocID),
				IP: ip})
		}
	case sctpAddrUnreachable:
		if Notificator != nil {
			Notificator(&SctpPeerAddrUnreachable{
				ID:  int(c.assocID),
				IP:  ip,
				Err: c.spcError})
		}
	case sctpAddrRemoved:
		if Notificator != nil {
			Notificator(&SctpPeerAddrRemoved{
				ID:  int(c.assocID),
				IP:  ip,
				Err: c.spcError})
		}
	case sctpAddrAdded:
		if Notificator != nil {
			Notificator(&SctpPeerAddrAdded{
				ID: int(c.assocID),
				IP: ip})
		}
	case sctpAddrMadePrim:
		if Notificator != nil {
			Notificator(&SctpPeerAddrMadePrim{
				ID: int(c.assocID),
				IP: ip})
		}
	case sctpAddrConfirmed:
		if Notificator != nil {
			Notificator(&SctpPeerAddrConfirmed{
				ID: int(c.assocID),
				IP: ip})
		}
	}
}

// SctpSendFailed is the error type that indicate
// SCTP cannot deliver a message.
type SctpSendFailed struct {
	ID int
}

func (e *SctpSendFailed) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"message send failed on association(id=%d)", e.ID)
}

func (l *SCTPListener) sendFailedNotify(buf []byte) {
	type ntfy struct {
		sstype   uint16
		flags    uint16
		length   uint32
		ssfError uint32
		info     sndrcvInfo
		assocID  assocT
		data     []byte
	}
	c := (*ntfy)(unsafe.Pointer(&buf[0]))
	if Notificator != nil {
		Notificator(&SctpSendFailed{
			ID: int(c.assocID)})
	}
}

// SctpRemoteError is the error type that indicate
// remote peer send an Operational Error message.
type SctpRemoteError struct {
	ID  int
	Err uint16
}

func (e *SctpRemoteError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"remote peer send error on association(id=%d)", e.ID)
}

func (l *SCTPListener) remoteErrorNotify(buf []byte) {
	type ntfy struct {
		sstype   uint16
		flags    uint16
		length   uint32
		sreError uint16
		assocID  assocT
		data     []byte
	}
	c := (*ntfy)(unsafe.Pointer(&buf[0]))
	if Notificator != nil {
		Notificator(&SctpRemoteError{
			ID: int(c.assocID)})
	}
}

// SctpShutdown is the error type that indicate
// the association is required shutdown.
type SctpShutdown struct {
	ID int
}

func (e *SctpShutdown) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"association(id=%d) is required shutdown", e.ID)
}

func (l *SCTPListener) shutdownNotify(buf []byte) {
	type ntfy struct {
		chtype  uint16
		flags   uint16
		length  uint32
		assocID assocT
	}

	c := (*ntfy)(unsafe.Pointer(&buf[0]))
	if Notificator != nil {
		Notificator(&SctpShutdown{
			ID: int(c.assocID)})
	}
}
