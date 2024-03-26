package safe

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup // avoid reentrancy
	val any
	err error
}

type Group struct {
	mu sync.Mutex // protects mL
	m  map[string]*call
}

// Once is able to make fn just called once
func (g *Group) Once(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
