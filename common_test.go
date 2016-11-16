package extnet

import (
	"net"
	"os"
	"strconv"
	"testing"
)

var (
	testAddrs  = make([]string, 10)
	basePort   = 10000
	offsetPort = 10
	testStr    = "this is test"
)

func TestMain(m *testing.M) {

	var addrs []net.IP
	result := 0

	if ipns, e := net.InterfaceAddrs(); e != nil {
		os.Exit(1)
	} else {
		for _, ipn := range ipns {
			if a, ok := (ipn).(*net.IPNet); !ok {
				os.Exit(1)
			} else if a.IP.IsGlobalUnicast() {
				addrs = append(addrs, a.IP)
			}
		}
	}

	// test for IP v4 address
	for _, addr := range addrs {
		if addr.To4() != nil {
			initAddrs(addr.String())
			result += m.Run()
			break
		}
	}

	// test for IP v6 address
	for _, addr := range addrs {
		if addr.To4() == nil && addr.To16() != nil {
			initAddrs("[" + addr.String() + "]")
			result += m.Run()
			break
		}
	}

	// test for multiple IP v4 address
	s := ""
	for _, addr := range addrs {
		if addr.To4() != nil {
			s = s + "/" + addr.String()
		}
	}
	if len(s) != 0 {
		initAddrs(s[1:])
		result += m.Run()
	}

	// test for multiple IP v6 address
	s = ""
	for _, addr := range addrs {
		if addr.To4() == nil && addr.To16() != nil {
			s = s + "/[" + addr.String() + "]"
		}
	}
	if len(s) != 0 {
		initAddrs(s[1:])
		result += m.Run()
	}

	if result > 0 {
		os.Exit(1)
	}
}

func initAddrs(addr string) {
	p := basePort
	for i := range testAddrs {
		testAddrs[i] = addr + ":" + strconv.Itoa(p)
		p += offsetPort
	}
}
