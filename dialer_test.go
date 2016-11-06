package extnet

import "testing"

func TestListenSCTP(t *testing.T) {
	l := listenOn(testAddrs[1], t)
	e := l.Close()
	if e != nil {
		t.Errorf("close faied: %s", e)
	}
}

func TestDialSCTP(t *testing.T) {
	l, c := dialTo(testAddrs[0], testAddrs[1], t)
	closeAll(l, c, t)
}
