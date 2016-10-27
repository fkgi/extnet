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

	sctpAddrAvailable   = 0x0001
	sctpAddrUnreachable = 0x0002
	sctpAddrRemoved     = 0x0003
	sctpAddrAdded       = 0x0004
	sctpAddrMadePrim    = 0x0005
	sctpAddrConfirmed   = 0x0006
)

type assocT uint32

var (
	fsctpBindx      *syscall.Proc
	fsctpConnectx   *syscall.Proc
	fsctpSend       *syscall.Proc
	fsctpSendMsg    *syscall.Proc
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
	fsctpSendMsg, e = dll.FindProc("internal_sctp_sendmsg")
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

func setNotify(fd int) error {
	type opt struct {
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

	event := opt{
		dataIo:          1,
		association:     1,
		address:         0,
		sendFailure:     0,
		peerError:       0,
		shutdown:        1,
		partialDelivery: 0,
		adaptationLayer: 0,
		authentication:  0,
		senderDry:       0,
		streamReset:     0}
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
	sock, e := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_SEQPACKET,
		ipprotoSctp)
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

func sctpBindx(fd int, ptr unsafe.Pointer, l int) error {
	n, _, e := fsctpBindx.Call(
		uintptr(fd),
		uintptr(ptr),
		uintptr(l),
		sctpBindxAddAddr)
	if int(n) < 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, ptr unsafe.Pointer, l int) (assocT, error) {
	t := assocT(0)
	n, _, e := fsctpConnectx.Call(
		uintptr(fd),
		uintptr(ptr),
		uintptr(l),
		uintptr(unsafe.Pointer(&t)))
	if int(n) < 0 {
		return t, e
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

func sctpPolling(fd int) error {
	addr := &syscall.RawSockaddrInet4{
		Family: syscall.AF_INET,
		Port:   0,
		Addr:   [4]byte{0x7f, 0x00, 0x00, 0x01}}
	n, _, e := fsctpSendMsg.Call(
		uintptr(fd),
		uintptr(unsafe.Pointer(&[]byte{0x00}[0])),
		uintptr(1),
		uintptr(unsafe.Pointer(addr)),
		uintptr(unsafe.Sizeof(*addr)),
		uintptr(CloseNotifyPpid),   // ppid
		uintptr(0),                 // flags
		uintptr(0),                 // stream_no
		uintptr(0),                 // timetolive
		uintptr(CloseNotifyCotext)) // context

	if int(n) < 0 {
		return e
	}
	return nil
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

func sctpGetladdrs(fd int, id assocT) (unsafe.Pointer, int, error) {
	var addr unsafe.Pointer
	n, _, e := fsctpGetladdrs.Call(
		uintptr(fd),
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, int(n), e
	}
	return addr, int(n), nil
}

func sctpFreeladdrs(addr unsafe.Pointer) {
	fsctpFreeladdrs.Call(uintptr(addr))
}

func sctpGetpaddrs(fd int, id assocT) (unsafe.Pointer, int, error) {
	var addr unsafe.Pointer
	n, _, e := fsctpGetpaddrs.Call(
		uintptr(fd),
		uintptr(id),
		uintptr(unsafe.Pointer(&addr)))
	if int(n) <= 0 {
		return nil, int(n), e
	}
	return addr, int(n), nil
}

func sctpFreepaddrs(addr unsafe.Pointer) {
	fsctpFreepaddrs.Call(uintptr(addr))
}
