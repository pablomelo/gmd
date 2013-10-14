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
	regis map[string]field.Node
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
		regis: map[string]field.Node{},
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
		p.regis[n.ID()] = n
		log.Printf("%s: OK, added", input)

	case "remove", "rm", "r", "del":
		if len(toks) != 2 {
			log.Printf("err: %s: not right args", input)
			return
		}
		n, ok := p.regis[toks[1]]
		if !ok {
			log.Printf("err: %s: lost it", input)
			return
		}
		if err := p.remove(n); err != nil {
			log.Printf("err: %s: %s", input, err)
			return
		}
		delete(p.regis, n.ID())
		log.Printf("%s: OK, removed", input)

	case "connect", "conn", "c":
		if len(toks) != 3 {
			log.Printf("err: %s: not right args", input)
			return
		}
		src, ok := p.regis[toks[1]]
		if !ok {
			log.Printf("err: %s: no src", input)
			return
		}
		dst, ok := p.regis[toks[2]]
		if !ok {
			log.Printf("err: %s: no dst", input)
		}
		err := p.connect(src, dst)
		log.Printf("%s: %v", input, err)

	case "disconnect", "disconn", "discon", "d":
		if len(toks) != 3 {
			log.Printf("err: %s: not right args", input)
			return
		}
		src, ok := p.regis[toks[1]]
		if !ok {
			log.Printf("err: %s: no src", input)
			return
		}
		dst, ok := p.regis[toks[2]]
		if !ok {
			log.Printf("err: %s: no dst", input)
		}
		err := p.disconnect(src, dst)
		log.Printf("%s: %v", input, err)

	default:
		log.Printf("err: %s: wha", input)
	}
}

func (p *platform) add(n field.Node) error {
	return p.field.AddNode(n)
}

func (p *platform) remove(n field.Node) error {
	if n.ID() == "mixer" {
		return errNo
	}
	return p.field.RemoveNode(n)
}

func (p *platform) connect(src, dst field.Node) error {
	if src.ID() == "mixer" {
		return errNo
	}
	log.Printf("platform: connect: %s to %s", src.ID(), dst.ID())
	return p.field.AddEdge(src, dst)
}

func (p *platform) disconnect(src, dst field.Node) error {
	log.Printf("platform: disconnect: %s to %s", src.ID(), dst.ID())
	return p.field.RemoveEdge(src, dst)
}
