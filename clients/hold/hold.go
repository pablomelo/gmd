package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	var (
		gmd  = flag.String("gmd", "localhost:5432", "gmd UDP endpoint")
		tgt  = flag.String("tgt", "f", "target of keypress")
		note = flag.Int("note", 69, "MIDI note to press")
		mod  = flag.Int("mod", 1, "modulo tick (0=none)")
		dur  = flag.Duration("for", 0, "duration (0=forever)")
		wait = flag.Duration("wait", 0, "time to wait afterwards")
	)
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *gmd)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	func() {
		down := fmt.Sprintf("send %s keydown %d", *tgt, *note)
		up := fmt.Sprintf("send %s keyup %d", *tgt, *note)
		if *mod > 0 {
			down = fmt.Sprintf("%% %d %s", *mod, down)
			up = fmt.Sprintf("%% %d %s", *mod, up)
		}

		conn.Write([]byte(down))
		defer conn.Write([]byte(up))

		if *dur > 0 {
			time.Sleep(*dur)
		} else {
			<-interrupt()
		}
	}()

	time.Sleep(*wait)
}

func interrupt() chan error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	e := make(chan error)
	go func() { log.Printf("interrupt: %s", <-c); e <- nil }()
	return e
}
