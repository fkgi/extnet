package extnet

import (
	"fmt"
	"log"
	"net"
)

const (
	// RxBufferSize is network recieve queue size
	RxBufferSize = 10240
	// BacklogSize is accept queue size
	BacklogSize = 128
)

// Notificator is called when error or trace event are occured
var Notificator = func(e error) {
	log.Println(e)
}

type sndrcvInfo struct {
	stream     uint16
	ssn        uint16
	flags      uint16
	ppid       uint32
	context    uint32
	timetolive uint32
	tsn        uint32
	cumtsn     uint32
	assocID    assocT
}

// bind SCTP socket
func bindsocket(laddr *SCTPAddr) (int, error) {
	// create SCTP connection socket
	sock, e := sockOpen()
	if e != nil {
		e = &net.OpError{Op: "makesock", Net: "sctp", Addr: nil, Err: e}
		return -1, e
	}

	// set notifycation enabled
	e = setNotify(sock)
	if e != nil {
		sockClose(sock)
		e = &net.OpError{Op: "setsockopt", Net: "sctp", Addr: nil, Err: e}
		return -1, e
	}

	// bind SCTP connection
	ptr, n := laddr.rawAddr()
	e = sctpBindx(sock, ptr, n)
	if e != nil {
		e = &net.OpError{Op: "bindx", Net: "sctp", Addr: nil, Err: e}
		sockClose(sock)
		return -1, e
	}
	return sock, nil
}

// SctpError is the erro type returned by SCTP functions.
type SctpError struct {
	timeout bool
	Err     error
}

func (e *SctpError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Err.Error()
}

// Timeout indicate timeout is occured.
func (e *SctpError) Timeout() bool {
	return e.timeout
}

// SctpRecieveData is the error type that indicate recieve data form the association.
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
		"A new association(id=%d) is ready", e.ID)
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
		"The association(id=%d) has failed", e.ID)
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
		"The association(id=%d) has gracefully closed", e.ID)
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
		"The association(id=%d) peer has restarted", e.ID)
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
		"The association(id=%d) failed to setup", e.ID)
}

// SctpPeerAddrAvailable is the error type that indicate
// this address is now reachable.
type SctpPeerAddrAvailable struct {
	ID   int
	Addr []net.Addr
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
	ID   int
	Addr []net.Addr
	Err  int
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
	ID   int
	Addr []net.Addr
	Err  int
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
	ID   int
	Addr []net.Addr
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
	ID   int
	Addr []net.Addr
}

func (e *SctpPeerAddrMadePrim) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) has now been made to be the primary destination", e.ID)
}

// SctpPeerAddrConfirmed is the error type that indicate
// this address is confirmed from peer.
type SctpPeerAddrConfirmed struct {
	ID   int
	Addr []net.Addr
}

func (e *SctpPeerAddrConfirmed) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"address of the association(id=%d) is confirmed from peer", e.ID)
}
