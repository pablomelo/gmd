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

func (p *platform) parse(input string) {
	log.Printf("platform: parse: %s", input)
	input = strings.TrimSpace(input)
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
	return p.field.AddEdge(src, dst)
}

func (p *platform) disconnect(src, dst field.Node) error {
	return p.field.RemoveEdge(src, dst)
}
