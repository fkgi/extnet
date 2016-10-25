package extnet

import "testing"

func TestResolveSCTPAddr(t *testing.T) {
	// invalid input
	str := "127.0.0.1:10000"
	if _, e := ResolveSCTPAddr("aaa", str); e == nil {
		t.Error("no failure in invalid network case")
	}

	str = "invalid"
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in invalid address case")
	}

	str = "127.0.0.1:invalid"
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in invalid port number case")
	}

	str = "a:10000"
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in invalid ip address case")
	}

	str = "127.0.0.1:10000"
	if _, e := ResolveSCTPAddr("sctp6", str); e == nil {
		t.Error("no failure in network mismatch case")
	}

	str = "[::1]:10000"
	if _, e := ResolveSCTPAddr("sctp4", str); e == nil {
		t.Error("no failure in network mismatch case")
	}

	str = "127.0.0.1/[::1]:10000"
	if _, e := ResolveSCTPAddr("sctp", str); e == nil {
		t.Error("no failure in version mismatch case")
	}

	// valid input
	str = "127.0.0.1:10000"
	if a, e := ResolveSCTPAddr("sctp", str); e != nil {
		t.Errorf("failure in single ipv4 address: %s", e)
	} else if a.String() != str {
		t.Errorf("output %s is not same as %s", a.String(), str)
	}

	str = "127.0.0.1/127.0.0.2:10000"
	if a, e := ResolveSCTPAddr("sctp", str); e != nil {
		t.Errorf("failure in multi ipv4 address: %s", e)
	} else if a.String() != str {
		t.Errorf("output %s is not same as %s", a.String(), str)
	}

	str = "[::1]:10000"
	if a, e := ResolveSCTPAddr("sctp", str); e != nil {
		t.Errorf("failure in single ipv6 address: %s", e)
	} else if a.String() != str {
		t.Errorf("output %s is not same as %s", a.String(), str)
	}

	str = "[::1]/[::2]:10000"
	if a, e := ResolveSCTPAddr("sctp", str); e != nil {
		t.Errorf("failure in multi ipv6 address: %s", e)
	} else if a.String() != str {
		t.Errorf("output %s is not same as %s", a.String(), str)
	}

}
