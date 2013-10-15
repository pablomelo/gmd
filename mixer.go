package main

import (
	"log"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/peterbourgon/field"
)

const (
	iChan = 1     // portaudio input channels
	oChan = 1     // portaudio output channels
	sRate = 44100 // audio sample rate
	bufSz = 1024  // size of buffer in each portaudio ProcessAudio call
)

type identifier interface {
	ID() string
}

type audioReceiver interface {
	identifier
	receive(audioOut <-chan []float32)
}

type mixer struct {
	stream   *portaudio.Stream
	gain     float32
	incoming chan (<-chan []float32) // connections from upstream
	audio    chan chan []float32
	quit     chan chan struct{}
}

func newMixer() (*mixer, error) {
	m := &mixer{
		stream:   nil,
		gain:     0.1, // TODO make mutable
		incoming: make(chan (<-chan []float32)),
		audio:    make(chan chan []float32),
		quit:     make(chan chan struct{}),
	}

	stream, err := portaudio.OpenDefaultStream(iChan, oChan, sRate, bufSz, m)
	if err != nil {
		return nil, err
	}
	if err := stream.Start(); err != nil {
		return nil, err
	}
	log.Printf("mixer: stream started")
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
	incoming := []<-chan []float32{}
	for {
		select {
		case c := <-m.incoming:
			incoming = append(incoming, c)

		case c := <-m.audio:
			var buf []float32
			incoming, buf = mux(incoming, m.gain)
			c <- buf

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

func (m *mixer) ProcessAudio(in, out []float32) {
	c := make(chan []float32)
	m.audio <- c
	buf := <-c
	for i := 0; i < bufSz; i++ {
		out[i] = buf[i]
	}
}

// receive implements the audioReceiver interface.
func (m *mixer) receive(audioOut <-chan []float32) {
	m.incoming <- audioOut
}

var zeroBuf = make([]float32, 0.0)

func mux(incoming []<-chan []float32, gain float32) ([]<-chan []float32, []float32) {
	survivors := make([]<-chan []float32, 0, len(incoming))
	out := make([]float32, bufSz)
	for _, c := range incoming {
		var buf []float32
		ok, timeout := true, false
		select {
		case buf, ok = <-c:
			break
		default:
			timeout = true
		}
		if !ok {
			log.Printf("mixer: mux: ch %x closed, culling", c)
			continue
		}
		if timeout {
			log.Printf("mixer: mux: ch %x timeout, skipping", c)
			buf = zeroBuf
		}
		if len(buf) != len(out) {
			panic("bad buf sz") // TODO don't crash
		}

		survivors = append(survivors, c)
		for i := range out {
			out[i] += gain * buf[i]
		}
	}
	//log.Printf("mux: %d (first=%.2f) => %d", len(incoming), out[0], len(survivors))
	return survivors, out
}

func (m *mixer) ID() string { return "mixer" }

func (m *mixer) Connect(n field.Node) error {
	log.Printf("mixer: Connect (rejected): %s", n.ID())
	return errNo
}

func (m *mixer) Connection(n field.Node) error {
	log.Printf("mixer: Connection: %s (ignored)", n.ID())
	return nil
}

func (m *mixer) Disconnect(n field.Node) {
	log.Printf("mixer: Disconnect (ignored): %s", n.ID())
}

func (m *mixer) Disconnection(n field.Node) {
	log.Printf("mixer: Disconnection: %s (ignored)", n.ID())
}
