package extnet

import "testing"

func TestListenSCTP(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	l0, e := ListenSCTP("sctp", a0)
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	s := l0.Addr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestListen(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	d := &SCTPDialer{LocalAddr: a0}
	l0, e := d.Listen()
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	s := l0.Addr().String()
	if s != testAddrs[0] {
		t.Errorf("output %s is not same as %s", s, testAddrs[0])
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestDialSCTP(t *testing.T) {
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
}

func TestDial(t *testing.T) {
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

	d := &SCTPDialer{LocalAddr: a1}
	c1, e := d.Dial("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("dial faied: %s", e)
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
}
