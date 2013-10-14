package main

import (
	"fmt"
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
	subscribeAudio(string) <-chan []float32
	unsubscribeAudio(string)
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
	active := map[string]audioSender{}
	for {
		select {
		case c := <-m.audio:
			//log.Printf("mixer: audio")
			c <- mux(incoming, m.gain)

		case s := <-m.connections:
			log.Printf("mixer: connections: %s", s.ID())
			active[s.ID()] = s
			incoming[s.ID()] = s.subscribeAudio(m.ID())

		case s := <-m.disconnections:
			log.Printf("mixer: disconnections: %s", s.ID())
			delete(incoming, s.ID())
			if _, ok := active[s.ID()]; !ok {
				panic(fmt.Sprintf("mixer disconnection from inactive sender '%s'", s.ID()))
			}
			active[s.ID()].unsubscribeAudio(m.ID())
			delete(active, s.ID())

		case q := <-m.quit:
			log.Printf("mixer: quit")
			defer log.Printf("mixer: done")
			if m.stream == nil {
				panic("mixer: double-stop")
			}

			// log.Printf("mixer: stream stopping...")
			// if err := m.stream.Stop(); err != nil {
			// 	log.Printf("mixer: stream stop: %s", err)
			// }
			// log.Printf("mixer: stream stopped")

			log.Printf("mixer: stream closing...")
			if err := m.stream.Close(); err != nil {
				log.Printf("mixer: stream close: %s", err)
			}
			log.Printf("mixer: stream closed")

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
