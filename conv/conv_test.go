package conv

import (
	"reflect"
	"testing"
)

func TestQuickS2B(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "case1",
			args: args{str: "test"},
			want: []byte("test"),
		},
		{
			name: "case2",
			args: args{str: "ad&*(2_?>P{>?A"},
			want: []byte("ad&*(2_?>P{>?A"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuickS2B(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QuickS2B(%v) = %v, want %v", tt.args.str, got, tt.want)
			}
		})
	}
}

func TestQuickB2S(t *testing.T) {
	type args struct {
		bs []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{bs: []byte("test")},
			want: "test",
		},
		{
			name: "case2",
			args: args{bs: []byte("ad&*(2_?>P{>?A")},
			want: "ad&*(2_?>P{>?A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuickB2S(tt.args.bs); got != tt.want {
				t.Errorf("QuickB2S(%v) = %v, want %v", tt.args.bs, got, tt.want)
			}
		})
	}
}
