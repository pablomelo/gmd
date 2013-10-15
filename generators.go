package main

import (
	"fmt"
	"log"

	"github.com/peterbourgon/field"
)

// demoGenerator will output a sine wave at 440Hz.
type demoGenerator struct {
	id          string
	connects    chan connectRequest
	disconnects chan string
	connected   string
	output      chan []float32
	quit        chan chan struct{}
}

func newDemoGenerator(id string) *demoGenerator {
	g := &demoGenerator{
		id:          id,
		connects:    make(chan connectRequest),
		disconnects: make(chan string),
		connected:   "",
		output:      nil,
		quit:        make(chan chan struct{}),
	}
	go g.loop()
	return g
}

func (g *demoGenerator) stop() {
	q := make(chan struct{})
	g.quit <- q
	<-q
}

func (g *demoGenerator) loop() {
	log.Printf("%s: started", g.ID())
	defer log.Printf("%s: done", g.ID())

	var phase float32
	for {
		select {
		case g.output <- nextBuffer(sine, 440.0, &phase):
			//log.Printf("%s ♪", g.ID())
			break

		case r := <-g.connects:
			if g.output != nil {
				r.e <- fmt.Errorf("%s already connected to %s", g.ID(), g.connected)
				continue
			}
			g.connected = r.r.ID()
			g.output = make(chan []float32, 1)
			r.r.receive(g.output)
			r.e <- nil
			log.Printf("%s → %s", g.ID(), r.r.ID())

		case id := <-g.disconnects:
			if g.output == nil {
				log.Printf("%s: disconnect, but not connected", g.ID())
				continue
			}
			if g.connected != id {
				log.Printf("%s: connected to %s, not %s (bug in field)", g.ID(), g.connected, id)
				continue
			}
			g.connected = ""
			close(g.output)
			g.output = nil
			log.Printf("%s ✕ %s", g.ID(), id)

		case q := <-g.quit:
			close(q)
			return
		}
	}
}

func (g *demoGenerator) ID() string { return g.id }

func (g *demoGenerator) Connect(n field.Node) error {
	r, ok := n.(audioReceiver)
	if !ok {
		return fmt.Errorf("%s not audioReceiver", n.ID())
	}
	req := connectRequest{r, make(chan error)}
	g.connects <- req
	return <-req.e
}

func (g *demoGenerator) Disconnect(n field.Node) {
	if _, ok := n.(audioReceiver); !ok {
		log.Printf("%s not audioReceiver", n.ID())
		return
	}
	g.disconnects <- n.ID()
}

func (g *demoGenerator) Connection(n field.Node) error {
	log.Printf("%s: Connection(%s): no", g.ID(), n.ID())
	return errNo
}

func (g *demoGenerator) Disconnection(n field.Node) {
	log.Printf("%s: Disonnection(%s): ignored", g.ID(), n.ID())
}

type connectRequest struct {
	r audioReceiver
	e chan error
}
