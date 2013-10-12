package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	var (
		listen = flag.String("listen", ":6543", "UDP listen address")
	)
	flag.Parse()

	listenAddr, err := net.ResolveUDPAddr("udp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer time.Sleep(100 * time.Millisecond)

	quit := make(chan struct{})
	defer close(quit)
	in := make(chan string)
	p := newPlatform()

	go rd(in, conn, quit)
	go wr(p, in, quit)

	<-interrupt()
}

func rd(out chan string, conn *net.UDPConn, quit chan struct{}) {
	log.Printf("rd: %s", conn.LocalAddr())
	defer log.Printf("rd: done")
	const maxSize = 4096
	for {
		select {
		case <-quit:
			return

		default:
			b := make([]byte, maxSize)
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, remoteAddr, err := conn.ReadFrom(b)
			if err != nil {
				continue
			}
			if n >= maxSize {
				log.Printf("%s: too big", remoteAddr)
				continue
			}
			out <- string(b[:n])
		}
	}
}

func wr(p *platform, in chan string, quit chan struct{}) {
	defer log.Printf("wr: done")
	for {
		select {
		case s := <-in:
			log.Printf("wr: %s %v", strings.TrimSpace(s), []byte(s))

		case <-quit:
			return
		}
	}
}

func interrupt() chan error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	e := make(chan error)
	go func() { log.Printf("interrupt: %s", <-c); e <- nil }()
	return e
}
