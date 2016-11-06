package extnet

import "testing"

func TestReadWrite(t *testing.T) {
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

		n, e := lc.Write([]byte(testStr))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}
		buf := make([]byte, 1024)
		n, e = c.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

		n, e = c.Write([]byte(testStr))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}
		buf = make([]byte, 1024)
		n, e = lc.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testStr) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testStr))
		}

	}

	closeAll(l, c, t)
}

func TestCloseFromInit(t *testing.T) {
	l, c := dialTo(testAddrs[0], testAddrs[1], t)
	e := c.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}

	e = l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAbortFromInit(t *testing.T) {
	l, c := dialTo(testAddrs[0], testAddrs[1], t)
	e := c.Abort("test abort")
	if e != nil {
		t.Errorf("abort faied: %s", e)
	}

	e = l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestCloseFromResp(t *testing.T) {
	l, _ := dialTo(testAddrs[0], testAddrs[1], t)
	if lc, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		e = lc.Close()
		if e != nil {
			t.Errorf("close faied: %s", e)
		}
	}
	e := l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestAbortFromResp(t *testing.T) {
	l, _ := dialTo(testAddrs[0], testAddrs[1], t)
	if lc, e := l.AcceptSCTP(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		e = lc.Abort("test abort")
		if e != nil {
			t.Errorf("abort faied: %s", e)
		}
	}
	e := l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}
