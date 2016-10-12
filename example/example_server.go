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

	log.Println("starting simple echo server")
	// get option flag
	var ips ipList
	flag.Var(&ips, "a", "bind IP address")
	pt := flag.String("p", "10001", "bind port number")
	flag.Parse()

	if len(ips) == 0 {
		log.Fatal("ERROR: no IP address")
	}

	log.Print("creating address...")
	addr, e := extnet.ResolveSCTPAddr(string(ips)[1:] + ":" + *pt)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success as ", addr)

	log.Print("starting listen... ")
	l, e := extnet.ListenSCTP(addr)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success on ", l.Addr())

	for {
		log.Print("accepting...")
		c, e := l.AcceptSCTP()
		if e != nil {
			log.Println(e)
			continue
		}

		go func(c *extnet.SCTPConn) {
			log.Print("new connection is available")
			log.Print(" local : ", c.LocalAddr())
			log.Print(" remote: ", c.RemoteAddr())

			buf := make([]byte, 1024)
			log.Print("reading...")
			n, e := c.Read(buf)
			if e != nil {
				log.Print("ERROR:", e)
				c.Close()
				return
			}
			log.Print("success as \"", string(buf[:n]), "\", length is ", n)

			log.Print("writing...")
			_, e = c.Write(buf[:n])
			if e != nil {
				log.Print("ERROR: ", e)
			}

			log.Print("closing...")
			c.Close()
			if e != nil {
				log.Print("ERROR: ", e)
			}
			log.Print("success")
		}(c)
	}
}
