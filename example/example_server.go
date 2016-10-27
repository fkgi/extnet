package main

import (
	"flag"
	"log"

	"github.com/fkgi/extnet"
)

type IPList string

func (l *IPList) String() string {
	return string(*l)
}

func (l *IPList) Set(s string) error {
	*l = IPList(string(*l) + "/" + s)
	return nil
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	extnet.Notificator = func(e error) {
		log.Println(e)
	}
	log.Println("starting simple echo server")

	// get olpion flag
	var li IPList
	flag.Var(&li, "la", "local IP address")
	lp := flag.String("lp", "10001", "local port number")
	flag.Parse()

	if len(li) == 0 {
		log.Fatal("ERROR: no IP address")
	}

	log.Print("creating address...")
	addr, e := extnet.ResolveSCTPAddr("sctp", string(li)[1:]+":"+*lp)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success as ", addr)

	log.Print("starting listen... ")
	l, e := extnet.ListenSCTP("sctp", addr)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("success on ", l.Addr())

	for {
		log.Print("accelping...")
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
