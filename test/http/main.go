package main

import (
	"context"
	"fmt"
	cbcache "github.com/cold-bin/cb-cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	cbcache.NewGroup("scores", 2<<10, cbcache.WithGetter(func(ctx context.Context, key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return []byte{}, fmt.Errorf("%s does not exist", key)
	}))
	addr := "localhost:9999"
	peers := cbcache.NewHTTPPool(addr, 50)
	log.Printf("[cb-cache service] is running at %s \n", addr)
	log.Fatalln(http.ListenAndServe(addr, peers))
}
