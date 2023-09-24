package cb_cache

import (
	"context"
	lruk "github.com/cold-bin/cb-cache/lru-k"
	"reflect"
	"testing"
)

func TestGroup_Get(t *testing.T) {
	type fields struct {
		namespace string
		cache     cacheProxy
		getter    GetterFunc
	}
	type args struct {
		ctx context.Context
		k   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ByteView
		wantErr bool
	}{
		{
			name: "get data from getter",
			fields: fields{
				namespace: "rank",
				cache: cacheProxy{
					cache: lruk.NewCache(2, lruk.WithMaxItem(2), lruk.WithInactiveLimit(1)),
				},
				getter: func(ctx context.Context, k string) (v []byte, err error) {
					if k == "key1" {
						return []byte("good case"), nil
					} else {
						return []byte{}, nil
					}
				},
			},
			args: args{
				ctx: context.Background(),
				k:   "key1",
			},
			want: ByteView{
				b: []byte("good case"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Group{
				namespace: tt.fields.namespace,
				cache:     tt.fields.cache,
				getter:    tt.fields.getter,
			}
			got, err := g.Get(tt.args.ctx, tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("Group.Get(%v, %v) error = %v, wantErr %v", tt.args.ctx, tt.args.k, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Group.Get(%v, %v) = %v, want %v", tt.args.ctx, tt.args.k, got, tt.want)
			}
		})
	}
}
