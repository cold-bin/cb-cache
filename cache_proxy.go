package cb_cache

import (
	"context"
	"errors"
	"fmt"
	lruk "github.com/cold-bin/cb-cache/lru-k"
	"github.com/cold-bin/cb-cache/safe"
	"github.com/cold-bin/cb-cache/serialization/pb"
	"math/rand"
	"sync"
	"sync/atomic"
)

var (
	ErrKeyEmpty = errors.New("[cb-cache] k is empty")
)

// Stats should be operated atomically
type Stats struct {
	Gets             uint64 // any Get request, including from peers
	CacheHits        uint64 // local mainCache hit
	PeerLoads        uint64 // either remote load or remote mainCache hit (not an error)
	PeerErrors       uint64
	GetterFuncFrom   uint64 // total good getter loads
	GetterFuncFailed uint64 // total bad getter loads
	ServerRequests   uint64 // gets that came over the network from peers

	rlock sync.RWMutex
}

type StatsCopy struct {
	Gets             uint64 // any Get request, including from peers
	CacheHits        uint64 // local mainCache hit
	PeerLoads        uint64 // either remote load or remote mainCache hit (not an error)
	PeerErrors       uint64
	GetterFuncFrom   uint64 // total good getter loads
	GetterFuncFailed uint64 // total bad getter loads
	ServerRequests   uint64 // gets that came over the network from peers
}

// PrintEasyStatisticsInGroup
//
//	mainCache hit rate, peer load rate, rate of data from network etc...
func (s *Stats) PrintEasyStatisticsInGroup() {
	s.rlock.RLock()
	state := &Stats{
		Gets:             s.Gets,
		CacheHits:        s.CacheHits,
		PeerLoads:        s.PeerLoads,
		PeerErrors:       s.PeerErrors,
		GetterFuncFrom:   s.GetterFuncFrom,
		GetterFuncFailed: s.GetterFuncFailed,
		ServerRequests:   s.ServerRequests,
	}
	s.rlock.RUnlock()
	if state.Gets == 0 {
		fmt.Println("gets is zero")
		return
	}
	fmt.Println(fmt.Sprintf(
		" cache rate: %.2f%% \n peer load rate: %.2f%% \n data from network: %.2f%% \n",
		float64(state.CacheHits)/float64(state.Gets)*100,
		float64(state.PeerLoads)/float64(state.Gets)*100,
		float64(state.ServerRequests)/float64(state.Gets)*100,
	))
}

// GetterFunc implement Getter in order to Pass in the get function directly
type GetterFunc func(ctx context.Context, k string) (v []byte, err error)

// Group is divided by namespace,and mainCache is different every Group
type Group struct {
	namespace string
	nBytes    int64 // max bytes

	mainCache cacheProxy // cached hot keys from local machine
	hotCache  cacheProxy // cached hot keys from remote machine to avoid to request the same keys again

	getter GetterFunc  // if got not in mainCache, use getter. this maybe prevent mainCache breakdown
	peers  PeerPicker  // as a remote get-function from the other peers.
	loader *safe.Group // make sure that every key is visited only once at the same time

	Stats Stats // statics data of every group
}

type GOption func(*Group)

// WithRetirementPolicy Provides a self-implementing mainCache retirement strategy
func WithRetirementPolicy(cache1, cache2 lruk.Cache) GOption {
	return func(g *Group) {
		g.mainCache.cache = cache1
		g.hotCache.cache = cache2
	}
}

// WithConcurrentMaxGNum prevents goroutines blocking issues caused by a large number of client accesses when the mainCache is broken down.
// if maxg <=0, mean that there is no limit to visit
func WithConcurrentMaxGNum(maxg int64) GOption {
	return func(g *Group) {
		g.loader.SetMaxg(maxg)
	}
}

func WithGetter(getter GetterFunc) GOption {
	return func(g *Group) {
		g.getter = getter
	}
}

// NewGroup
// nBytes: max bytes of group
// maxitems: config of lru-k
// nBytes will
func NewGroup(namespace string, nBytes int64, maxitems int, opts ...GOption) *Group {
	gmu.Lock()
	defer gmu.Unlock()

	if _, dup := groups[namespace]; dup {
		panic("duplicate registration of group " + namespace)
	}

	g := &Group{
		namespace: namespace,
		nBytes:    nBytes,
		mainCache: cacheProxy{cache: lruk.NewCache(2)},
		hotCache:  cacheProxy{cache: lruk.NewCache(2)},
		getter: func(ctx context.Context, k string) (v []byte, err error) {
			return []byte{}, nil
		}, /*default getter*/
		loader: &safe.Group{},
	}

	for _, opt := range opts {
		opt(g)
	}
	groups[namespace] = g

	return g
}

var (
	gmu    sync.RWMutex // used in lock groups
	groups = make(map[string]*Group)
)

const DefaultQueueCap = 50

// GetGroup get group in read-lock
func GetGroup(name string) *Group {
	gmu.RLock()
	g := groups[name]
	gmu.RUnlock()
	return g
}

func (g *Group) Name() string {
	return g.namespace
}

func (g *Group) PutPeers(pp PeerPicker) {
	if g.peers != nil {
		return
	}

	g.peers = pp
}

func (g *Group) Get(ctx context.Context, k string) (ByteView, error) {
	if k == "" {
		return ByteView{}, ErrKeyEmpty
	}
	atomic.AddUint64(&g.Stats.Gets, 1)

	// first,try to get v from main cache and hot cache
	value, cacheHit := g.localCache(k)
	if cacheHit {
		atomic.AddUint64(&g.Stats.CacheHits, 1)
		return value, nil
	}

	// second,try to get v from the remote peers
	fn := func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(k); ok {
				var (
					err error
					req = &pb.Request{Group: g.namespace, Key: k}
					res = &pb.Response{}
				)
				if res, err = peer.Get(ctx, req); err == nil {
					atomic.AddUint64(&g.Stats.PeerLoads, 1)
					// should store the remote data from other peers in hotCache,
					// but we can't store every key from remote. only P = 1/10
					v := ByteView{b: res.Value}
					if rand.Intn(10) == 0 {
						g.populateCache(k, v, &g.hotCache)
					}
					return v, nil
				}
				atomic.AddUint64(&g.Stats.PeerErrors, 1)
			}
		}

		// not got in local cache, then got in g.Getter and store in mainCache locally
		bs, err := g.getter(ctx, k)
		if err != nil {
			atomic.AddUint64(&g.Stats.GetterFuncFailed, 1)
			return ByteView{}, err
		}

		atomic.AddUint64(&g.Stats.GetterFuncFrom, 1)
		bw := ByteView{b: cloneBytes(bs)}

		// populate local cache
		g.populateCache(k, value, &g.mainCache)

		return bw, nil
	}

	v, err := g.loader.Once(k, fn)
	return v.(ByteView), err
}

func (g *Group) localCache(k string) (value ByteView, ok bool) {
	if g.nBytes <= 0 {
		return
	}

	value, ok = g.mainCache.get(k)
	if ok {
		return
	}

	value, ok = g.hotCache.get(k)
	return
}

func (g *Group) populateCache(key string, value ByteView, cache *cacheProxy) {
	if g.nBytes <= 0 {
		return
	}

	cache.set(key, value)

	// Evict items from cache(s) if necessary.
	for {
		mainBytes := g.mainCache.nBytes()
		hotBytes := g.hotCache.nBytes()
		if mainBytes+hotBytes <= g.nBytes /*no eviction*/ {
			return
		}

		// need eviction
		victim := &g.mainCache
		if hotBytes > mainBytes/8 {
			/*if hotCache is 1/8 of mainCache, will evict old key on hotCache*/
			victim = &g.hotCache
		}
		victim.removeOldest()
	}
}

type CacheType uint8

const (
	MainCache = iota + 1
	HotCache
)

func (g *Group) CacheStates(cacheType CacheType) CacheStats {
	switch cacheType {
	case MainCache:
		return g.mainCache.stats()
	case HotCache:
		return g.hotCache.stats()
	default:
		return CacheStats{}
	}
}

// cacheProxy proxy of lru_k.cache to add other functions,
// such as providing concurrent access, statistics of hit rate etc.
type cacheProxy struct {
	cache lruk.Cache
	mu    sync.RWMutex

	nbytes     int64 // all keys and bytes
	nhit, nget int64
	nevict     int64 // number of evictions
}

func (c *cacheProxy) stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Bytes:     c.nbytes,
		Items:     int64(c.cache.Len()),
		Gets:      c.nget,
		Hits:      c.nhit,
		Evictions: c.nevict,
	}
}

func (c *cacheProxy) nBytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nbytes
}

func (c *cacheProxy) nItems() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int64(c.cache.Len())
}

func (c *cacheProxy) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = lruk.NewCache(2, lruk.WithOnEliminate(func(k string, v any) {
			c.nbytes -= int64(len(k)) + int64(v.(ByteView).Len())
			c.nevict++
		}))
	}
	c.cache.Set(key, value)
	c.nbytes += int64(len(key)) + int64(value.Len())
}

func (c *cacheProxy) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nget++
	if c.cache == nil {
		return
	}

	if v, ok := c.cache.Get(key); ok { /*hit*/
		c.nhit++
		return v.(ByteView), ok
	}

	return
}

func (c *cacheProxy) removeOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nevict--
	c.cache.RemoveOldest()
}

// CacheStats is state of current mainCache
type CacheStats struct {
	Bytes     int64
	Items     int64
	Gets      int64
	Hits      int64
	Evictions int64
}
