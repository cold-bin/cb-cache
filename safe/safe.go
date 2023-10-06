package safe

import (
	"errors"
	"sync"
)

var (
	ErrServerBussy = errors.New("server busy, please wait for a while to request")
)

type call struct {
	wg  sync.WaitGroup // avoid reentrancy
	val any
	err error
}

type Group struct {
	mu sync.Mutex // protects mL

	m    map[string]*call
	vnum map[string]int64 // visited number
	maxg int64            // max visited number of concurrent goroutine
}

func (g *Group) SetMaxg(maxg int64) {
	g.maxg = maxg
}

// Once is able to make fn just called once
func (g *Group) Once(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if g.vnum == nil {
		g.vnum = make(map[string]int64)
	}

	if c, ok := g.m[key]; ok {
		if g.maxg <= 0 || g.vnum[key] <= g.maxg { /*no limit or less than or equal to maxg*/
			g.vnum[key]++
			g.mu.Unlock()
			c.wg.Wait()
			return c.val, c.err
		} else { /* the others */
			g.mu.Unlock()
			return nil, ErrServerBussy
		}
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
