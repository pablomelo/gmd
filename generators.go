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
	disconnects chan connectRequest
	connected   string
	output      chan []float32
	quit        chan chan struct{}
}

func newDemoGenerator(id string) *demoGenerator {
	g := &demoGenerator{
		id:          id,
		connects:    make(chan connectRequest),
		disconnects: make(chan connectRequest),
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

		case r := <-g.disconnects:
			if g.output == nil {
				r.e <- fmt.Errorf("%s not connected", g.ID())
				continue
			}
			if g.connected != r.r.ID() {
				r.e <- fmt.Errorf("%s connected to %s, not %s (bug in field)", g.ID(), g.connected, r.r.ID())
				continue
			}
			g.connected = ""
			close(g.output)
			g.output = nil
			r.e <- nil
			log.Printf("%s ✕ %s", g.ID(), r.r.ID())

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
	r, ok := n.(audioReceiver)
	if !ok {
		log.Printf("%s not audioReceiver", n.ID())
		return
	}
	req := connectRequest{r, make(chan error)}
	g.disconnects <- req
	<-req.e
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

/*
// audioSubscriptions satisfies audioSender methods by pushing requests through
// the subscriptions and unsubscriptions channels. It's designed to be embedded
// in a generator structure with a loop method that selects over the channels.
type audioSubscriptions struct {
	subscriptions   chan subscriptionRequest
	unsubscriptions chan string
}

func makeAudioSubscriptions() audioSubscriptions {
	return audioSubscriptions{
		subscriptions:   make(chan subscriptionRequest),
		unsubscriptions: make(chan string),
	}
}

type subscriptionRequest struct {
	id string
	c  chan []float32
}

func (s audioSubscriptions) subscribeAudio(id string) <-chan []float32 {
	log.Printf("audioSubscriptions: subscribeAudio(%s)", id)
	c := make(chan []float32)
	req := subscriptionRequest{id, c}
	s.subscriptions <- req
	return c
}

func (s audioSubscriptions) unsubscribeAudio(id string) {
	log.Printf("audioSubscriptions: unsubscribeAudio(%s)", id)
	s.unsubscriptions <- id
}
*/
