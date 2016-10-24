package extnet

import (
	"errors"
	"fmt"
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
			Op:   "accept",
			Net:  "sctp",
			Addr: l.Addr(),
			Err:  errors.New("socket is closed")}
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

// Connect create new connection of this listener
func (l *SCTPListener) Connect(addr net.Addr) error {
	if a, ok := addr.(*SCTPAddr); ok {
		return l.ConnectSCTP(a)
	}
	laddr := l.Addr()
	return &net.OpError{
		Op:     "connect",
		Net:    "sctp",
		Source: laddr,
		Addr:   addr,
		Err:    errors.New("invalid Addr, not SCTPAddr")}
}

// ConnectSCTP create new connection of this listener
func (l *SCTPListener) ConnectSCTP(raddr *SCTPAddr) error {
	if l.sock == -1 {
		return &net.OpError{
			Op:     "connect",
			Net:    "sctp",
			Source: l.Addr(),
			Addr:   raddr,
			Err:    errors.New("socket is closed")}
	}

	// connect SCTP connection to raddr
	ptr, n := raddr.rawAddr()
	_, e := sctpConnectx(l.sock, ptr, n)
	if e != nil {
		return &net.OpError{
			Op:     "connect",
			Net:    "sctp",
			Source: l.Addr(),
			Addr:   raddr,
			Err:    e}
	}
	return nil
}

// read data from buffer
func read(l *SCTPListener) {
	type sctpTlv struct {
		snType   uint16
		snFlags  uint16
		snLength uint32
	}

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
				l.sendFailedNotify(buf[:n])
			case sctpRemoteError:
				l.remoteErrorNotify(buf[:n])
			case sctpShutdownEvent:
				l.shutdownNotify(buf[:n])
			case sctpPartialDeliveryEvent:
				// l.partialDeliveryNotify(buf[:n])
			case sctpAdaptationIndication:
				// l.adaptationIndicationNotify(buf[:n])
			case sctpSenderDryEvent:
				// l.senderDryNotify(buf[:n])
			}
		} else {
			if Notificator != nil {
				Notificator(&SctpRecieveData{
					ID: int(info.assocID)})
			}
			// matching exist connection
			if p, ok := l.con[info.assocID]; ok {
				p.queue(buf[:n], nil)
			} else {
				panic(fmt.Sprintf(
					"data recieved from unknown assoc id %d",
					info.assocID))
			}
		}
	}
}

// SctpRecieveData is the error type that indicate
// recieve data form the association.
type SctpRecieveData struct {
	ID int
}

func (e *SctpRecieveData) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"recieve data from association(id=%d)", e.ID)
}
