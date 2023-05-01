package model

import (
	"lintech/rego/iregoter"
	"sync"
)

type IdGenerator struct {
	id iregoter.ID
	mu sync.Mutex
}

var RgIdGenerator = IdGenerator{id: 0}

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
