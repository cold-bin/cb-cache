package cb_cache

import (
	"context"
	"errors"
	lruk "github.com/cold-bin/cb-cache/lru-k"
	"github.com/cold-bin/cb-cache/safe"
	"github.com/cold-bin/cb-cache/serialization/pb"
	"log"
	"sync"
)

var (
	ErrKeyEmpty = errors.New("[cb-cache] k is empty")
)

// GetterFunc implement Getter in order to Pass in the get function directly
type GetterFunc func(ctx context.Context, k string) (v []byte, err error)

// Group is divided by namespace,and cache is different every Group
type Group struct {
	namespace string
	cache     cacheProxy
	getter    GetterFunc // if got not in cache, use getter. this maybe prevent cache breakdown

	peers  PeerPicker  // as a remote get-function from the other peers.
	loader *safe.Group // make sure that every key is visited only once at the same time
}

type GOption func(*Group)

// WithRetirementPolicy Provides a self-implementing cache retirement strategy
func WithRetirementPolicy(cache lruk.Cache) GOption {
	return func(g *Group) {
		g.cache.cache = cache
	}
}

func WithGetter(getter GetterFunc) GOption {
	return func(g *Group) {
		g.getter = getter
	}
}

func NewGroup(namespace string, maxitems int, opts ...GOption) *Group {
	gmu.Lock()
	defer gmu.Unlock()

	g := &Group{
		namespace: namespace,
		getter: func(ctx context.Context, k string) (v []byte, err error) {
			return []byte{}, nil
		}, /*default getter*/
		cache:  cacheProxy{cache: lruk.NewCache(2, lruk.WithMaxItem(maxitems), lruk.WithInactiveLimit(maxitems/2))}, /*default cache*/
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

// GetGroup get group in read-lock
func GetGroup(name string) *Group {
	gmu.RLock()
	g := groups[name]
	gmu.RUnlock()
	return g
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

	// first step, try to get v from the remote peers
	fn := func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(k); ok {
				var (
					err error
					req = &pb.Request{Group: g.namespace, Key: k}
					res = &pb.Response{}
				)
				if res, err = peer.Get(ctx, req); err == nil {
					return ByteView{b: res.Value}, nil
				}
				log.Println("[cb-cache] failed to get from peer:", err)
			}
		}

		// second step, got in cache locally
		if bw, ok := g.cache.get(k); ok {
			return bw, nil
		}

		return ByteView{}, nil
	}
	g.loader.Once(k, fn)

	// not got in cache, then got in g.Getter and store in cache locally
	bs, err := g.getter(ctx, k)
	if err != nil {
		return ByteView{}, err
	}
	bw := ByteView{b: cloneBytes(bs)}
	g.cache.set(k, bw)

	return bw, nil
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

// CacheStats is state of current cache
type CacheStats struct {
	Bytes     int64
	Items     int64
	Gets      int64
	Hits      int64
	Evictions int64
}
