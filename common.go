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

type sndrcvinfo struct {
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

// TCP/SCTP address interface
/*
type Addr interface {
	String() string
	Equals(Addr) bool
	Network() string
}

type Listener interface {
	Close() error
	Accept() (Conn, error)
	Connect(Addr) error
	LocalAddr() (Addr, error)
}

type Conn interface {
	Close() error
	Abort() error
	Write([]byte) (int, error)
	Read() ([]byte, error)
	LocalAddr() (Addr, error)
	RemoteAddr() (Addr, error)
}

func Listen(a Addr) (Listener, error) {
	s, ok := a.(*SCTPAddr)
	if ok {
		return ListenSCTP(s)
	}
	//t, ok := a.(TCPAddr)
	if ok {
		// return ListenTCP(addr)
	}
	return nil, errors.New("unknown address type")
}

func ResolveAddr(s, addr string) (Addr, error) {
	return ResolveSCTPAddr(addr)
}
*/
