package cb_cache

import (
	"context"
	"net/http"
	"strings"
)

const defaultBasePath = "/_cb-cache/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self     string
	basePath string
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
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

	bv, err := group.Get(context.Background(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	if _, err = w.Write(bv.ByteSlice()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
