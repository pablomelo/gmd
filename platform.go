package main

import (
	"errors"
	"log"
	"strings"

	"github.com/peterbourgon/field"
)

var (
	errNo = errors.New("no")
)

// platform holds the music objects.
type platform struct {
	mixer *mixer
	field *field.Field
}

func newPlatform() (*platform, error) {
	m, err := newMixer()
	if err != nil {
		return nil, err
	}

	f := field.New()
	f.AddNode(m) // a platform always has a permanent mixer

	return &platform{
		mixer: m,
		field: f,
	}, nil
}

func (p *platform) stop() {
	p.mixer.stop()
	log.Printf("platform: stopped")
}

func (p *platform) parse(input string) {
	input = strings.TrimSpace(input)
	log.Printf("platform: parse: %s", input)
	toks := strings.Split(input, " ")
	if len(toks) <= 0 {
		return
	}

	switch toks[0] {
	case "add", "a":
		if len(toks) != 3 {
			log.Printf("err: %s: not right args", input)
			return
		}
		if toks[2] == "" {
			log.Printf("err: %s: empty name", input)
			return
		}

		var n field.Node
		switch toks[1] {
		case "demo":
			n = newDemoGenerator(toks[2])
		default:
			log.Printf("err: %s: bad type", input)
			return
		}

		if err := p.add(n); err != nil {
			log.Printf("%s: %v", input, err)
			return
		}
		log.Printf("%s: OK, added", input)

	case "remove", "rm", "r", "del":
		if len(toks) != 2 {
			log.Printf("err: %s: not right args", input)
			return
		}
		if err := p.remove(toks[1]); err != nil {
			log.Printf("err: %s: %s", input, err)
			return
		}
		log.Printf("%s: OK, removed", input)

	case "connect", "conn", "c":
		if len(toks) != 3 {
			log.Printf("err: %s: not right args", input)
			return
		}
		err := p.connect(toks[1], toks[2])
		log.Printf("%s: %v", input, err)

	case "disconnect", "disconn", "discon", "d":
		if len(toks) != 3 {
			log.Printf("err: %s: not right args", input)
			return
		}
		err := p.disconnect(toks[1], toks[2])
		log.Printf("%s: %v", input, err)

	default:
		log.Printf("err: %s: aroo", input)
	}
}

func (p *platform) add(n field.Node) error {
	return p.field.AddNode(n)
}

func (p *platform) remove(id string) error {
	if id == "mixer" {
		return errNo
	}
	return p.field.RemoveNode(id)
}

func (p *platform) connect(srcID, dstID string) error {
	if srcID == "mixer" {
		return errNo
	}
	log.Printf("platform: connect: %s to %s", srcID, dstID)
	return p.field.AddEdge(srcID, dstID)
}

func (p *platform) disconnect(srcID, dstID string) error {
	log.Printf("platform: disconnect: %s to %s", srcID, dstID)
	return p.field.RemoveEdge(srcID, dstID)
}
