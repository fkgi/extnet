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
)

const (
	sctpEoF       = C.SCTP_EOF
	sctpAbort     = C.SCTP_ABORT
	sctpUnordered = C.SCTP_UNORDERED
	sctpAddrOver  = C.SCTP_ADDR_OVER

	// SCTP_SENDALL = C.SCTP_SENDALL
	// SCTP_EOR = C.SCTP_EOR

	//SCTP_SACK_IMMEDIATELY = C.SCTP_SACK_IMMEDIATELY

	// SOL_SCTP    = C.SOL_SCTP
	// SCTP_EVENTS = C.SCTP_EVENTS

	msgNotification          = C.MSG_NOTIFICATION
	sctpAssocChange          = C.SCTP_ASSOC_CHANGE
	sctpPeerAddrChange       = C.SCTP_PEER_ADDR_CHANGE
	sctpRemoteError          = C.SCTP_REMOTE_ERROR
	sctpSendFailed           = C.SCTP_SEND_FAILED
	sctpShutdownEvent        = C.SCTP_SHUTDOWN_EVENT
	sctpAdaptationIndication = C.SCTP_ADAPTATION_INDICATION
	sctpPartialDeliveryEvent = C.SCTP_PARTIAL_DELIVERY_EVENT
	sctpSenderDryEvent       = C.SCTP_SENDER_DRY_EVENT

	sctpCommUp       = C.SCTP_COMM_UP
	sctpCommLost     = C.SCTP_COMM_LOST
	sctpRestart      = C.SCTP_RESTART
	sctpShutdownComp = C.SCTP_SHUTDOWN_COMP
	sctpCantStrAssoc = C.SCTP_CANT_STR_ASSOC

	sctpInitMsg   = C.SCTP_INITMSG
	sctpRtoInfo   = C.SCTP_RTOINFO
	sctpAssocInfo = C.SCTP_ASSOCINFO
	sctpNodelay   = C.SCTP_NODELAY
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
	l := unsafe.Sizeof(event)
	p := unsafe.Pointer(&event)

	return setSockOpt(fd, C.SCTP_EVENTS, p, l)
}

func setSockOpt(fd, opt int, p unsafe.Pointer, l uintptr) error {
	n, e := C.setsockopt(
		C.int(fd),
		C.SOL_SCTP,
		C.int(opt),
		p,
		C.socklen_t(l))
	if int(n) < 0 {
		return e
	}
	return nil
}

func sockOpen() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_SEQPACKET, C.IPPROTO_SCTP)
}

func sockListen(fd int) error {
	return syscall.Listen(fd, BacklogSize)
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
