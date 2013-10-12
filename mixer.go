package main

import (
	"log"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/peterbourgon/field"
)

const (
	iChan = 1
	oChan = 1
	sRate = 44100
	bufSz = 1024
)

type identifier interface {
	ID() string
}

type audioSender interface {
	identifier
	audioOut() <-chan []float32
}

type mixer struct {
	stream         *portaudio.Stream
	gain           float32
	audio          chan chan []float32
	connections    chan audioSender
	disconnections chan audioSender
	quit           chan chan struct{}
}

func newMixer() (*mixer, error) {
	m := &mixer{
		stream:         nil,
		gain:           0.1, // TODO make mutable
		audio:          make(chan chan []float32),
		connections:    make(chan audioSender),
		disconnections: make(chan audioSender),
		quit:           make(chan chan struct{}),
	}

	stream, err := portaudio.OpenDefaultStream(iChan, oChan, sRate, bufSz, m)
	if err != nil {
		return nil, err
	}
	if err := stream.Start(); err != nil {
		return nil, err
	}
	m.stream = stream

	go m.loop()
	return m, nil
}

func (m *mixer) stop() {
	q := make(chan struct{})
	m.quit <- q
	<-q
}

func (m *mixer) loop() {
	incoming := map[string]<-chan []float32{}
	for {
		select {
		case c := <-m.audio:
			c <- mux(incoming, m.gain)

		case s := <-m.connections:
			incoming[s.ID()] = s.audioOut()

		case s := <-m.disconnections:
			delete(incoming, s.ID())

		case q := <-m.quit:
			if m.stream == nil {
				panic("mixer: double-stop")
			}
			if err := m.stream.Stop(); err != nil {
				log.Printf("mixer: stop: %s", err)
			}
			if err := m.stream.Close(); err != nil {
				log.Printf("mixer: close: %s", err)
			}
			m.stream = nil
			close(q)
			return
		}
	}
}

func mux(m map[string]<-chan []float32, gain float32) []float32 {
	out := make([]float32, bufSz)
	for _, c := range m {
		buf, ok := <-c
		if !ok {
			buf = make([]float32, bufSz)
		}
		if len(buf) != len(out) {
			panic("bad buf sz") // TODO don't crash
		}
		for i := range out {
			out[i] += gain * buf[i]
		}
	}
	return out
}

func (m *mixer) ProcessAudio(in, out []float32) {
	c := make(chan []float32)
	m.audio <- c
	out = <-c
}

func (m *mixer) ID() string { return "mixer" }

func (m *mixer) Connect(n field.Node) {
	log.Printf("mixer: connect (ignored): %s", n.ID())
}

func (m *mixer) Disconnect(n field.Node) {
	log.Printf("mixer: disconnect (ignored): %s", n.ID())
}

func (m *mixer) Connection(n field.Node) {
	log.Printf("mixer: connection: %s", n.ID())
	if s, ok := n.(audioSender); ok {
		m.connections <- s
	}
}

func (m *mixer) Disconnection(n field.Node) {
	log.Printf("mixer: disconnection: %s", n.ID())
	if s, ok := n.(audioSender); ok {
		m.disconnections <- s
	}
}
