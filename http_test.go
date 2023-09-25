package cb_cache

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestNewHTTPPool(t *testing.T) {
	localdata := map[string]string{
		"xjj": "99",
		"lss": "100",
		"acc": "95",
	}
	NewGroup("score", 2, WithGetter(func(ctx context.Context, k string) (v []byte, err error) {
		log.Println("[local data] search key", k)
		if v, ok := localdata[k]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", k)
	}))
	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("cb-cache is running at", addr)
	http.ListenAndServe(addr, peers)
}
