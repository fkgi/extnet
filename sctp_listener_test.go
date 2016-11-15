package extnet

import "testing"

func TestAccept(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l0, e := ListenSCTP("sctp", a0)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}

	c1, e := DialSCTP(a1, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	c0, e := l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}
	s := c0.LocalAddr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}
	s = c0.RemoteAddr().String()
	if s != testAddrs[1] {
		t.Errorf("output %s is not same as %s", s, testAddrs[1])
	}

	e = c1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAcceptMulti(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a2, e := ResolveSCTPAddr("sctp", testAddrs[2])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l0, e := ListenSCTP("sctp", a0)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}

	c1, e := DialSCTP(a1, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	_, e = l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	c2, e := DialSCTP(a2, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}
	s := c2.LocalAddr().String()
	if s != testAddrs[2] {
		t.Errorf("output %s is not same as %s", s, testAddrs[2])
	}
	s = c2.RemoteAddr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}

	c0_2, e := l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}
	s = c0_2.LocalAddr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}
	s = c0_2.RemoteAddr().String()
	if s != testAddrs[2] {
		t.Errorf("output %s is not same as %s", s, testAddrs[2])
	}

	e = c1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = c2.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestConnect(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l0, e := ListenSCTP("sctp", a0)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	l1, e := ListenSCTP("sctp", a1)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}

	e = l1.Connect(a0)
	if e != nil {
		t.Errorf("connect failed: %s", e)
	}

	_, e = l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	c1, e := l1.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}
	s := c1.LocalAddr().String()
	if s != testAddrs[1] {
		t.Errorf("output %s is not same as %s", s, testAddrs[1])
	}
	s = c1.RemoteAddr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}

	e = c1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAcceptConnect(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a2, e := ResolveSCTPAddr("sctp", testAddrs[2])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l0, e := ListenSCTP("sctp", a0)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}

	c1, e := DialSCTP(a1, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	_, e = l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	l2, e := ListenSCTP("sctp", a2)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}

	e = l0.Connect(a2)
	if e != nil {
		t.Errorf("connect failed: %s", e)
	}

	c2, e := l2.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}
	s := c2.LocalAddr().String()
	if s != testAddrs[2] {
		t.Errorf("output %s is not same as %s", s, testAddrs[2])
	}
	s = c2.RemoteAddr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}

	e = c2.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l2.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = c1.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}
