package main

import (
	"log"
)

type commandBuffer struct {
	parser parser
	buffer map[uint64][]string

	requests chan queueRequest
	ticks    chan uint64
	quit     chan chan struct{}
}

func newCommandBuffer(c *clock, p parser) *commandBuffer {
	b := &commandBuffer{
		parser: p,
		buffer: map[uint64][]string{},

		requests: make(chan queueRequest),
		ticks:    make(chan uint64),
		quit:     make(chan chan struct{}),
	}
	go b.loop(c)
	return b
}

func (b *commandBuffer) loop(c *clock) {
	log.Printf("cmdbuf: started")
	defer log.Printf("cmdbuf: done")

	c.subscribe(b)
	defer c.unsubscribe(b)

	for {
		select {
		case req := <-b.requests:
			b.buffer[req.modulo] = append(b.buffer[req.modulo], req.command)

		case tick := <-b.ticks:
			log.Printf("cmdbuf: tick %d pending %d", tick, len(b.buffer))
			b.buffer = fwd(b.parser, b.buffer, tick)

		case q := <-b.quit:
			close(q)
			return
		}
	}
}

func (b *commandBuffer) ID() string { return "cmdbuf" }

func (b *commandBuffer) tick(tick uint64) {
	b.ticks <- tick
}

func (b *commandBuffer) queue(modulo uint64, command string) {
	b.requests <- queueRequest{modulo, command}
}

func (b *commandBuffer) stop() {
	q := make(chan struct{})
	b.quit <- q
	<-q
}

type queueRequest struct {
	modulo  uint64
	command string
}

func fwd(p parser, buffer map[uint64][]string, tick uint64) map[uint64][]string {
	toParse, survivors := []string{}, map[uint64][]string{}
	for modulo, commands := range buffer {
		if tick%modulo == 0 {
			toParse = append(toParse, commands...)
		} else {
			survivors[modulo] = commands
		}
	}

	for _, command := range toParse {
		p.parse(command)
	}

	return survivors
}
