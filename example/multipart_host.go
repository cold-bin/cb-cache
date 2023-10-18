package main

import (
	"context"
	"flag"
	"fmt"
	cb_cache "github.com/cold-bin/cb-cache"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

var db = []map[string]string{
	{
		"a": "1",
		"b": "2",
		"c": "3",
	},
	{
		"d": "4",
		"e": "5",
		"f": "6",
	},
	{
		"g": "7",
		"h": "8",
		"k": "9",
	},
}

func createGroup(i int) *cb_cache.Group {
	return cb_cache.NewGroup("scores", 2<<10, cb_cache.WithGetter(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Println("[db] search key", key)
			if i == -1 {
				i = 0
			}
			if v, ok := db[i][key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *cb_cache.Group) {
	peers := cb_cache.NewHTTPPool(addr, 50)
	peers.Set(addrs...)
	gee.PutPeers(peers)
	log.Println("server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *cb_cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			if key == "change_cache" {
				db[0]["a"] = "a"
				db[1]["d"] = "d"
				db[2]["g"] = "g"
				log.Println(db)
				gee.Publish("a")
				gee.Publish("d")
				gee.Publish("g")
				return
			}
			view, err := gee.Get(context.Background(), key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
			go func() {
				gee.Stats.PrintEasyStatisticsInGroup()
			}()
		}))
	log.Println("api server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var apiport int
	flag.IntVar(&port, "port", 8001, "server port")
	flag.IntVar(&apiport, "api", -1, "Start a api server?")
	flag.Parse()

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup(apiport)
	if apiport != -1 {
		// 9000 9001 9002
		go startAPIServer("http://localhost:"+strconv.Itoa(apiport+9000), g)
	}
	startCacheServer(addrMap[port], addrs, g)
}
