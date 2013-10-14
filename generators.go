package main

import (
	"fmt"
	"github.com/peterbourgon/field"
	"log"
	"reflect"
)

// demoGenerator will output a sine wave at 440Hz.
type demoGenerator struct {
	noopNode
	audioSubscriptions
	output map[string]chan []float32 // active audio output channels
	quit   chan chan struct{}
}

func newDemoGenerator(id string) *demoGenerator {
	g := &demoGenerator{
		noopNode:           noopNode(id),
		audioSubscriptions: makeAudioSubscriptions(),
		output:             map[string]chan []float32{},
		quit:               make(chan chan struct{}),
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
	defer log.Printf("%s: done", g.ID())

	gen := make(chan []float32)
	makeSoundQuit := make(chan chan struct{})
	go makeSound(gen, sine, 440.0, makeSoundQuit)

	for {
		select {
		case buffer := <-gen:
			g.demux(g.output, buffer)

		case req := <-g.audioSubscriptions.subscriptions:
			if _, ok := g.output[req.id]; ok {
				panic(fmt.Sprintf("%s: double-subscribe of '%s'", g.ID(), req.id))
			}
			g.output[req.id] = req.c

		case id := <-g.audioSubscriptions.unsubscriptions:
			if _, ok := g.output[id]; !ok {
				panic(fmt.Sprintf("%s: impossible unsubscribe of '%s'", g.ID(), id))
			}
			delete(g.output, id)

		case q := <-g.quit:
			qq := make(chan struct{})
			makeSoundQuit <- qq
			<-qq
			close(q)
			return
		}
	}
}

func (g *demoGenerator) demux(dst map[string]chan []float32, src []float32) {
	for id, c := range dst {
		select {
		case c <- src:
			// OK
		default:
			log.Printf("%s: demux to '%s': fail", g.ID(), id)
		}
	}
}

func makeSound(c chan []float32, f generatorFunction, hz float32, quit chan chan struct{}) {
	defer log.Printf("makeSound %s %.2f done", reflect.ValueOf(f).Kind(), hz)
	var phase float32
	buf := make([]float32, bufSz)
	for {
		for i := 0; i < bufSz; i++ {
			buf[i] = nextGeneratorFunctionValue(f, hz, &phase)
		}

		select {
		case c <- buf:
			// OK
		case q := <-quit:
			close(q)
			return
		}
	}
}

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

func (s audioSubscriptions) subscribeAudio(id string) chan []float32 {
	c := make(chan []float32)
	req := subscriptionRequest{id, c}
	s.subscriptions <- req
	return c
}

func (s audioSubscriptions) unsubscribeAudio(id string) {
	s.unsubscriptions <- id
}

// noopNode satisfies field.Node methods, returning the string as the ID, and
// doing nothing on all events. It's designed to be embedded into generators,
// which can implement the specific event methods they care about.
type noopNode string

func (n noopNode) ID() string               { return string(n) }
func (n noopNode) Connect(field.Node)       { log.Printf("%s: Connect (noop)", n) }
func (n noopNode) Disconnect(field.Node)    { log.Printf("%s: Disconnect (noop)", n) }
func (n noopNode) Connection(field.Node)    { log.Printf("%s: Connection (noop)", n) }
func (n noopNode) Disconnection(field.Node) { log.Printf("%s: Disconnection (noop)", n) }
