package consistencyhash

import (
	"github.com/cold-bin/cb-cache/conv"
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Map struct {
	hash    Hash           // dependency inject
	replica int            // number of per real node's virtual node
	keys    []int          // sorted hash ring
	hashMap map[int]string // a map from virtual node to real node
}

type MOpt func(*Map)

func WithHash(hash Hash) MOpt {
	return func(m *Map) {
		m.hash = hash
	}
}

func NewMap(replica int, opts ...MOpt) *Map {
	m := &Map{
		replica: replica,
		hashMap: make(map[int]string),
	}
	for _, opt := range opts {
		opt(m)
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	if replica < 0 {
		panic("illegal replica")
	}

	return m
}

// Set adds some keys into hash
func (m *Map) Set(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replica; i++ {
			// just 64-bit system
			hash := int(m.hash(conv.QuickS2B(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to provided key
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash(conv.QuickS2B(key)))
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
	// if idx==len(m.Keys), return the first key in the cycle
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
