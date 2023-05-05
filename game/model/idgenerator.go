package model

import (
	"sync"
)

type IdGenerator struct {
	id ID
	mu sync.Mutex
}

// ID start from 100
var RgIdGenerator = IdGenerator{id: 100}

func (g *IdGenerator) GenId() ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.id += 1
	return g.id
}

func (g *IdGenerator) CurrentId() ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.id
}
