package extnet

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"syscall"
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
	if len(t) < 2 {
		return nil, errors.New("invalid input")
	}

	// set IP address
	addr.IP = make([]net.IP, len(t)-1)
	for i, s := range t[0 : len(t)-1] {
		addr.IP[i] = net.ParseIP(s)
		if addr.IP[i] == nil {
			return nil, errors.New("invalid input")
		}
	}

	// set port
	var e error
	addr.Port, e = strconv.Atoi(t[len(t)-1])
	return addr, e
}

func (a *SCTPAddr) rawAddr() []syscall.RawSockaddrInet4 {
	addr := make([]syscall.RawSockaddrInet4, len(a.IP))
	for n, i := range a.IP {
		i = i.To4()
		addr[n].Family = syscall.AF_INET
		addr[n].Port = uint16(a.Port)
		addr[n].Addr = [4]byte{i[0], i[1], i[2], i[3]}
	}
	return addr
}

func resolveFromRawAddr(addr []syscall.RawSockaddrInet4) *SCTPAddr {
	raddr := &SCTPAddr{}
	raddr.Port = int(addr[0].Port)
	raddr.IP = make([]net.IP, len(addr))
	for n, i := range addr {
		raddr.IP[n] = net.IPv4(i.Addr[0], i.Addr[1], i.Addr[2], i.Addr[3])
	}
	return raddr
}

func (a *SCTPAddr) String() (s string) {
	for _, i := range a.IP {
		s += i.String()
		s += ":"
	}
	s += strconv.Itoa(a.Port)
	return
}

// Network returns the address's network name, "sctp".
func (a *SCTPAddr) Network() string {
	return "sctp"
}

/*
func (a *SCTPAddr) Equals(addr Addr) bool {
	b, ok := addr.(*SCTPAddr)
	if !ok {
		return false
	}
	if a.Port != b.Port {
		return false
	}
	if len(a.IP) != len(b.IP) {
		return false
	}
aAddr:
	for _, i := range a.IP {
		for _, j := range b.IP {
			if i.Equal(j) {
				continue aAddr
			}
		}
		return false
	}
	return true
}
*/
