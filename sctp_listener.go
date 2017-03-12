package extnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"unsafe"
)

// SCTPListener is a SCTP network listener.
type SCTPListener struct {
	sock   int
	ppid   uint32
	uo     uint16
	con    map[assocT]*SCTPConn
	accept chan *SCTPConn
	m      sync.Mutex
	close  chan bool
}

// Accept implements the Accept method in the Listener interface;
// it waits for the next call and returns a generic Conn.
func (l *SCTPListener) Accept() (net.Conn, error) {
	return l.AcceptSCTP()
}

// AcceptSCTP accepts the next incoming call and returns the new connection.
func (l *SCTPListener) AcceptSCTP() (c *SCTPConn, e error) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.close == nil {
		c = <-l.accept
	}
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
	if l.close == nil {
		l.close = make(chan bool)
		sctpPolling(l.sock)
		<-l.close
	}
	return
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
	return &net.OpError{
		Op:     "connect",
		Net:    "sctp",
		Source: l.Addr(),
		Addr:   addr,
		Err:    errors.New("invalid Addr, not SCTPAddr")}
}

// ConnectSCTP create new connection of this listener
func (l *SCTPListener) ConnectSCTP(raddr *SCTPAddr) error {
	if l.close != nil {
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

// SctpHandlerStart is the error type that indicate start sctp message handler.
type SctpHandlerStart struct {
	Addr net.Addr
}

func (e *SctpHandlerStart) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("start sctp message handler on %s", e.Addr)
}

// SctpHandlerStop is the error type that indicate stop sctp message handler.
type SctpHandlerStop struct {
	Addr net.Addr
	Err  error
}

func (e *SctpHandlerStop) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return fmt.Sprintf("stop sctp message handler on %s", e.Addr)
	}
	return fmt.Sprintf(
		"recieve message failed, stop sctp message handler on %s: %s",
		e.Addr, e.Err)
}

// SctpHandlerError is the error type that indicate failure in message handler.
type SctpHandlerError struct {
	Addr net.Addr
	Err  error
}

func (e *SctpHandlerError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("message handling failed on %s: %s", e.Addr, e.Err)
}

// read data from buffer
func read(l *SCTPListener, ready chan bool) {
	if Notificator != nil {
		Notificator(&SctpHandlerStart{Addr: l.Addr()})
	}

	type sctpTlv struct {
		snType   uint16
		snFlags  uint16
		snLength uint32
	}

	ready <- true
	for {
		buf := make([]byte, RxBufferSize)
		info := sndrcvInfo{}
		flag := 0

		// receive message
		n, e := sctpRecvmsg(l.sock, buf, &info, &flag)
		if e != nil {
			eno, ok := e.(*syscall.Errno)
			if ok && eno.Temporary() {
				if Notificator != nil {
					Notificator(&SctpHandlerError{
						Addr: l.Addr(), Err: e})
				}
				continue
			} else {
				if Notificator != nil {
					Notificator(&SctpHandlerStop{
						Addr: l.Addr(), Err: e})
				}
				break
			}
		}

		// check message type is notify
		if flag&msgNotification == msgNotification {
			tlv := (*sctpTlv)(unsafe.Pointer(&buf[0]))
			switch tlv.snType {
			case sctpAssocChange:
				l.assocChangeNotify(buf[:n])
			case sctpPeerAddrChange:
				l.paddrChangeNotify(buf[:n])
			case sctpRemoteError:
				l.remoteErrorNotify(buf[:n])
			case sctpSendFailed:
				l.sendFailedNotify(buf[:n])
			case sctpShutdownEvent:
				l.shutdownNotify(buf[:n])
			case sctpAdaptationIndication:
				l.adaptationIndicationNotify(buf[:n])
			case sctpPartialDeliveryEvent:
				l.partialDeliveryNotify(buf[:n])
			case sctpSenderDryEvent:
				l.senderDryNotify(buf[:n])
			default:
				panic(fmt.Sprintf(
					"unknown notification type %d",
					tlv.snType))
			}
		} else {
			if Notificator != nil {
				Notificator(&SctpRecieveData{
					ID:        int(info.assocID),
					Stream:    int(info.stream),
					PPID:      int(info.ppid),
					Unordered: info.flags&sctpUnordered == sctpUnordered,
					Data:      buf[:n]})
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

	for _, c := range l.con {
		c.queue(nil, io.EOF)
	}
	sockClose(l.sock)

	if l.close == nil {
		l.close = make(chan bool)
	} else {
		l.close <- true
	}
}

// SctpRecieveData is the error type that indicate
// recieve data form the association.
type SctpRecieveData struct {
	ID        int
	Stream    int
	PPID      int
	Unordered bool
	Data      []byte
}

func (e *SctpRecieveData) Error() string {
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
	return fmt.Sprintf(
		"recieve data from assoc(id=%d, stream=%d, ppid=%s%s): % x",
		e.ID, e.Stream, s, uo, e.Data)
}
