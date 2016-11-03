package extnet

import "testing"

func TestAccept(t *testing.T) {
	l, c := dialTo(addr1, addr2, t)

	if lc, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc.LocalAddr().String() != addr2 {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), addr2)
		}
		if lc.RemoteAddr().String() != addr1 {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), addr1)
		}
	}

	closeAll(l, c, t)
}

func TestAcceptMulti(t *testing.T) {
	l, c1 := dialTo(addr1, addr2, t)

	if lc1, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc1.LocalAddr().String() != addr2 {
			t.Errorf("output %s is not same as %s",
				lc1.LocalAddr().String(), addr2)
		}
		if lc1.RemoteAddr().String() != addr1 {
			t.Errorf("output %s is not same as %s",
				lc1.RemoteAddr().String(), addr1)
		}
	}

	la, e := ResolveSCTPAddr("sctp", addr3)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	ra, e := ResolveSCTPAddr("sctp", addr2)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	c2, e := DialSCTP(la, ra)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	if c2.LocalAddr().String() != addr3 {
		t.Errorf("output %s is not same as %s",
			c2.LocalAddr().String(), addr3)
	}
	if c2.RemoteAddr().String() != addr2 {
		t.Errorf("output %s is not same as %s",
			c2.RemoteAddr().String(), addr2)
	}
	if lc2, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc2.LocalAddr().String() != addr2 {
			t.Errorf("output %s is not same as %s",
				lc2.LocalAddr().String(), addr2)
		}
		if lc2.RemoteAddr().String() != addr3 {
			t.Errorf("output %s is not same as %s",
				lc2.RemoteAddr().String(), addr3)
		}
	}

	c2.Close()
	closeAll(l, c1, t)
}

func TestConnect(t *testing.T) {
	l1 := listenOn(addr2, t)
	l2 := listenOn(addr1, t)

	ra, e := ResolveSCTPAddr("sctp", addr2)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	e = l2.Connect(ra)
	if e != nil {
		t.Errorf("connect failed: %s", e)
	}
	if lc, e := l2.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc.LocalAddr().String() != addr1 {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), addr1)
		}
		if lc.RemoteAddr().String() != addr2 {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), addr2)
		}
		e = lc.Close()
		if e != nil {
			t.Errorf("close faied: %s", e)
		}
	}

	e = l1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
	e = l2.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAcceptConnect(t *testing.T) {
	l, c1 := dialTo(addr1, addr2, t)

	if lc1, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc1.LocalAddr().String() != addr2 {
			t.Errorf("output %s is not same as %s",
				lc1.LocalAddr().String(), addr2)
		}
		if lc1.RemoteAddr().String() != addr1 {
			t.Errorf("output %s is not same as %s",
				lc1.RemoteAddr().String(), addr1)
		}
	}

	l2 := listenOn(addr3, t)
	ra, e := ResolveSCTPAddr("sctp", addr3)
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	e = l.Connect(ra)
	if e != nil {
		t.Errorf("connect failed: %s", e)
	}
	c2, e := l2.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}
	if c2.LocalAddr().String() != addr3 {
		t.Errorf("output %s is not same as %s",
			c2.LocalAddr().String(), addr3)
	}
	if c2.RemoteAddr().String() != addr2 {
		t.Errorf("output %s is not same as %s",
			c2.RemoteAddr().String(), addr2)
	}

	c2.Close()
	l2.Close()
	closeAll(l, c1, t)
}
