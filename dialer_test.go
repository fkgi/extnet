package extnet

import "testing"

func TestListenSCTP(t *testing.T) {
	l := listenOn(remoteAddr, t)
	e := l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
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

func TestDialSCTP(t *testing.T) {
	l, c := dialTo(localAddr, remoteAddr, t)
	closeAll(l, c, t)
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
