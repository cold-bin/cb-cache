package serialization

import (
	"strconv"
	"sync"
	"testing"
)

func TestGob_ConcurrencySerialize(t *testing.T) {
	type MM struct {
		A int
		B string
		C map[int]string
	}

	ms := make([]*MM, 1000)
	for i := 0; i < 1000; i++ {
		ms[i] = &MM{
			A: i + 1,
			B: strconv.Itoa(i + 1),
			C: map[int]string{i + 1: strconv.Itoa(i + 1)},
		}
	}
	gob := &Gob{}
	var wg sync.WaitGroup
	for i := 0; i < len(ms); i++ {
		wg.Add(1)
		i := i
		func() {
			var m2 = MM{}
			bs, err := gob.Marshal(ms[i])
			if err != nil {
				t.Error(err)
				return
			}

			if err := gob.Unmarshal(bs, &m2); err != nil {
				t.Error(err)
				return
			}
			//fmt.Println(m2)
			if m2.A != i+1 {
				t.Error("failed")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
