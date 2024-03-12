package main

import (
	"context"
	"flag"
	"fmt"
	cbcache "github.com/cold-bin/cb-cache"
	"log"
	"net/http"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var addrMap = map[int]string{
	8001: "http://localhost:8001",
	8002: "http://localhost:8002",
	8003: "http://localhost:8003",
}

func main() {
	// 命令行参数解析
	var (
		port int
		api  bool
	)
	// -poot=8001 -api=1
	flag.IntVar(&port, "port", 8001, "Cache server port")
	flag.BoolVar(&api, "api", true, "Start a api server?")
	flag.Parse()

	// 创建本地group
	g := cbcache.NewGroup("scores", 2<<10,
		cbcache.WithHotCacheBytes(2<<8),
		cbcache.WithGetter(func(ctx context.Context, key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			time.Sleep(time.Second)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return []byte{}, fmt.Errorf("%s does not exist", key)
		}))

	// 启动api服务器
	if api {
		// curl http://localhost:9999/api?key=Tom
		go func() {
			apiServerAddr := "localhost:9999"
			http.Handle(cbcache.DefaultBasePath+"api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Query().Get("key")
				view, err := g.Get(r.Context(), key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println(key, view.String())
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write([]byte(fmt.Sprintf("value: %s\n", view.ByteSlice())))
			}))
			log.Printf("api server is running at %s \n", apiServerAddr)
			log.Fatalln(http.ListenAndServe(apiServerAddr, nil))
		}()
	}

	// 启动缓存服务器
	addr := addrMap[port]
	pool := cbcache.NewHTTPPool(addr, 50)
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	err := pool.EtcdRegistry(context.Background(), []string{"localhost:2379"}...)
	if err != nil {
		panic(err)
	}
	pool.Set(addrs...)
	//pool.SetETCDRegistry(context.Background(), "49.233.30.197:2379")
	// 注册给group，这样group就可以从远程服务器获取缓存了
	g.PutPeers(pool)
	log.Fatalln(http.ListenAndServe(addr[7:], pool))
}
