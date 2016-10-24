package main

import (
	"flag"
	"log"

	"github.com/fkgi/extnet"
	"github.com/fkgi/extnet/example"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	extnet.Notificator = func(e error) {
		log.Println(e)
	}
	log.Println("starting simple echo client")

	// get option flag
	var li, ri example.IPList
	flag.Var(&li, "la", "local IP address")
	flag.Var(&ri, "ra", "remote IP address")
	lp := flag.String("lp", "10002", "local port number")
	rp := flag.String("rp", "10001", "remote port number")
	flag.Parse()

	if len(li) == 0 || len(ri) == 0 {
		log.Fatal("no IP address")
	}

	log.Print("creating address...")
	laddr, e := extnet.ResolveSCTPAddr(string(li)[1:] + ":" + *lp)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success as ", laddr, "(local)")

	raddr, e := extnet.ResolveSCTPAddr(string(ri)[1:] + ":" + *rp)
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
	e = c.Close()
	if e != nil {
		log.Print("ERROR: ", e)
	}
	log.Print("success")
}
