package extnet

import "testing"

func TestAccept(t *testing.T) {
	l, c := dialTo(testAddrs[0], testAddrs[1], t)

	if lc, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc.LocalAddr().String() != testAddrs[1] {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), testAddrs[1])
		}
		if lc.RemoteAddr().String() != testAddrs[0] {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), testAddrs[0])
		}
	}

	closeAll(l, c, t)
}

func TestAcceptMulti(t *testing.T) {
	l, c1 := dialTo(testAddrs[0], testAddrs[1], t)

	if lc1, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc1.LocalAddr().String() != testAddrs[1] {
			t.Errorf("output %s is not same as %s",
				lc1.LocalAddr().String(), testAddrs[1])
		}
		if lc1.RemoteAddr().String() != testAddrs[0] {
			t.Errorf("output %s is not same as %s",
				lc1.RemoteAddr().String(), testAddrs[0])
		}
	}

	la, e := ResolveSCTPAddr("sctp", testAddrs[2])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	ra, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	c2, e := DialSCTP(la, ra)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	if c2.LocalAddr().String() != testAddrs[2] {
		t.Errorf("output %s is not same as %s",
			c2.LocalAddr().String(), testAddrs[2])
	}
	if c2.RemoteAddr().String() != testAddrs[1] {
		t.Errorf("output %s is not same as %s",
			c2.RemoteAddr().String(), testAddrs[1])
	}
	if lc2, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc2.LocalAddr().String() != testAddrs[1] {
			t.Errorf("output %s is not same as %s",
				lc2.LocalAddr().String(), testAddrs[1])
		}
		if lc2.RemoteAddr().String() != testAddrs[2] {
			t.Errorf("output %s is not same as %s",
				lc2.RemoteAddr().String(), testAddrs[2])
		}
	}

	c2.Close()
	closeAll(l, c1, t)
}

func TestConnect(t *testing.T) {
	l1 := listenOn(testAddrs[1], t)
	l2 := listenOn(testAddrs[0], t)

	ra, e := ResolveSCTPAddr("sctp", testAddrs[1])
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
		if lc.LocalAddr().String() != testAddrs[0] {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), testAddrs[0])
		}
		if lc.RemoteAddr().String() != testAddrs[1] {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), testAddrs[1])
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
	l, c1 := dialTo(testAddrs[0], testAddrs[1], t)

	if lc1, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc1.LocalAddr().String() != testAddrs[1] {
			t.Errorf("output %s is not same as %s",
				lc1.LocalAddr().String(), testAddrs[1])
		}
		if lc1.RemoteAddr().String() != testAddrs[0] {
			t.Errorf("output %s is not same as %s",
				lc1.RemoteAddr().String(), testAddrs[0])
		}
	}

	l2 := listenOn(testAddrs[2], t)
	ra, e := ResolveSCTPAddr("sctp", testAddrs[2])
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
	if c2.LocalAddr().String() != testAddrs[2] {
		t.Errorf("output %s is not same as %s",
			c2.LocalAddr().String(), testAddrs[2])
	}
	if c2.RemoteAddr().String() != testAddrs[1] {
		t.Errorf("output %s is not same as %s",
			c2.RemoteAddr().String(), testAddrs[1])
	}

	c2.Close()
	l2.Close()
	closeAll(l, c1, t)
}
