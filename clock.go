package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/peterbourgon/field"
)

type tickReceiver interface {
	identifier
	tick(uint64)
}

type clock struct {
	subs map[string]tickReceiver

	newBPM          chan float32
	subscriptions   chan tickReceiver
	unsubscriptions chan tickReceiver
	quit            chan chan struct{}
}

func newClock(bpm float32) *clock {
	c := &clock{
		subs: map[string]tickReceiver{},

		newBPM:          make(chan float32),
		subscriptions:   make(chan tickReceiver),
		unsubscriptions: make(chan tickReceiver),
		quit:            make(chan chan struct{}),
	}
	go c.loop(bpm)
	return c
}

func (c *clock) loop(bpm float32) {
	log.Printf("clock: started")
	defer log.Printf("clock: done")

	t, n := time.NewTicker(bpm2duration(bpm)), uint64(0)
	for {
		select {
		case <-t.C:
			//log.Printf("clock: ⦿ (%d → %d)", n, len(c.subs))
			for _, sub := range c.subs {
				sub.tick(n)
			}
			n++

		case bpm := <-c.newBPM:
			log.Printf("clock: %.2f", bpm)
			t.Stop()
			t = time.NewTicker(bpm2duration(bpm))

		case r := <-c.subscriptions:
			if _, ok := c.subs[r.ID()]; ok {
				log.Printf("clock: double-subscribe %s", r.ID())
				return
			}
			c.subs[r.ID()] = r

		case r := <-c.unsubscriptions:
			if _, ok := c.subs[r.ID()]; !ok {
				log.Printf("clock: %s not found to unsubscribe", r.ID())
				return
			}
			delete(c.subs, r.ID())

		case q := <-c.quit:
			close(q)
			return
		}
	}
}

func (c *clock) subscribe(r tickReceiver) {
	c.subscriptions <- r
}

func (c *clock) unsubscribe(r tickReceiver) {
	c.unsubscriptions <- r
}

func (c *clock) parse(input string) {
	input = strings.TrimSpace(strings.ToLower(input))
	toks := strings.Split(input, " ")
	if len(toks) <= 0 {
		log.Printf("clock: parse empty")
		return
	}

	switch toks[0] {
	case "bpm":
		if len(toks) != 2 {
			log.Printf("clock: %s: bad args", input)
			return
		}
		bpm, err := strconv.ParseFloat(toks[1], 32)
		if err != nil {
			log.Printf("clock: %s: %s", input, err)
			return
		}
		c.newBPM <- float32(bpm)

	default:
		log.Printf("clock: %s: aroo", input)
	}
}

func (c *clock) stop() {
	q := make(chan struct{})
	c.quit <- q
	<-q
}

func (c *clock) ID() string                  { return "clock" }
func (c *clock) Connect(field.Node) error    { return errNo }
func (c *clock) Connection(field.Node) error { return errNo }
func (c *clock) Disconnect(field.Node)       {}
func (c *clock) Disconnection(field.Node)    {}

func bpm2duration(bpm float32) time.Duration {
	return time.Duration((60.0 / bpm) * float32(time.Second))
}
