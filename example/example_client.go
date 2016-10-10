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
	*l = ipList(string(*l) + "," + s)
	return nil
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	log.Println("starting simple echo client")
	// get option flag
	var lips, rips ipList
	flag.Var(&lips, "la", "bind IP address")
	flag.Var(&rips, "ra", "bind IP address")
	lpt := flag.String("lp", "10002", "bind port number")
	rpt := flag.String("rp", "10001", "bind port number")
	flag.Parse()

	if len(lips) == 0 || len(rips) == 0 {
		log.Fatal("no IP address")
	}

	log.Println("crreate address")
	laddr, e := extnet.ResolveSCTPAddr(string(lips)[1:] + ":" + *lpt)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("address is " + laddr.String())
	raddr, e := extnet.ResolveSCTPAddr(string(rips)[1:] + ":" + *rpt)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("address is " + raddr.String())

	log.Println("dial")
	c, e := extnet.DialSCTP(laddr, raddr)
	if e != nil {
		log.Fatal(e)
	}

	log.Println("new connection")
	log.Println("local:" + c.LocalAddr().String())
	log.Println("remote:" + c.RemoteAddr().String())

	buf := []byte("test message")
	log.Panicln("write")
	_, e = c.Write(buf)
	if e != nil {
		log.Println(e)
	}

	buf = make([]byte, 1024)
	log.Println("read")
	n, e := c.Read(buf)
	if e != nil {
		log.Println(e)
		c.Close()
		return
	}
	log.Println(n)
	log.Println(buf[:n])

	log.Panicln("close")
	c.Close()
}
