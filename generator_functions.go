package main

import (
	"math"
)

// A generatorFunction should define output for input [0..1]. We scale that to
// the range [0..0.25]. Call that scaled output 'F'. We generate a waveform
// based on phase [0..1] as follows:
//
//    phase < 0.25: output = F
//    phase < 0.50: output = F mirrored horizontally
//    phase < 0.75: output = F mirrored vertically
//    phase < 1.00: output = F mirrored horizontally + vertically
//
// (Thanks to Alexander Surma for the idea on this one.)
type generatorFunction func(float32) float32

func nextGeneratorFunctionValue(f generatorFunction, hz float32, phase *float32) float32 {
	var val, p float32 = 0.0, 0.0
	switch {
	case *phase <= 0.25:
		p = (*phase - 0.00) * 4
		val = f(p) // no mirror
	case *phase <= 0.50:
		p = (*phase - 0.25) * 4
		val = f(1 - p) // horizontal mirror
	case *phase <= 0.75:
		p = (*phase - 0.50) * 4
		val = -f(p) // vertical mirror
	case *phase <= 1.00:
		p = (*phase - 0.75) * 4
		val = -f(1 - p) // horizontal + vertical mirror
	default:
		panic("unreachable")
	}
	*phase += hz / sRate
	if *phase > 1.0 {
		*phase -= 1.0
	}
	return val
}

func saw(x float32) float32 {
	return x
}

func sine(x float32) float32 {
	// want only 1/4 sine over range [0..1], so need x/4
	return float32(math.Sin(2 * math.Pi * float64(x/4)))
}

func square(x float32) float32 {
	if x < 0.5 {
		return 1.0
	}
	return 0.0
}

func nextBuffer(f generatorFunction, hz float32, phase *float32) []float32 {
	buf := make([]float32, bufSz)
	for i := 0; i < bufSz; i++ {
		buf[i] = nextGeneratorFunctionValue(f, hz, phase)
	}
	return buf
}

func nextBufferMany(f generatorFunction, keys keySet) []float32 {
	buf := make([]float32, bufSz)
	for midi, phase := range keys {
		for i := 0; i < bufSz; i++ {
			buf[i] += nextGeneratorFunctionValue(f, midi2hz(midi), &phase)
		}
		keys[midi] = phase
	}
	return buf
}

func midi2hz(midi int) float32 {
	if midi < 20 {
		midi = 20
	}
	if midi > 108 {
		midi = 108
	}
	return float32((440.0 / 32) * math.Pow(2.0, (float64(midi)-9.0)/12.0))
}
