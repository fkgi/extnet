package extnet

import "log"

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
