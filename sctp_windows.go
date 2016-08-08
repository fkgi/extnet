package extnet

import (
	"log"
	"syscall"
	"unsafe"
)

const (
	IPPROTO_SCTP        = 0x84
	SCTP_BINDX_ADD_ADDR = 0x00008001
	SCTP_BINDX_REM_ADDR = 0x00008002

	SCTP_EOF       = 0x0100
	SCTP_ABORT     = 0x0200
	SCTP_UNORDERED = 0x0400
	SCTP_ADDR_OVER = 0x0800
	//	SCTP_SENDALL = 0x1000
	//	SCTP_EOR = 0x2000
	SCTP_SACK_IMMEDIATELY = 0x4000

	SOL_SCTP    = 132
	SCTP_EVENTS = 0x0000000c

	MSG_NOTIFICATION            = 0x1000
	SCTP_ASSOC_CHANGE           = 0x0001
	SCTP_PEER_ADDR_CHANGE       = 0x0002
	SCTP_REMOTE_ERROR           = 0x0003
	SCTP_SEND_FAILED            = 0x0004
	SCTP_SHUTDOWN_EVENT         = 0x0005
	SCTP_ADAPTATION_INDICATION  = 0x0006
	SCTP_PARTIAL_DELIVERY_EVENT = 0x0007
	SCTP_SENDER_DRY_EVENT       = 0x000a

	SCTP_COMM_UP        = 0x0001
	SCTP_COMM_LOST      = 0x0002
	SCTP_RESTART        = 0x0003
	SCTP_SHUTDOWN_COMP  = 0x0004
	SCTP_CANT_STR_ASSOC = 0x0005
)

var (
	fsctpBindx      *syscall.Proc
	fsctpConnectx   *syscall.Proc
	fsctpSend       *syscall.Proc
	fsctpRecvmsg    *syscall.Proc
	fsctpGetladdrs  *syscall.Proc
	fsctpFreeladdrs *syscall.Proc
	fsctpGetpaddrs  *syscall.Proc
	fsctpFreepaddrs *syscall.Proc
	dll             *syscall.DLL
)

func init() {
	var d syscall.WSAData
	e := syscall.WSAStartup(uint32(0x202), &d)
	if e != nil {
		log.Fatal(e)
	}

	dll, e = syscall.LoadDLL("sctpsp.dll")
	if e != nil {
		log.Fatal(e)
	}
	// dll.Release()

	fsctpBindx, e = dll.FindProc("internal_sctp_bindx")
	if e != nil {
		log.Fatal(e)
	}
	fsctpConnectx, e = dll.FindProc("internal_sctp_connectx")
	if e != nil {
		log.Fatal(e)
	}
	fsctpSend, e = dll.FindProc("internal_sctp_send")
	if e != nil {
		log.Fatal(e)
	}
	fsctpRecvmsg, e = dll.FindProc("internal_sctp_recvmsg")
	if e != nil {
		log.Fatal(e)
	}
	fsctpGetladdrs, e = dll.FindProc("internal_sctp_getladdrs")
	if e != nil {
		log.Fatal(e)
	}
	fsctpFreeladdrs, e = dll.FindProc("internal_sctp_freeladdrs")
	if e != nil {
		log.Fatal(e)
	}
	fsctpGetpaddrs, e = dll.FindProc("internal_sctp_getpaddrs")
	if e != nil {
		log.Fatal(e)
	}
	fsctpFreepaddrs, e = dll.FindProc("internal_sctp_freepaddrs")
	if e != nil {
		log.Fatal(e)
	}
}

type sctpEventSubscribe struct {
	dataIoEvent          uint8
	associationEvent     uint8
	addressEvent         uint8
	sendFailureEvent     uint8
	peerErrorEvent       uint8
	shutdownEvent        uint8
	partialDeliveryEvent uint8
	adaptationLayerEvent uint8
	authenticationEvent  uint8
	senderDryEvent       uint8
	streamResetEvents    uint8
}

func setNotify(fd int) error {
	event := sctp_event_subscribe{}
	event.data_io_event = 1
	event.association_event = 0
	event.address_event = 0
	event.send_failure_event = 0
	event.peer_error_event = 0
	event.shutdown_event = 1
	event.partial_delivery_event = 0
	event.adaptation_layer_event = 0
	event.authentication_event = 0
	event.sender_dry_event = 0
	event.stream_reset_events = 0

	return syscall.Setsockopt(
		syscall.Handle(fd),
		IPPROTO_SCTP,
		SCTP_EVENTS,
		(*byte)(unsafe.Pointer(&event)),
		int32(unsafe.Sizeof(event)))
}

func sockOpen() (int, error) {
	sock, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_SEQPACKET, IPPROTO_SCTP)
	return int(sock), e
}

func sockListen(fd int) error {
	return syscall.Listen(syscall.Handle(fd), ListenBufferSize)
}

func sockClose(fd int) error {
	e1 := syscall.Shutdown(syscall.Handle(fd), syscall.SHUT_RD)
	e2 := syscall.Closesocket(syscall.Handle(fd))
	if e1 != nil {
		return e1
	}
	return e2
}

func sctpBindx(fd int, addr []syscall.RawSockaddrInet4) error {
	n, _, e := fsctp_bindx.Call(
		fd,
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		SCTP_BINDX_ADD_ADDR)
	if int(n) < 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, addr []syscall.RawSockaddrInet4) (int, error) {
	t := 0
	n, _, e := fsctp_connectx.Call(
		fd,
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		uintptr(unsafe.Pointer(&t)))
	if int(n) < 0 {
		return 0, e
	}
	return t, nil
}

func sctpSend(fd int, b []byte, info *sndrcvinfo, flag int) (int, error) {
	buf := uintptr(0)
	if len(b) != 0 {
		buf = uintptr(unsafe.Pointer(&b[0]))
	}
	n, _, e := fsctp_send.Call(
		fd,
		buf,
		uintptr(len(b)),
		uintptr(unsafe.Pointer(info)),
		uintptr(flag))
	if int(n) < 0 {
		return -1, e
	}
	return int(n), nil
}

func sctpRecvmsg(fd int, b []byte, info *sndrcvinfo, flag *int) (int, error) {
	n, _, e := fsctp_recvmsg.Call(
		fd,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(len(b)),
		0,
		0,
		uintptr(unsafe.Pointer(info)),
		uintptr(unsafe.Pointer(flag)))
	if int(n) <= 0 {
		return -1, e
	}
	return int(n), nil
}

func sctpGetladdrs(fd int, id int) ([]syscall.RawSockaddrInet4, error) {
	addr := make([]syscall.RawSockaddrInet4, 100)
	n, _, e := fsctp_getladdrs.Call(
		fd,
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := addr[:int(n)]
	fsctp_freeladdrs.Call(uintptr(unsafe.Pointer(&addr[0])))

	return r, nil
}

func sctpGetpaddrs(fd int, id int) ([]syscall.RawSockaddrInet4, error) {
	addr := make([]syscall.RawSockaddrInet4, 100)
	n, _, e := fsctp_getpaddrs.Call(
		fd,
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := addr[:int(n)]
	fsctp_freepaddrs.Call(uintptr(unsafe.Pointer(&addr[0])))

	return r, nil
}
