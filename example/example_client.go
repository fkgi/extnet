package main

import (
	"flag"
	"log"

	"github.com/fkgi/extnet"
)

type ipList string

func (l *ipList) String() string {
	return string(*l)
}

func (l *ipList) Set(s string) error {
	*l = ipList(string(*l) + "/" + s)
	return nil
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	extnet.TraceEnable()

	log.Println("starting simple echo client")
	// get option flag
	var lips, rips ipList
	flag.Var(&lips, "la", "bind local IP address")
	flag.Var(&rips, "ra", "bind remote IP address")
	lpt := flag.String("lp", "10002", "bind local port number")
	rpt := flag.String("rp", "10001", "bind remote port number")
	flag.Parse()

	if len(lips) == 0 || len(rips) == 0 {
		log.Fatal("no IP address")
	}

	log.Print("creating address...")
	laddr, e := extnet.ResolveSCTPAddr(string(lips)[1:] + ":" + *lpt)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success as ", laddr, "(local)")

	raddr, e := extnet.ResolveSCTPAddr(string(rips)[1:] + ":" + *rpt)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success as ", raddr, "(remote)")

	log.Print("dialing...")
	c, e := extnet.DialSCTP(laddr, raddr)
	if e != nil {
		log.Fatal(e)
	}

	log.Print("new connection is available")
	log.Print(" local : ", c.LocalAddr())
	log.Print(" remote: ", c.RemoteAddr())

	buf := []byte("test message")
	log.Print("writing...")
	_, e = c.Write(buf)
	if e != nil {
		log.Print("ERROR:", e)
		c.Close()
		return
	}

	buf = make([]byte, 1024)
	log.Print("reading...")
	n, e := c.Read(buf)
	if e != nil {
		log.Print("ERROR: ", e)
		c.Close()
		return
	}
	log.Print("success as \"", string(buf[:n]), "\", length is ", n)

	log.Print("closing...")
	c.Close()
	if e != nil {
		log.Print("ERROR: ", e)
	}
	log.Print("success")
}
