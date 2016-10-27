package extnet

import "testing"

func TestAccept(t *testing.T) {
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
	}

	closeAll(l, c, t)
}

func TestConnect(t *testing.T) {
	l1 := listenOn(remoteAddr, t)
	l2 := listenOn(localAddr, t)

	ra, e := ResolveSCTPAddr("sctp", remoteAddr)
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
		if lc.LocalAddr().String() != localAddr {
			t.Errorf("output %s is not same as %s",
				lc.LocalAddr().String(), localAddr)
		}
		if lc.RemoteAddr().String() != remoteAddr {
			t.Errorf("output %s is not same as %s",
				lc.RemoteAddr().String(), remoteAddr)
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
