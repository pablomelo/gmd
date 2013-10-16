package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/peterbourgon/field"
)

type demoGenerator struct {
	id            string
	keyDownEvents chan keyEvent
	keyUpEvents   chan keyEvent
	keysDown      keySet // MIDI keys
	connects      chan connectRequest
	disconnects   chan string
	connected     string
	output        chan []float32
	quit          chan chan struct{}
}

func newDemoGenerator(id string) *demoGenerator {
	g := &demoGenerator{
		id:            id,
		keysDown:      keySet{},
		keyDownEvents: make(chan keyEvent),
		keyUpEvents:   make(chan keyEvent),
		connects:      make(chan connectRequest),
		disconnects:   make(chan string),
		connected:     "",
		output:        nil,
		quit:          make(chan chan struct{}),
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
	log.Printf("%s: started", g.ID())
	defer log.Printf("%s: done", g.ID())

	for {
		select {
		case g.output <- nextBufferMany(sine, g.keysDown):
			//log.Printf("%s ♪", g.ID())
			break

		case k := <-g.keyDownEvents:
			log.Printf("%s: press %d", g.ID(), k.midi)
			g.keysDown.add(k.midi) // TODO velocity
			log.Printf("%s: keys down %v", g.ID(), g.keysDown)

		case k := <-g.keyUpEvents:
			log.Printf("%s: lift %d", g.ID(), k.midi)
			if k.midi == 0 {
				g.keysDown = keySet{} // reset
			} else {
				g.keysDown.remove(k.midi)
			}
			log.Printf("%s: keys down %v", g.ID(), g.keysDown)

		case r := <-g.connects:
			if g.output != nil {
				r.e <- fmt.Errorf("%s already connected to %s", g.ID(), g.connected)
				continue
			}
			g.connected = r.r.ID()
			g.output = make(chan []float32, 1)
			r.r.receive(g.output)
			r.e <- nil
			log.Printf("%s → %s", g.ID(), r.r.ID())

		case id := <-g.disconnects:
			if g.output == nil {
				log.Printf("%s: disconnect, but not connected", g.ID())
				continue
			}
			if g.connected != id {
				log.Printf("%s: connected to %s, not %s (bug in field)", g.ID(), g.connected, id)
				continue
			}
			g.connected = ""
			close(g.output)
			g.output = nil
			log.Printf("%s ✕ %s", g.ID(), id)

		case q := <-g.quit:
			close(q)
			return
		}
	}
}

func (g *demoGenerator) parse(input string) {
	input = strings.TrimSpace(strings.ToLower(input))
	toks := strings.Split(input, " ")
	if len(toks) <= 0 {
		log.Printf("clock: parse empty")
		return
	}

	switch toks[0] {
	case "keydown", "kd", "down", "d":
		if len(toks) < 2 {
			log.Printf("%s: %s: not enough", g.ID(), input)
			return
		}

		midi, err := strconv.ParseInt(toks[1], 10, 32)
		if err != nil {
			log.Printf("%s: %s: %s", g.ID(), input, err)
			return
		}

		velocity := 1.0
		if len(toks) >= 3 {
			velocity, err = strconv.ParseFloat(toks[2], 32)
			if err != nil {
				log.Printf("%s: %s: %s", g.ID(), input, err)
				return
			}
		}

		g.keyDownEvents <- keyEvent{int(midi), float32(velocity)}

	case "keyup", "ku", "up", "u":
		if len(toks) < 2 {
			log.Printf("%s: %s: not enough", g.ID(), input)
			return
		}

		midi, err := strconv.ParseInt(toks[1], 10, 32)
		if err != nil {
			log.Printf("%s: %s: %s", g.ID(), input, err)
			return
		}

		velocity := 1.0
		if len(toks) >= 3 {
			velocity, err = strconv.ParseFloat(toks[2], 32)
			if err != nil {
				log.Printf("%s: %s: %s", g.ID(), input, err)
				return
			}
		}

		g.keyUpEvents <- keyEvent{int(midi), float32(velocity)}

	case "reset", "release":
		g.keyUpEvents <- keyEvent{}
	}
}

func (g *demoGenerator) ID() string { return g.id }

func (g *demoGenerator) Connect(n field.Node) error {
	r, ok := n.(audioReceiver)
	if !ok {
		return fmt.Errorf("%s not audioReceiver", n.ID())
	}
	req := connectRequest{r, make(chan error)}
	g.connects <- req
	return <-req.e
}

func (g *demoGenerator) Disconnect(n field.Node) {
	if _, ok := n.(audioReceiver); !ok {
		log.Printf("%s not audioReceiver", n.ID())
		return
	}
	g.disconnects <- n.ID()
}

func (g *demoGenerator) Connection(n field.Node) error {
	log.Printf("%s: Connection(%s): no", g.ID(), n.ID())
	return errNo
}

func (g *demoGenerator) Disconnection(n field.Node) {
	log.Printf("%s: Disonnection(%s): ignored", g.ID(), n.ID())
}

type connectRequest struct {
	r audioReceiver
	e chan error
}

type keyEvent struct {
	midi     int     // key
	velocity float32 // 0..1
}

type keySet map[int]float32 // MIDI key: phase

func (s keySet) add(i int)    { s[i] = 0.0 }
func (s keySet) remove(i int) { delete(s, i) }
