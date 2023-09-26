package cb_cache

import (
	"context"
	"fmt"
	"github.com/cold-bin/cb-cache/serialization"
	"github.com/cold-bin/cb-cache/serialization/pb"
	"io"
	"net/http"
	"net/url"
)

// PeerPicker is the interface that must be implemented to locate
// the peers that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peers.
type PeerGetter interface {
	Get(ctx context.Context, req *pb.Request) (_r *pb.Response, _err error)
}

type httpGetter struct {
	baseURL    string
	serializer serialization.Serializer
}

func (h *httpGetter) Get(ctx context.Context, req *pb.Request) (_r *pb.Response, _err error) {
	res, err := http.Get(fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(req.GetGroup()),
		url.QueryEscape(req.GetKey()),
	))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	// unmarshal
	_r = &pb.Response{}
	if err = h.serializer.Unmarshal(bytes, _r); err != nil {
		return nil, err
	}

	return _r, nil
}
