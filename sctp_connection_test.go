package extnet

import (
	"testing"
	"time"
)

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

	buf := make([]byte, 1024)
	loop := true
	time.AfterFunc(time.Second*time.Duration(1), func() { loop = false })

	for loop {
		if n, e := c0.Write([]byte(testStr)); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		} else if n, e = c1.Read(buf); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

		if n, e := c1.Write([]byte(testStr)); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		} else if n, e = c0.Read(buf); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
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

func TestReadWriteWithStream(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	d0 := &SCTPDialer{
		LocalAddr: a0,
		OutStream: 128,
		InStream:  64}
	l0a, e := d0.Listen()
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	l0 := l0a.(*SCTPListener)

	d1 := &SCTPDialer{
		LocalAddr: a1,
		OutStream: 128,
		InStream:  64}
	c1a, e := d1.Dial("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}
	c1 := c1a.(*SCTPConn)

	c0, e := l0.AcceptSCTP()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	buf := make([]byte, 1024)

	for i := 0; i < 63; i++ {
		if n, e := c0.WriteToStream([]byte(testStr), uint16(i), uint32(i)); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		} else if n, e = c1.Read(buf); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}
		i++

		if n, e := c1.WriteToStream([]byte(testStr), uint16(i), uint32(i)); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		} else if n, e = c0.Read(buf); e != nil {
			t.Errorf("write data failed: %s", e)
		} else if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}
	}

	if _, e := c0.WriteToStream([]byte(testStr), 64, 64); e == nil {
		t.Errorf("data writing must be failed")
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

func TestReadWriteUnordered(t *testing.T) {
	Notificator = func(e error) { t.Log(e) }

	a0, e := ResolveSCTPAddr("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}
	a1, e := ResolveSCTPAddr("sctp", testAddrs[1])
	if e != nil {
		t.Fatalf("address generation failure: %s", e)
	}

	d0 := &SCTPDialer{
		LocalAddr: a0,
		Unordered: true}
	l0a, e := d0.Listen()
	if e != nil {
		t.Fatalf("listen faied: %s", e)
	}
	l0 := l0a.(*SCTPListener)

	d1 := &SCTPDialer{
		LocalAddr: a1,
		Unordered: true}
	c1a, e := d1.Dial("sctp", testAddrs[0])
	if e != nil {
		t.Fatalf("dial faied: %s", e)
	}
	c1 := c1a.(*SCTPConn)

	c0, e := l0.AcceptSCTP()
	if e != nil {
		t.Errorf("accept faied: %s", e)
	}

	buf := make([]byte, 1024)

	if n, e := c0.Write([]byte(testStr)); e != nil {
		t.Errorf("write data failed: %s", e)
	} else if n != len(testStr) {
		t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
	} else if n, e = c1.Read(buf); e != nil {
		t.Errorf("write data failed: %s", e)
	} else if n != len(testStr) {
		t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
	}

	if n, e := c1.Write([]byte(testStr)); e != nil {
		t.Errorf("write data failed: %s", e)
	} else if n != len(testStr) {
		t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
	} else if n, e = c0.Read(buf); e != nil {
		t.Errorf("write data failed: %s", e)
	} else if n != len(testStr) {
		t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
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
