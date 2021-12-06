package fastchan

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

var (
	value = []byte("test-value")
)

func TestFastChan(t *testing.T) {
	c := NewFastChan(4096)
	c.Put(value)

	if c.Size() != 1 {
		t.Errorf("FastChan has the wrong size [%v]", c.Size())
	}

	val := c.Read()
	if !bytes.Equal(val.([]byte), value) {
		t.Errorf("Values do not match: [%v] <>[%v]", val, value)
	}

	for i := 0; i < 4096; i++ {
		c = NewFastChan(4096)
		for j := 0; j < i; j++ {
			c.Put(value)
		}

		if c.Size() != uint64(i) {
			t.Errorf("FastChan has the wrong size [%v]. Should be [%v]", c.Size(), i)
		}

		for j := 0; j < i; j++ {
			c.Read()
		}
	}

	if !c.IsEmpty() {
		t.Errorf("FastChan should be empty")
	}
}

func testFastChanPutGetOrder(t *testing.T, grp, cnt int) (
	residue int) {
	var wg sync.WaitGroup
	var idPut, idGet int32
	wg.Add(grp)
	c := NewFastChan(4096)
	for i := 0; i < grp; i++ {
		go func(g int) {
			defer wg.Done()
			for j := 0; j < cnt; j++ {
				idPut++
				id := []byte(fmt.Sprintf("%d", atomic.LoadInt32(&idPut)))
				o := value
				o = append(o, id...)
				c.Put(o)
			}
		}(i)
	}
	wg.Wait()
	wg.Add(grp)
	for i := 0; i < grp; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < cnt; {
				val := c.Read().([]byte)
				j++
				idGet++
				if fmt.Sprintf("%s%d", string(value), idGet) != string(val) {
					t.Errorf("Get.Err %s%d <> %s\n", string(value), idGet, string(val))
				}
			}
		}()
	}
	wg.Wait()
	return
}

func TestFastChanPutGetOrder(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	grp := 1
	cnt := 100

	testFastChanPutGetOrder(t, grp, cnt)
	t.Logf("Grp: %d, Times: %d", grp, cnt)
}

func BenchmarkFastChanPut16(b *testing.B) {
	benchmarkFastChanPut(16, b)
}
func BenchmarkFastChanPut64(b *testing.B) {
	benchmarkFastChanPut(64, b)
}
func BenchmarkFastChanPut256(b *testing.B) {
	benchmarkFastChanPut(256, b)
}
func BenchmarkFastChanPut1024(b *testing.B) {
	benchmarkFastChanPut(1024, b)
}
func BenchmarkFastChanPut4096(b *testing.B) {
	benchmarkFastChanPut(4096, b)
}
func BenchmarkFastChanPut16384(b *testing.B) {
	benchmarkFastChanPut(16384, b)
}
func BenchmarkFastChanPut65536(b *testing.B) {
	benchmarkFastChanPut(65536, b)
}
func BenchmarkFastChanPut262144(b *testing.B) {
	benchmarkFastChanPut(262144, b)
}
func BenchmarkFastChanPut1048576(b *testing.B) {
	benchmarkFastChanPut(1048576, b)
}

var mu sync.Mutex

func benchmarkFastChanPut(size int64, b *testing.B) {
	b.StopTimer()
	c := NewFastChan(4096)
	b.SetBytes(size)

	mu.Lock()
	n := b.N
	mu.Unlock()

	go func() {
		for i := 0; i < n; i++ {
			c.Read()
		}
	}()

	data := make([]byte, size)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Put(data)
	}

}

func BenchmarkFastChanGet16(b *testing.B) {
	benchmarkFastChanGet(16, b)
}
func BenchmarkFastChanGet64(b *testing.B) {
	benchmarkFastChanGet(64, b)
}
func BenchmarkFastChanGet256(b *testing.B) {
	benchmarkFastChanGet(256, b)
}
func BenchmarkFastChanGet1024(b *testing.B) {
	benchmarkFastChanGet(1024, b)
}
func BenchmarkFastChanGet4096(b *testing.B) {
	benchmarkFastChanGet(4096, b)
}
func BenchmarkFastChanGet16384(b *testing.B) {
	benchmarkFastChanGet(16384, b)
}
func BenchmarkFastChanGet65536(b *testing.B) {
	benchmarkFastChanGet(65536, b)
}
func BenchmarkFastChanGet262144(b *testing.B) {
	benchmarkFastChanGet(262144, b)
}
func BenchmarkFastChanGet1048576(b *testing.B) {
	benchmarkFastChanGet(1048576, b)
}

func benchmarkFastChanGet(size int64, b *testing.B) {
	b.StopTimer()
	c := NewFastChan(4096)
	b.SetBytes(size)

	mu.Lock()
	n := b.N
	mu.Unlock()

	data := make([]byte, size)
	go func() {
		for i := 0; i < n; i++ {
			c.Put(data)
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Read()
	}
}
