package extnet

import (
	"bytes"
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

// ResolveSCTPAddr  parses addr as a SCTP address
// of the form "host:port" or "[ipv6-host%zone]:port" and
// resolves a pair of domain name and port name on the network net,
// which must be "sctp", "sctp4" or "sctp6".
// A literal address or host name for IPv6 must be enclosed in square brackets,
// as in "[::1]:80", "[ipv6-host]:http" or "[ipv6-host%zone]:80".
// Multiple address are separated by slash "/" as "host/host:port".
func ResolveSCTPAddr(network, str string) (*SCTPAddr, error) {
	switch network {
	case "sctp", "sctp4", "sctp6":
	case "":
		network = "sctp"
	default:
		return nil, net.UnknownNetworkError(network)
	}

	addr := &SCTPAddr{}
	if i := strings.LastIndex(str, ":"); i < 0 || i == len(str) {
		return nil, &net.AddrError{
			Err:  "missing port in address",
			Addr: str}
	} else if p, e := net.LookupPort(network, str[i+1:]); e != nil {
		return nil, &net.AddrError{
			Err:  "missing port in address: " + e.Error(),
			Addr: str}
	} else {
		addr.Port = p
		str = str[:i]
	}

	// set IP address
	a := strings.Split(str, "/")
	addr.IP = make([]net.IP, len(a))
	for i, s := range a {
		if s[0] == '[' {
			if s[len(s)-1] != ']' {
				return nil, &net.AddrError{
					Err:  "missing ']' in address",
					Addr: str}
			}
			s = s[1 : len(s)-1]
		}
		ip, e := net.ResolveIPAddr("ip", s)
		if e != nil {
			return nil, e
		}
		if (ip.IP.To4() != nil && network == "sctp6") ||
			(ip.IP.To4() == nil && network == "sctp4") {
			return nil, &net.AddrError{
				Err:  "mismatch of address version",
				Addr: str}
		}
		addr.IP[i] = ip.IP
	}

	if len(addr.IP) > 1 {
		for i := 0; i < len(addr.IP)-1; i++ {
			if (addr.IP[i].To4() != nil && addr.IP[i+1].To4() == nil) ||
				(addr.IP[i].To4() == nil && addr.IP[i+1].To4() != nil) {
				return nil, &net.AddrError{
					Err:  "mismatch of address version",
					Addr: str}
			}
		}
	}
	return addr, nil
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
			if i == nil {
				return nil, 0
			}
			addr[n].Family = syscall.AF_INET
			addr[n].Port = p
			addr[n].Addr = [4]byte{i[0], i[1], i[2], i[3]}
		}
		return unsafe.Pointer(&addr[0]), len(a.IP)
	} else if a.IP[0].To16() != nil {
		addr := make([]syscall.RawSockaddrInet6, len(a.IP))
		for n, i := range a.IP {
			if i.To4() != nil {
				return nil, 0
			}
			i = i.To16()
			addr[n].Family = syscall.AF_INET6
			addr[n].Port = p
			// addr[n].Flowinfo
			addr[n].Addr = [net.IPv6len]byte{}
			for j := 0; j < net.IPv6len; j++ {
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
	p := 0
	addr.IP = make([]net.IP, n)

	switch (*(*syscall.RawSockaddrAny)(ptr)).Addr.Family {
	case syscall.AF_INET:
		p = int((*(*syscall.RawSockaddrInet4)(ptr)).Port)

		for i := 0; i < n; i++ {
			a := *(*syscall.RawSockaddrInet4)(unsafe.Pointer(
				uintptr(ptr) + uintptr(16*i)))
			addr.IP[i] = net.IPv4(a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3])
		}
	case syscall.AF_INET6:
		p = int((*(*syscall.RawSockaddrInet6)(ptr)).Port)

		for i := 0; i < n; i++ {
			a := *(*syscall.RawSockaddrInet6)(unsafe.Pointer(
				uintptr(ptr) + uintptr(28*i)))
			addr.IP[i] = make([]byte, net.IPv6len)
			for j := 0; j < net.IPv6len; j++ {
				addr.IP[i][j] = a.Addr[j]
			}
		}
	default:
		panic("invalid family of address")
	}

	addr.Port = (p & 0xff) << 8
	addr.Port |= (p & 0xff00) >> 8
	return addr
}

func (a *SCTPAddr) String() string {
	var b bytes.Buffer

	for n, i := range a.IP {
		if a.IP[n].To4() != nil {
			b.WriteRune('/')
			b.WriteString(i.String())
		} else if a.IP[n].To16() != nil {
			b.WriteRune('/')
			b.WriteRune('[')
			b.WriteString(i.String())
			b.WriteRune(']')
		}
	}
	b.WriteRune(':')
	b.WriteString(strconv.Itoa(a.Port))

	return b.String()[1:]
}

// Network returns the address's network name, "sctp".
func (a *SCTPAddr) Network() string {
	return "sctp"
}
