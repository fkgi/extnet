package extnet

import (
	"log"
	"os"
	"testing"
)

const (
	addr1   = "127.0.0.1:10000"
	addr2   = "127.0.0.1:10010"
	addr3   = "127.0.0.1:10020"
	addr4   = "127.0.0.1:10030"
	testStr = "this is test"
)

func TestMain(m *testing.M) {
	Notificator = func(e error) { log.Println(e) }
	// initiate
	code := m.Run()
	// destract
	os.Exit(code)
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
