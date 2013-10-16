package main

import (
	"math"
	"testing"
)

func TestMidi2hz(t *testing.T) {
	for midi, expected := range map[int]float32{
		21: 27.50,   // A0
		60: 261.62,  // C4,
		68: 415.30,  // Ab4
		69: 440.00,  // A4 (aww yiss)
		70: 466.16,  // Bb4
		96: 2093.00, // C7
	} {
		if got := midi2hz(midi); !cmpFloat32(got, expected, 0.01) {
			t.Errorf("%d: expected %.4f, got %.4f", midi, expected, got)
		}
	}
}

func cmpFloat32(f, expected, tolerance float32) bool {
	return math.Abs(float64(f-expected)) < float64(tolerance)
}
