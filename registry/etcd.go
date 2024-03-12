package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	mu sync.Mutex

	kv        etcdv3.KV
	watcher   etcdv3.Watcher
	grantid   etcdv3.LeaseID
	endpoints []string
	prefix    string
}

func New(ctx context.Context, prefix string, endpoints []string) (Client, error) {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	lease := etcdv3.NewLease(client)
	grant, err := lease.Grant(ctx, keepAliveTTL)
	if err != nil {
		return nil, err
	}

	ch, err := lease.KeepAlive(ctx, grant.ID)
	if err != nil {
		return nil, err
	}

	go func() {
		for range ch {
		}
	}()

	return &etcd{
		kv:        etcdv3.NewKV(client),
		watcher:   etcdv3.NewWatcher(client),
		grantid:   grant.ID,
		endpoints: endpoints,
		prefix:    prefix,
	}, nil
}

// Register a node and start goroutine to keep lease
func (r *etcd) Register(ctx context.Context, addr string) error {
	key := fmt.Sprintf("%s%s", r.prefix, addr)
	_, err := r.kv.Put(ctx, key, addr, etcdv3.WithLease(r.grantid))
	return err
}

func (r *etcd) Deregister(ctx context.Context, addr string) error {
	key := fmt.Sprintf("%s%s", r.prefix, addr)
	_, err := r.kv.Delete(ctx, key, etcdv3.WithLease(r.grantid))
	return err
}

// GetAddress get all active node's address
func (r *etcd) GetAddress(ctx context.Context) ([]string, error) {
	resp, err := r.kv.Get(ctx, r.prefix, etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addrs := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		addrs[i] = string(kv.Value)
	}
	return addrs, nil
}

// Watch etcd event
// case 1: maybe some nodes are broken, then other clients will delete it
// case 2: some nodes are added, then other clients will add it
func (r *etcd) Watch(ctx context.Context) <-chan Event {
	watchChan := r.watcher.Watch(ctx, r.prefix, etcdv3.WithPrefix())
	ch := make(chan Event, eventChanSize)
	go func() {
		for watchRsp := range watchChan {
			for _, event := range watchRsp.Events {
				switch event.Type {
				case mvccpb.PUT:
					ch <- Event{Address: string(event.Kv.Value), Type: PUT}
				case mvccpb.DELETE:
					ch <- Event{Address: string(event.Kv.Key[len(r.prefix):]), Type: REMOVE}
				}
			}
		}
		close(ch)
	}()
	return ch
}
