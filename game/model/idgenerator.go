package model

import (
	"lintech/rego/iregoter"
	"sync"
)

type IdGenerator struct {
	id iregoter.ID
	mu sync.Mutex
}

// ID start from 100
var RgIdGenerator = IdGenerator{id: 100}

func (g *IdGenerator) GenId() iregoter.ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.id += 1
	return g.id
}

func (g *IdGenerator) CurrentId() iregoter.ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.id
}
