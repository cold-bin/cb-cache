package lru_k

import (
	"fmt"
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	var getTests = []struct {
		name       string
		keyToAdd   string
		keyToGet   string
		expectedOk bool
	}{
		{"hit", "myKey", "myKey", true},
		{"miss", "myKey", "nonsense", false},
	}
	for _, tt := range getTests {
		lru := NewCache(10)
		lru.Set(tt.keyToAdd, 1234)

		val, ok := lru.Get(tt.keyToGet)
		val, ok = lru.Get(tt.keyToGet)
		val, ok = lru.Get(tt.keyToGet)
		val, ok = lru.Get(tt.keyToGet)

		if ok != tt.expectedOk {
			t.Fatalf("%s: cache hit = %v; want %v %v", tt.name, ok, !ok, val)
		} else if ok && val != 1234 {
			t.Fatalf("%s expected get to return 1234 but got %v", tt.name, val)
		}
	}
}

//func TestRemove(t *testing.T) {
//	lru := NewCache(2)
//	lru.Set("myKey", 1234)
//	if val, ok := lru.Get("myKey"); !ok {
//		t.Fatal("TestRemove returned no match")
//	} else if val != 1234 {
//		t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
//	}
//
//	//lru.Remove("myKey")
//	if _, ok := lru.Get("myKey"); !ok {
//		t.Fatal("TestRemove returned a removed Entry")
//	}
//}

func TestEliminate(t *testing.T) {
	OnEliminateKeys := make([]string, 0)
	OnEliminateFun := func(key string, value any) {
		OnEliminateKeys = append(OnEliminateKeys, key)
	}

	maxitems := 20
	lru := NewCache(2, WithOnEliminate(OnEliminateFun))
	for i := 0; i < maxitems; i++ {
		lru.Set(fmt.Sprintf("myKey%d", i), 1234)
	}

	// visit more than two times
	for i := 0; i < 10; i++ {
		lru.Get("myKey1")
		lru.Get("myKey0")
	}

	for i := maxitems; i < 22; i++ {
		lru.Set(fmt.Sprintf("myKey%d", i), 1234)
	}

	for lru.Len() > maxitems {
		lru.RemoveOldest()
	}

	if len(OnEliminateKeys) != 2 {
		t.Fatalf("got %d evicted keys; want 2", len(OnEliminateKeys))
	}
	if OnEliminateKeys[0] != "myKey2" {
		t.Fatalf("got %v in first evicted key; want %s", OnEliminateKeys[0], "myKey2")
	}
	if OnEliminateKeys[1] != "myKey3" {
		t.Fatalf("got %v in second evicted key; want %s", OnEliminateKeys[1], "myKey3")
	}
}

func TestHotKey(t *testing.T) {
	OnEliminateKeys := make([]string, 0)
	OnEliminateFun := func(key string, value any) {
		OnEliminateKeys = append(OnEliminateKeys, key)
	}

	maxitems := 20
	lru := NewCache(2, WithOnEliminate(OnEliminateFun))
	for i := 0; i < maxitems; i++ {
		lru.Set(fmt.Sprintf("myKey%d", i), 1234)
	}

	// hot key
	for i := 0; i < 100; i++ {
		_, ok := lru.Get("myKey2")
		if !ok {
			log.Fatalf("got: No.%d no myKey2,want: myKey2", i)
		}
	}

	// visit only one time recently. fake hot key
	for i := 0; i < 20; i++ {
		v, ok := lru.Get(fmt.Sprintf("myKey%d", i))
		if !ok {
			log.Fatalf("got: %s,want: myKey%d", v.(*Entry).k, i)
		}
	}

	// eliminate
	for i := maxitems; i < 22; i++ {
		lru.Set(fmt.Sprintf("myKey%d", i), 1234)
	}

	for lru.Len() > maxitems {
		lru.RemoveOldest()
	}

	if len(OnEliminateKeys) != 2 {
		t.Fatalf("got %d evicted keys; want 1", len(OnEliminateKeys))
	}
	if OnEliminateKeys[0] != "myKey0" {
		t.Fatalf("got %v in first evicted key; want %s", OnEliminateKeys[0], "myKey0")
	}
}
