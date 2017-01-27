/*
Package extnet provides a portable interface for network I/O, including SCTP.
*/
package extnet

const (
	// RxBufferSize is network recieve queue size
	RxBufferSize = 10240
	// BacklogSize is accept queue size
	BacklogSize = 128

	// CloseNotifyPpid is used close listener
	CloseNotifyPpid = 9999
	// CloseNotifyCotext is used close listener
	CloseNotifyCotext = 9999
)

// Notificator is called when error or trace event are occured
var Notificator func(e error)

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

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
