package cb_cache

import (
	"context"
	"fmt"
	"github.com/cold-bin/cb-cache/serialization"
	"log"
	"net/http"
	"testing"
)

var localdata = map[string]string{
	"xjj": "99",
	"lss": "100",
	"acc": "95",
}

func TestNewHTTPPool(t *testing.T) {
	NewGroup("score", 2,
		WithGetter(func(ctx context.Context, k string) (v []byte, err error) {
			log.Println("[local data] search key", k)
			if v, ok := localdata[k]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", k)
		}))
	addr := "localhost:9999"
	peers := NewHTTPPool(addr, 50, WithSerializer(&serialization.Gob{}))
	log.Println("cb-cache is running at", addr)
	http.ListenAndServe(addr, peers)
}
