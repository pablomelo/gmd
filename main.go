package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	cycle = 50 * time.Millisecond
)

func main() {
	log.SetFlags(log.Lshortfile)
	var (
		listen = flag.String("listen", ":6543", "UDP listen address")
	)
	flag.Parse()

	p, err := newPlatform()
	if err != nil {
		log.Fatal(err)
	}
	defer p.stop()

	listenAddr, err := net.ResolveUDPAddr("udp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer time.Sleep(2 * cycle)
	quit := make(chan struct{})
	defer close(quit)
	in := make(chan string)

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
			conn.SetReadDeadline(time.Now().Add(cycle))
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
			p.parse(s)

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
