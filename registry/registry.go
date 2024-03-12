package registry

import "context"

const (
	// 续约间隔，单位秒
	keepAliveTTL = 10
	// 事件通道缓冲区大小
	eventChanSize = 10
)

type Client interface {
	Registry
	Discovery
}

type Registry interface {
	Register(ctx context.Context, addr string) error
}

type Discovery interface {
	GetAddress(ctx context.Context) ([]string, error)
	Watch(ctx context.Context) <-chan Event
}

// Event 服务变化事件
type Event struct {
	Address string
	Type    EventType
}

type EventType string

const (
	REMOVE = "remove"
	PUT    = "put"
)
