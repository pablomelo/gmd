package main

import (
	"github.com/peterbourgon/field"
)

type platform struct {
	f *field.Field
}

func newPlatform() *platform {
	return &platform{
		f: field.New(),
	}
}
