package main

import (
	"flag"
	"log"

	"github.com/fkgi/extnet"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	log.Println("starting simple echo server")
	// get option flag
	ip := flag.String("a", "127.0.0.1", "bind IP address")
	pt := flag.String("p", "10001", "bind port number")
	flag.Parse()

	addr, e := extnet.ResolveSCTPAddr(*ip + ":" + *pt)
	if e != nil {
		log.Fatal(e)
	}

	l, e := extnet.ListenSCTP(addr)
	if e != nil {
		log.Fatal(e)
	}

	for {
		c, e := l.AcceptSCTP()
		if e != nil {
			log.Println(e)
			continue
		}
		go func(c *extnet.SCTPConn) {
			buf := make([]byte, 1024)
			n, e := c.Read(buf)
			if e != nil {
				log.Println(e)
				return
			}
			log.Println(n)
			log.Println(buf[:n])
		}(c)
	}
}
