package extnet

import "testing"

func TestResolveSCTPAddr(t *testing.T) {
	str := testAddrs[0]
	if a, e := ResolveSCTPAddr("sctp", str); e != nil {
		t.Errorf("failure in single ipv4 address: %s", e)
	} else if a.String() != str {
		t.Errorf("output %s is not same as %s", a.String(), str)
	}
}

func TestResolveSCTPAddrInvaridNetwork(t *testing.T) {
	str := testAddrs[0]
	if _, e := ResolveSCTPAddr("aaa", str); e == nil {
		t.Error("no failure in invalid network case")
	}
}

func TestResolveSCTPAddrInvaridString(t *testing.T) {
	str := "invalid"
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in invalid address case")
	}
}

func TestResolveSCTPAddrInvaridPort(t *testing.T) {
	str := testAddrs[0] + "999999999"
	if a, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Logf("%s", a)
		t.Errorf("no failure in invalid port number case %s", str)
	}
}

func TestResolveSCTPAddrInvaridAddress(t *testing.T) {
	str := "zzzzzzzz" + testAddrs[0]
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in invalid ip address case")
	}
}

func TestResolveSCTPAddrNetworkMissmatch(t *testing.T) {
	str := testAddrs[0]
	if str[0] == '[' {
		if _, e := ResolveSCTPAddr("sctp4", str); e == nil {
			t.Error("no failure in network mismatch case")
		}
	} else {
		if _, e := ResolveSCTPAddr("sctp6", str); e == nil {
			t.Error("no failure in network mismatch case")
		}
	}
}

func TestResolveSCTPAddrAddressMissmatch(t *testing.T) {
	str := testAddrs[0]
	if str[0] == '[' {
		str = "127.0.0.1/" + str
	} else {
		str = "[::1]/" + str
	}
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in version mismatch case")
	}
}
