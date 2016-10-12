package extnet

import (
	"log"
	"net"
	"os"
)

const (
	// RxBufferSize is network recieve queue size
	RxBufferSize = 10240
	// BacklogSize is accept queue size
	BacklogSize = 128
)

var (
	trace *log.Logger
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

// TraceEnable enables trace log output
func TraceEnable() {
	trace = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	trace.Println("sctp | trace | assoc id | trace enabled")
}

// TraceDisable disables trace log output
func TraceDisable() {
	if trace != nil {
		trace.Println("sctp | trace | assoc id | trace disabled")
	}
	trace = nil
}

func tracePrint(s interface{}, i assocT) {
	if trace != nil {
		trace.Println("sctp | trace |", i, "|", s)
	}
}

func errorPrint(s interface{}, i assocT) {
	if trace != nil {
		trace.Println("sctp | error |", i, "|", s)
	}
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
