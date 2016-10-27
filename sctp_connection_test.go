package extnet

import "testing"

func TestReadWrite(t *testing.T) {
	l, c := dialTo(localAddr, remoteAddr, t)

	if lc, e := l.Accept(); e != nil {
		t.Errorf("accept faied: %s", e)
	} else {
		if lc.LocalAddr().String() != remoteAddr {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), remoteAddr)
		}
		if lc.RemoteAddr().String() != localAddr {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), localAddr)
		}

		n, e := lc.Write([]byte(testString))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testString) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testString))
		}
		buf := make([]byte, 1024)
		n, e = c.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testString) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testString))
		}

		n, e = c.Write([]byte(testString))
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testString) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testString))
		}
		buf = make([]byte, 1024)
		n, e = lc.Read(buf)
		if e != nil {
			t.Errorf("write data failed: %s", e)
		}
		if n != len(testString) {
			t.Errorf("write data length is invalid: %d is not equal %d", n, len(testString))
		}

	}

	closeAll(l, c, t)
}
