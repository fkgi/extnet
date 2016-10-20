package extnet

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// SCTPAddr represents the address of a SCTP end point.
type SCTPAddr struct {
	IP   []net.IP
	Port int
}

// ResolveSCTPAddr parses addr as a SCTP address.
func ResolveSCTPAddr(str string) (*SCTPAddr, error) {
	addr := &SCTPAddr{}

	t := strings.Split(str, ":")
	if len(t) != 2 {
		return nil, errors.New("invalid input")
	}

	// set IP address
	a := strings.Split(t[0], "/")
	addr.IP = make([]net.IP, len(a))
	for i, s := range a {
		addr.IP[i] = net.ParseIP(s)
		if addr.IP[i] == nil {
			return nil, errors.New("invalid input")
		}
	}

	// set port
	var e error
	addr.Port, e = strconv.Atoi(t[1])
	return addr, e
}

func (a *SCTPAddr) rawAddr() (unsafe.Pointer, int) {
	if len(a.IP) == 0 {
		return nil, 0
	}
	p := uint16(a.Port<<8) & 0xff00
	p |= uint16(a.Port>>8) & 0x00ff

	if a.IP[0].To4() != nil {
		addr := make([]syscall.RawSockaddrInet4, len(a.IP))
		for n, i := range a.IP {
			i = i.To4()
			addr[n].Family = syscall.AF_INET
			addr[n].Port = p
			addr[n].Addr = [4]byte{i[0], i[1], i[2], i[3]}
		}
		return unsafe.Pointer(&addr[0]), len(a.IP)
	} else if a.IP[0].To16() != nil {
		addr := make([]syscall.RawSockaddrInet6, len(a.IP))
		for n, i := range a.IP {
			i = i.To16()
			addr[n].Family = syscall.AF_INET6
			addr[n].Port = p
			// addr[n].Flowinfo
			addr[n].Addr = [16]byte{}
			for j := 0; j < 16; j++ {
				addr[n].Addr[j] = i[j]
			}
			// addr[n].Scope_id
		}
		return unsafe.Pointer(&addr[0]), len(a.IP)
	} else {
		return nil, 0
	}
}

func resolveFromRawAddr(ptr unsafe.Pointer, n int) *SCTPAddr {
	addr := &SCTPAddr{}
	switch (*(*syscall.RawSockaddrAny)(ptr)).Addr.Family {
	case syscall.AF_INET:
		p := int((*(*syscall.RawSockaddrInet4)(ptr)).Port)
		addr.Port = (p & 0xff) << 8
		addr.Port |= (p & 0xff00) >> 8

		addr.IP = make([]net.IP, n)
		for i := 0; i < n; i++ {
			a := *(*syscall.RawSockaddrInet4)(unsafe.Pointer(uintptr(ptr) + uintptr(16*i)))
			addr.IP[i] = net.IPv4(a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3])
		}
	case syscall.AF_INET6:
		p := int((*(*syscall.RawSockaddrInet6)(ptr)).Port)
		addr.Port = (p & 0xff) << 8
		addr.Port |= (p & 0xff00) >> 8

		addr.IP = make([]net.IP, n)
		for i := 0; i < n; i++ {
			a := *(*syscall.RawSockaddrInet6)(unsafe.Pointer(uintptr(ptr) + uintptr(28*i)))
			addr.IP[i] = make([]byte, 16)
			for j := 0; j < 16; j++ {
				addr.IP[i][j] = a.Addr[j]
			}
		}
	default:
		panic("invalid family of address")
	}
	return addr
}

func (a *SCTPAddr) String() (s string) {
	for _, i := range a.IP {
		s += i.String()
		s += "/"
	}
	s = s[:len(s)-1] + ":" + strconv.Itoa(a.Port)
	return
}

// Network returns the address's network name, "sctp".
func (a *SCTPAddr) Network() string {
	return "sctp"
}
