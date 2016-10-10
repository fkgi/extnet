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

	log.Println("starting simple echo server")
	// get option flag
	var ips ipList
	flag.Var(&ips, "a", "bind IP address")
	pt := flag.String("p", "10001", "bind port number")
	flag.Parse()

	if len(ips) == 0 {
		log.Fatal("no IP address")
	}

	log.Println("crreate address")
	addr, e := extnet.ResolveSCTPAddr(string(ips)[1:] + ":" + *pt)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("address is " + addr.String())

	log.Println("listen")
	l, e := extnet.ListenSCTP(addr)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("listening on " + l.Addr().String())

	for {
		log.Println("accept")
		c, e := l.AcceptSCTP()
		if e != nil {
			log.Println(e)
			continue
		}

		go func(c *extnet.SCTPConn) {
			log.Println("new connection")
			log.Println("local:" + c.LocalAddr().String())
			log.Println("remote:" + c.RemoteAddr().String())

			buf := make([]byte, 1024)
			log.Println("read")
			n, e := c.Read(buf)
			if e != nil {
				log.Println(e)
				c.Close()
				return
			}
			log.Println(n)
			log.Println(buf[:n])

			log.Panicln("write")
			_, e = c.Write(buf)
			if e != nil {
				log.Println(e)
			}

			log.Panicln("close")
			c.Close()
		}(c)
	}
}
