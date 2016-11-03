package extnet

import "testing"

func TestListenSCTP(t *testing.T) {
	l := listenOn(addr2, t)
	e := l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestDialSCTP(t *testing.T) {
	l, c := dialTo(addr1, addr2, t)
	closeAll(l, c, t)
}
