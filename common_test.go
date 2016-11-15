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

	if ipns, e := net.InterfaceAddrs(); e != nil {
		os.Exit(1)
	} else {
		for _, ipn := range ipns {
			if a, ok := (ipn).(*net.IPNet); !ok {
				os.Exit(1)
			} else if a.IP.To4() != nil {
				if a.IP.String() != "127.0.0.1" {
					addrs = append(addrs, a.IP)
				}
			} else if a.IP.To16() != nil {
				if a.IP.String() != "::1" {
					addrs = append(addrs, a.IP)
				}
			}
		}
	}

	// test for IP v4 address
	for _, addr := range addrs {
		if addr.To4() != nil {
			initAddrs(addr.String())
			break
		}
	}
	if m.Run() == 1 {
		os.Exit(1)
	}

	// test for multiple IP v4 address
	s := ""
	for _, addr := range addrs {
		if addr.To4() != nil {
			s = s + "/" + addr.String()
		}
	}
	initAddrs(s[1:])
	if m.Run() == 1 {
		os.Exit(1)
	}

	// test for IP v6 address
	for _, addr := range addrs {
		if addr.To4() == nil && addr.To16() != nil {
			initAddrs(addr.String())
			break
		}
	}
	if m.Run() == 1 {
		os.Exit(1)
	}

	// test for multiple IP v6 address
	s = ""
	for _, addr := range addrs {
		if addr.To4() == nil && addr.To16() != nil {
			s = s + "/" + addr.String()
		}
	}
	initAddrs(s[1:])
	os.Exit(m.Run())
}

func initAddrs(addr string) {
	p := basePort
	for i := range testAddrs {
		testAddrs[i] = addr + ":" + strconv.Itoa(p)
		p += offsetPort
	}
}

/*
func listenOn(addr string, t *testing.T) *SCTPListener {
	la, e := ResolveSCTPAddr("sctp", addr)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l, e := ListenSCTP("sctp", la)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	if l.Addr().String() != addr {
		t.Errorf("output %s is not same as %s",
			l.Addr().String(), addr)
	}

	return l
}

func dialTo(laddr, raddr string, t *testing.T) (*SCTPListener, *SCTPConn) {
	l := listenOn(raddr, t)

	la, e := ResolveSCTPAddr("sctp", laddr)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	ra, e := ResolveSCTPAddr("sctp", raddr)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	c, e := DialSCTP(la, ra)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	if c.LocalAddr().String() != laddr {
		t.Errorf("output %s is not same as %s",
			c.LocalAddr().String(), laddr)
	}
	if c.RemoteAddr().String() != raddr {
		t.Errorf("output %s is not same as %s",
			c.RemoteAddr().String(), raddr)
	}

	return l, c
}

func closeAll(l *SCTPListener, c *SCTPConn, t *testing.T) {
	e := c.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}
*/
