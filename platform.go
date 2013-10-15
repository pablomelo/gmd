package main

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/peterbourgon/field"
)

var (
	errNo = errors.New("no")
)

type parser interface {
	parse(input string)
}

// platform holds the music objects.
type platform struct {
	mixer  *mixer
	field  *field.Field
	clock  *clock
	buffer *commandBuffer
}

func newPlatform() (*platform, error) {
	p := &platform{}

	var err error
	p.mixer, err = newMixer()
	if err != nil {
		return nil, err
	}

	p.clock = newClock(120.0)
	p.buffer = newCommandBuffer(p.clock, p)

	p.field = field.New()
	p.field.AddNode(p.mixer) // a platform always has a permanent mixer
	p.field.AddNode(p.clock) // a platform always has a permanent clock

	return p, nil
}

func (p *platform) stop() {
	p.mixer.stop()
	p.buffer.stop()
	p.clock.stop()
	log.Printf("platform: stopped")
}

func (p *platform) parse(input string) {
	input = strings.TrimSpace(input)
	toks := strings.Split(input, " ")
	if len(toks) <= 0 {
		return
	}

	switch toks[0] {
	case "%":
		if len(toks) < 3 {
			log.Printf("%s: need moar", input)
			return
		}
		modulo, err := strconv.ParseUint(toks[1], 10, 64)
		if err != nil {
			log.Printf("%s: %s", input, err)
			return
		}
		command := strings.Join(toks[2:], " ")
		p.buffer.queue(modulo, command)
		log.Printf("queued %%%d: %s", modulo, command)

	case "add", "a":
		if len(toks) != 3 {
			log.Printf("%s: not right args", input)
			return
		}
		if toks[2] == "" {
			log.Printf("%s: empty name", input)
			return
		}

		var n field.Node
		switch toks[1] {
		case "demo":
			n = newDemoGenerator(toks[2])
		default:
			log.Printf("%s: bad type", input)
			return
		}

		if err := p.add(n); err != nil {
			log.Printf("%s: %v", input, err)
			if s, ok := n.(stopper); ok {
				s.stop()
			}
			return
		}
		log.Printf("%s: OK, added", input)

	case "remove", "rm", "r", "del":
		if len(toks) != 2 {
			log.Printf("%s: not right args", input)
			return
		}
		if err := p.remove(toks[1]); err != nil {
			log.Printf("%s: %s", input, err)
			return
		}
		log.Printf("%s: OK, removed", input)

	case "connect", "conn", "c":
		if len(toks) != 3 {
			log.Printf("%s: not right args", input)
			return
		}
		err := p.connect(toks[1], toks[2])
		log.Printf("%s: %v", input, err)

	case "disconnect", "disconn", "discon", "d":
		if len(toks) != 3 {
			log.Printf("%s: not right args", input)
			return
		}
		err := p.disconnect(toks[1], toks[2])
		log.Printf("%s: %v", input, err)

	case "send", "s":
		if len(toks) < 3 {
			log.Printf("%s: not right args", input)
			return
		}
		n, err := p.field.Get(toks[1])
		if err != nil {
			log.Printf("%s: %s: %s", input, toks[1], err)
			return
		}
		p, ok := n.(parser)
		if !ok {
			log.Printf("%s: it can't parse commands", toks[1])
			return
		}
		command := strings.Join(toks[2:], " ")
		log.Printf("sending to %s: %s", toks[1], command)
		p.parse(command)

	default:
		log.Printf("%s: aroo", input)
	}
}

func (p *platform) add(n field.Node) error {
	return p.field.AddNode(n)
}

func (p *platform) remove(id string) error {
	if id == "mixer" || id == "clock" {
		return errNo
	}
	return p.field.RemoveNode(id)
}

func (p *platform) connect(srcID, dstID string) error {
	if srcID == "mixer" || srcID == "clock" {
		return errNo
	}
	log.Printf("platform: connect: %s to %s", srcID, dstID)
	return p.field.AddEdge(srcID, dstID)
}

func (p *platform) disconnect(srcID, dstID string) error {
	log.Printf("platform: disconnect: %s to %s", srcID, dstID)
	return p.field.RemoveEdge(srcID, dstID)
}
