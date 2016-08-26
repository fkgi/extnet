package extnet

import (
	"log"
	"syscall"
	"unsafe"
)

const (
	ipprotoSctp      = 0x84
	sctpBindxAddAddr = 0x00008001
	// sctpBindxRemAddr = 0x00008002

	sctpEoF   = 0x0100
	sctpAbort = 0x0200
	// SCTP_UNORDERED = 0x0400
	// SCTP_ADDR_OVER = 0x0800
	// SCTP_SENDALL = 0x1000
	// SCTP_EOR = 0x2000
	// SCTP_SACK_IMMEDIATELY = 0x4000

	// solSctp     = 132
	sctpRtoInfo   = 0x00000001
	sctpAssocInfo = 0x00000002
	sctpInitMsg   = 0x00000003
	sctpNodelay   = 0x00000004
	sctpEvents    = 0x0000000c

	msgNotification          = 0x1000
	sctpAssocChange          = 0x0001
	sctpPeerAddrChange       = 0x0002
	sctpRemoteError          = 0x0003
	sctpSendFailed           = 0x0004
	sctpShutdownEvent        = 0x0005
	sctpAdaptationIndication = 0x0006
	sctpPartialDeliveryEvent = 0x0007
	sctpSenderDryEvent       = 0x000a

	sctpCommUp       = 0x0001
	sctpCommLost     = 0x0002
	sctpRestart      = 0x0003
	sctpShutdownComp = 0x0004
	sctpCantStrAssoc = 0x0005
)

type assocT uint32

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

type eventSubscribe struct {
	dataIo          uint8
	association     uint8
	address         uint8
	sendFailure     uint8
	peerError       uint8
	shutdown        uint8
	partialDelivery uint8
	adaptationLayer uint8
	authentication  uint8
	senderDry       uint8
	streamReset     uint8
}

func setNotify(fd int) error {
	event := eventSubscribe{}
	event.dataIo = 1
	event.association = 0
	event.address = 0
	event.sendFailure = 0
	event.peerError = 0
	event.shutdown = 1
	event.partialDelivery = 0
	event.adaptationLayer = 0
	event.authentication = 0
	event.senderDry = 0
	event.streamReset = 0

	l := unsafe.Sizeof(event)
	p := unsafe.Pointer(&event)

	return setSockOpt(fd, sctpEvents, p, l)
}

func setSockOpt(fd, opt int, p unsafe.Pointer, l uintptr) error {
	return syscall.Setsockopt(
		syscall.Handle(fd),
		ipprotoSctp,
		int32(opt),
		(*byte)(p),
		int32(l))

}

func sockOpen() (int, error) {
	sock, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_SEQPACKET, ipprotoSctp)
	return int(sock), e
}

func sockListen(fd int) error {
	return syscall.Listen(syscall.Handle(fd), BacklogSize)
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
	n, _, e := fsctpBindx.Call(
		uintptr(fd),
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		sctpBindxAddAddr)
	if int(n) < 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, addr []syscall.RawSockaddrInet4) (int, error) {
	t := 0
	n, _, e := fsctpConnectx.Call(
		uintptr(fd),
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		uintptr(unsafe.Pointer(&t)))
	if int(n) < 0 {
		return 0, e
	}
	return t, nil
}

func sctpSend(fd int, b []byte, info *sndrcvInfo, flag int) (int, error) {
	buf := uintptr(0)
	if len(b) != 0 {
		buf = uintptr(unsafe.Pointer(&b[0]))
	}
	n, _, e := fsctpSend.Call(
		uintptr(fd),
		buf,
		uintptr(len(b)),
		uintptr(unsafe.Pointer(info)),
		uintptr(flag))
	if int(n) < 0 {
		return -1, e
	}
	return int(n), nil
}

func sctpRecvmsg(fd int, b []byte, info *sndrcvInfo, flag *int) (int, error) {
	n, _, e := fsctpRecvmsg.Call(
		uintptr(fd),
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
	n, _, e := fsctpGetladdrs.Call(
		uintptr(fd),
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := addr[:int(n)]
	fsctpFreeladdrs.Call(uintptr(unsafe.Pointer(&addr[0])))

	return r, nil
}

func sctpGetpaddrs(fd int, id int) ([]syscall.RawSockaddrInet4, error) {
	addr := make([]syscall.RawSockaddrInet4, 100)
	n, _, e := fsctpGetpaddrs.Call(
		uintptr(fd),
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, e
	}
	r := addr[:int(n)]
	fsctpFreepaddrs.Call(uintptr(unsafe.Pointer(&addr[0])))

	return r, nil
}
