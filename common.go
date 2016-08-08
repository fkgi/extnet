package extnet

import (
	//	"errors"
	"net"
	//	"strconv"
	//	"syscall"
)

const (
	// RxBufferSize is network recieve queue size
	RxBufferSize = 10240
	// ListenBufferSize is accept queue size
	ListenBufferSize = 65535
	// MaxAddressCount is count of multi-homed IP address
	MaxAddressCount = 10
)

type sndrcvInfo struct {
	stream     uint16
	ssn        uint16
	flags      uint16
	ppid       uint32
	context    uint32
	timetolive uint32
	tsn        uint32
	cumtsn     uint32
	assocID    int32
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
	e = sctpBindx(sock, laddr.rawAddr())
	if e != nil {
		e = &net.OpError{Op: "bindx", Net: "sctp", Addr: nil, Err: e}
		sockClose(sock)
		return -1, e
	}
	return sock, nil
}
