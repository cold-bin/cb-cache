package safe

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Once("key", func() (interface{}, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	if err != nil {
		t.Errorf("Do error = %v", err)
	}
}

func TestDoErr(t *testing.T) {
	var g Group
	someErr := errors.New("some error")
	v, err := g.Once("key", func() (interface{}, error) {
		return nil, someErr
	})
	if err != someErr {
		t.Errorf("Do error = %v; want someErr", err)
	}
	if v != nil {
		t.Errorf("unexpected non-nil value %#v", v)
	}
}

func TestDoDupSuppress(t *testing.T) {
	var g Group
	c := make(chan string)
	var calls int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Once("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
			}
			fmt.Println(v)
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // let goroutines above block
	c <- "bar"
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of calls = %d; want 1", got)
	}
}

// simulate extreme concurrency scenarios
func TestMaxgVisit(t *testing.T) {
	g := &Group{maxg: 10}
	calls := 100000
	flag := false
	mu := &sync.RWMutex{}
	wg := sync.WaitGroup{}
	for i := 0; i < calls; i++ {
		wg.Add(1)
		i_ := i
		go func() {
			ans, err := g.Once("key", func() (any, error) {
				return i_, nil
			})
			mu.Lock()
			flag = flag || err != nil
			mu.Unlock()
			fmt.Println(ans)
		}()
		wg.Done()
	}
	wg.Wait()

	if !flag {
		t.Error("should have error: server busy")
	}
}
