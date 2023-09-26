package cb_cache

import (
	"fmt"
	"github.com/cold-bin/cb-cache/consistencyhash"
	"github.com/cold-bin/cb-cache/serialization"
	"github.com/cold-bin/cb-cache/serialization/pb"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_cb-cache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peers's base URL, e.g. "https://example.net:8000"
	self     string
	basePath string

	peers       *consistencyhash.Map   // store all of peers
	httpGetters map[string]*httpGetter // key marks different peers, like self

	serializer serialization.Serializer // dependency inject
	mu         sync.Mutex
}

type HPOpt func(*HTTPPool)

func WithSerializer(codec serialization.Serializer) HPOpt {
	return func(pool *HTTPPool) {
		pool.serializer = codec
	}
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string, opts ...HPOpt) *HTTPPool {
	h := &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.serializer == nil {
		h.serializer = &serialization.Protobuf{}
	}

	return h
}

// proxy all http requests
// url path: /:base_path/:group_name/:key
func (c *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, c.basePath) {
		panic("[cb-cache] HTTPPool serving unexpected path: " + r.URL.Path)
	}

	ss := strings.SplitN(r.URL.Path[len(c.basePath):], "/", 2)
	if len(ss) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupname, key := ss[0], ss[1]
	group := GetGroup(groupname)

	bv, err := group.Get(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	// marshal
	bs, err := c.serializer.Marshal(&pb.Response{Value: bv.ByteSlice()})
	if err != nil {
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	if _, err = w.Write(bs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *HTTPPool) Set(peers ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.peers = consistencyhash.NewMap(defaultReplicas)
	c.peers.Set(peers...)
	c.httpGetters = make(map[string]*httpGetter)
	for _, peer := range peers {
		c.httpGetters[peer] = &httpGetter{baseURL: fmt.Sprintf("%s%s", peer, c.basePath)}
	}
}

// PickPeer gets the closest peers, and then call get-function in this peers
func (c *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if peer := c.peers.Get(key); peer != "" && peer != c.self {
		getter, ok := c.httpGetters[peer]
		getter.serializer = c.serializer
		return getter, ok
	}

	return nil, false
}
