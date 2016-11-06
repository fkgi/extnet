package extnet

import (
	"log"
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
	Notificator = func(e error) { log.Println(e) }

	addrs, e := net.InterfaceAddrs()
	if e != nil {
		os.Exit(1)
	}
	for _, addr := range addrs {
		a, ok := (addr).(*net.IPNet)
		if !ok {
			os.Exit(1)
		}
		if a.IP.To4() == nil || a.IP.String() == "127.0.0.1" {
			continue
		}
		p := basePort
		for i := range testAddrs {
			testAddrs[i] = a.IP.String() + ":" + strconv.Itoa(p)
			p += offsetPort
		}
		break
	}
	if m.Run() == 1 {
		os.Exit(1)
	}

	p := basePort
	for i := range testAddrs {
		testAddrs[i] = ":" + strconv.Itoa(p)
		p += offsetPort
	}
	for _, addr := range addrs {
		a, ok := (addr).(*net.IPNet)
		if !ok {
			os.Exit(1)
		}
		if a.IP.To4() == nil || a.IP.String() == "127.0.0.1" {
			continue
		}
		for i := range testAddrs {
			testAddrs[i] = "/" + a.IP.String() + testAddrs[i]
		}
	}
	for i := range testAddrs {
		testAddrs[i] = testAddrs[i][1:]
	}
	if m.Run() == 1 {
		os.Exit(1)
	}

	for _, addr := range addrs {
		a, ok := (addr).(*net.IPNet)
		if !ok {
			os.Exit(1)
		}
		if a.IP.To4() != nil || a.IP.To16() == nil || a.IP.String() == "::1" {
			continue
		}
		p = basePort
		for i := range testAddrs {
			testAddrs[i] = a.IP.String() + ":" + strconv.Itoa(p)
			p += offsetPort
		}
	}
	if m.Run() == 1 {
		os.Exit(1)
	}

	p = basePort
	for i := range testAddrs {
		testAddrs[i] = ":" + strconv.Itoa(p)
		p += offsetPort
	}
	for _, addr := range addrs {
		a, ok := (addr).(*net.IPNet)
		if !ok {
			os.Exit(1)
		}
		if a.IP.To4() != nil || a.IP.To16() == nil || a.IP.String() == "::1" {
			continue
		}
		for i := range testAddrs {
			testAddrs[i] = "/" + a.IP.String() + testAddrs[i]
		}
	}
	for i := range testAddrs {
		testAddrs[i] = testAddrs[i][1:]
	}
	os.Exit(m.Run())
}

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
