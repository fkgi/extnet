package main

import (
    "github.com/fkgi/extnet"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

log.Println("starting simple echo server")
	// get option flag
	addr = flag.String("a", "127.0.0.1", "bind IP address")
	port = flag.String("p", "10001", "bind port number")
	flag.Parse()

extnet.ResolveSCTPAddr(*addr + ":" + *port)


}
