package lru_k

import (
	"container/list"
)

type Cache interface {
	Get(k string) (v any, ok bool)
	Set(k string, v any)
	Len() int
	Remove(k string)
	Clear()
}

// cache is an LRU-2 cache. It is not safe for concurrent access
type cache struct {
	k int
	// item's max number of cache. if zero, no limit for activeList and inactiveList
	maxItem int
	// the max size of inactiveList
	inactiveLimit int

	// InactiveList just store keys and values that is
	// visited less than 2 times and also use lru. these keys may be inactive,
	// and may make the hit rate of cache reduce. so we separate it from ActiveList
	inactiveList *list.List
	inactiveMap  map[string]*list.Element

	// ActiveList just store keys (values is stored in items) that is visited more than or equal to
	// 2 times and also use lru
	activeList *list.List
	activeMap  map[string]*list.Element // real data

	// callback function for the key before being eliminated
	onEliminate func(k string, v any)
}

// entry is used in cache.activeList
type entry struct {
	k   string
	v   any
	cnt uint64 // visited times
}

func NewCache(k int, opts ...Option) Cache {
	c := &cache{
		k:            k,
		inactiveList: list.New(),
		inactiveMap:  make(map[string]*list.Element),
		activeList:   list.New(),
		activeMap:    make(map[string]*list.Element),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.maxItem != 0 && c.maxItem <= c.inactiveLimit {
		panic("[cb-cache]: maxItem is more than inactiveLimit")
	}

	if k < 2 {
		panic("[cb-cache]: k is more than 2")
	}

	return c
}

func (c *cache) Get(k string) (v any, ok bool) {
	if c.isNil() {
		return
	}

	if e, ok_ := c.inactiveMap[k]; ok_ { /*first in inactive list*/
		entry := e.Value.(*entry)
		entry.cnt++
		if entry.cnt >= 2 { /*move to real cache*/
			c.moveToRealCache(entry, e)
		} else { /*move to frontend locally*/
			c.inactiveList.MoveToFront(e)
		}
		v, ok = entry.v, true
		return
	}

	if e, ok_ := c.activeMap[k]; ok_ { /*maybe in active list*/
		c.activeList.MoveToFront(e)
		v, ok = e.Value.(*entry).v, true
	} else { /*not in cache*/
		v, ok = nil, false
	}

	return
}

func (c *cache) moveToRealCache(entry_ *entry, e *list.Element) {
	c.activeList.PushFront(entry_)
	c.activeMap[entry_.k] = e
	c.inactiveList.Remove(e)
	delete(c.inactiveMap, entry_.k)

	if c.maxItem != 0 && c.activeList.Len() > c.maxItem-c.inactiveLimit { /*eliminate*/
		ele := c.activeList.Remove(c.activeList.Back()).(*entry)
		if c.onEliminate != nil {
			c.onEliminate(entry_.k, entry_.v)
		}
		delete(c.activeMap, ele.k)
	}
}

func (c *cache) Set(k string, v any) {
	if c.isNil() {
		c.fill()
	}

	if e, ok_ := c.inactiveMap[k]; ok_ { /*if k is hit in inactive list*/
		entry := e.Value.(*entry)
		entry.v = v
		entry.cnt++
		if entry.cnt >= 2 { /*move to real cache*/
			c.moveToRealCache(entry, e)
		} else { /*move to frontend locally*/
			c.inactiveList.MoveToFront(e)
		}
		return
	}

	if e, ok_ := c.activeMap[k]; ok_ { /*maybe hit in active list*/
		e.Value.(*entry).v = v
		c.activeList.MoveToFront(e)
	} else { /*not in cache,place the item inactive list*/
		e := c.inactiveList.PushFront(&entry{k: k, v: v})
		c.inactiveMap[k] = e
		if c.maxItem != 0 && c.inactiveLimit < c.inactiveList.Len() {
			ele := c.inactiveList.Remove(c.inactiveList.Back()).(*entry)
			if c.onEliminate != nil {
				c.onEliminate(ele.k, ele.v)
			}
			delete(c.inactiveMap, ele.k)
		}
	}
}

func (c *cache) Remove(k string) {
	if c.isNil() {
		return
	}

	if e, ok_ := c.inactiveMap[k]; ok_ { /*if k is hit in inactive list*/
		delete(c.inactiveMap, k)
		c.inactiveList.Remove(e)
		return
	}

	if e, ok_ := c.activeMap[k]; ok_ { /*maybe hit in active list*/
		delete(c.activeMap, k)
		c.activeList.Remove(e)
	}
}

func (c *cache) Clear() {
	if c.onEliminate != nil {
		for _, e := range c.activeMap {
			kv := e.Value.(*entry)
			c.onEliminate(kv.k, kv.v)
		}
	}
	c.activeMap = nil
	c.inactiveMap = nil
	c.activeList = nil
}

func (c *cache) Len() int {
	if c.isNil() {
		return 0
	}
	return len(c.inactiveMap) + len(c.activeMap)
}

func (c *cache) isNil() bool {
	return c.inactiveMap == nil || c.activeMap == nil
}

func (c *cache) fill() {
	c.inactiveList = list.New()
	c.inactiveMap = make(map[string]*list.Element)
	c.activeList = list.New()
	c.activeMap = make(map[string]*list.Element)
}
