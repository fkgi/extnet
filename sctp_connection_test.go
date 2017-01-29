package extnet

import "testing"
import "time"

func TestReadWrite(t *testing.T) {
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

	loop := true
	time.AfterFunc(time.Second*time.Duration(10), func() { loop = false })

	for loop {
		n, e := c0.Write([]byte(testStr))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

		buf := make([]byte, 1024)
		n, e = c1.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

		n, e = c1.Write([]byte(testStr))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

		buf = make([]byte, 1024)
		n, e = c0.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}
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

func TestCloseFromInit(t *testing.T) {
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

	_, e = l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
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

func TestAbortFromInit(t *testing.T) {
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

	_, e = l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	e = c1.Abort("test abort")
	if e != nil {
		t.Errorf("abort faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestCloseFromResp(t *testing.T) {
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

	_, e = DialSCTP(a1, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	c0, e := l0.Accept()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	e = c0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAbortFromResp(t *testing.T) {
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

	_, e = DialSCTP(a1, a0)
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}

	c0, e := l0.AcceptSCTP()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	e = c0.Abort("test abort")
	if e != nil {
		t.Errorf("abort faied: %s", e)
	}

	e = l0.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}
