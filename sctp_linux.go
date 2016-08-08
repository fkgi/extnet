package extnet

/*
#cgo CFLAGS: -Wall
#cgo LDFLAGS: -lsctp

#include <unistd.h>
#include <stdio.h>
#include <netinet/in.h>
#include <netinet/sctp.h>
#include <arpa/inet.h>

*/
import "C"

import (
	"syscall"
	"unsafe"
	// "log"
)

const (
	// IPPROTO_SCTP        = C.IPPROTO_SCTP
	// SCTP_BINDX_ADD_ADDR = C.SCTP_BINDX_ADD_ADDR
	// SCTP_BINDX_REM_ADDR = C.SCTP_BINDX_REM_ADDR

	SCTP_EOF       = C.SCTP_EOF
	SCTP_ABORT     = C.SCTP_ABORT
	SCTP_UNORDERED = C.SCTP_UNORDERED
	SCTP_ADDR_OVER = C.SCTP_ADDR_OVER

	// SCTP_SENDALL = C.SCTP_SENDALL
	// SCTP_EOR = C.SCTP_EOR

	SCTP_SACK_IMMEDIATELY = C.SCTP_SACK_IMMEDIATELY

	// SOL_SCTP    = C.SOL_SCTP
	// SCTP_EVENTS = C.SCTP_EVENTS

	MSG_NOTIFICATION            = C.MSG_NOTIFICATION
	SCTP_ASSOC_CHANGE           = C.SCTP_ASSOC_CHANGE
	SCTP_PEER_ADDR_CHANGE       = C.SCTP_PEER_ADDR_CHANGE
	SCTP_REMOTE_ERROR           = C.SCTP_REMOTE_ERROR
	SCTP_SEND_FAILED            = C.SCTP_SEND_FAILED
	SCTP_SHUTDOWN_EVENT         = C.SCTP_SHUTDOWN_EVENT
	SCTP_ADAPTATION_INDICATION  = C.SCTP_ADAPTATION_INDICATION
	SCTP_PARTIAL_DELIVERY_EVENT = C.SCTP_PARTIAL_DELIVERY_EVENT
	SCTP_SENDER_DRY_EVENT       = C.SCTP_SENDER_DRY_EVENT

	SCTP_COMM_UP        = C.SCTP_COMM_UP
	SCTP_COMM_LOST      = C.SCTP_COMM_LOST
	SCTP_RESTART        = C.SCTP_RESTART
	SCTP_SHUTDOWN_COMP  = C.SCTP_SHUTDOWN_COMP
	SCTP_CANT_STR_ASSOC = C.SCTP_CANT_STR_ASSOC
)

type assocT C.sctp_assoc_t

type eventSubscribe struct {
	dataIo          uint8
	association     uint8
	address         uint8
	peerError       uint8
	shutdown        uint8
	partialDelivery uint8
	adaptationLayer uint8
	authentication  uint8
}

func setNotify(fd int) error {
	event := eventSubscribe{}
	event.dataIo = 1
	event.association = 1
	event.address = 1
	event.peerError = 1
	event.shutdown = 1
	event.partialDelivery = 1
	event.adaptationLayer = 1
	event.authentication = 1

	n, e := C.setsockopt(
		C.int(fd),
		C.SOL_SCTP,
		C.SCTP_EVENTS,
		unsafe.Pointer(&event),
		C.socklen_t(unsafe.Sizeof(event)))
	if int(n) < 0 {
		return e
	}
	return nil
}

func sockOpen() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_SEQPACKET, C.IPPROTO_SCTP)
}

func sockListen(fd int) error {
	return syscall.Listen(fd, ListenBufferSize)
}

func sockClose(fd int) error {
	return syscall.Close(fd)
}

func sctpBindx(fd int, addr []syscall.RawSockaddrInet4) error {
	n, e := C.sctp_bindx(
		C.int(fd),
		(*C.struct_sockaddr)(unsafe.Pointer(&addr[0])),
		C.int(len(addr)),
		C.SCTP_BINDX_ADD_ADDR)
	if int(n) < 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, addr []syscall.RawSockaddrInet4) (int, error) {
	t := 0
	n, e := C.sctp_connectx(
		C.int(fd),
		(*C.struct_sockaddr)(unsafe.Pointer(&addr[0])),
		C.int(len(addr)),
		(*C.sctp_assoc_t)(unsafe.Pointer(&t)))
	if int(n) < 0 {
		return 0, e
	}
	return t, nil
}

func sctpSend(fd int, b []byte, info *sndrcvInfo, flag int) (int, error) {
	buf := unsafe.Pointer(nil)
	if len(b) > 0 {
		buf = unsafe.Pointer(&b[0])
	}
	n, e := C.sctp_send(
		C.int(fd),
		buf,
		C.size_t(len(b)),
		(*C.struct_sctp_sndrcvinfo)(unsafe.Pointer(info)),
		C.int(flag))
	if int(n) < 0 {
		return -1, e
	}
	return int(n), nil
}

func sctpRecvmsg(fd int, b []byte, info *sndrcvInfo, flag *int) (int, error) {
	n, e := C.sctp_recvmsg(
		C.int(fd),
		unsafe.Pointer(&b[0]),
		C.size_t(len(b)),
		nil,
		nil,
		(*C.struct_sctp_sndrcvinfo)(unsafe.Pointer(info)),
		(*C.int)(unsafe.Pointer(flag)))
	if int(n) < 0 {
		return -1, e
	}
	return int(n), nil
}

func sctpGetladdrs(fd int, id int) ([]syscall.RawSockaddrInet4, error) {
	addr := make([]syscall.RawSockaddrInet4, MaxAddressCount)
	n, e := C.sctp_getladdrs(
		C.int(fd),
		C.sctp_assoc_t(id),
		(**C.struct_sockaddr)(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := make([]syscall.RawSockaddrInet4, int(n))
	for i := range r {
		r[i] = addr[i]
	}
	C.sctp_freeladdrs((*C.struct_sockaddr)(unsafe.Pointer(&addr[0])))

	return r, nil
}

func sctpGetpaddrs(fd int, id int) ([]syscall.RawSockaddrInet4, error) {
	addr := make([]syscall.RawSockaddrInet4, MaxAddressCount)
	n, e := C.sctp_getpaddrs(
		C.int(fd),
		C.sctp_assoc_t(C.int(id)),
		(**C.struct_sockaddr)(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := make([]syscall.RawSockaddrInet4, int(n))
	for i := range r {
		r[i] = addr[i]
	}
	C.sctp_freepaddrs((*C.struct_sockaddr)(unsafe.Pointer(&addr[0])))

	return r, nil
}
