package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/geseq/fastchan"
)

func benchmarkFastChanPut(size uint64, n int) {
	runtime.LockOSThread()
	c := fastchan.NewFastChan(size)

	go func() {
		runtime.LockOSThread()
		for i := 0; i < n; i++ {
			c.Read()
		}
	}()

	start := time.Now()
	for i := 0; i < n; i++ {
		c.Put(0)
	}

	duration := time.Since(start)
	fmt.Printf("BenchmarkFastChanPut%d\t%d\t%d ns/op\n", size, n, duration.Nanoseconds()/int64(n))
}

func benchmarkFastChanGet(size uint64, n int) {
	runtime.LockOSThread()
	c := fastchan.NewFastChan(size)

	go func() {
		runtime.LockOSThread()
		for i := 0; i < n; i++ {
			c.Put(0)
		}
	}()

	start := time.Now()
	for i := 0; i < n; i++ {
		c.Read()
	}
	duration := time.Since(start)
	fmt.Printf("BenchmarkFastChanGet%d\t%d\t%d ns/op\n", size, n, duration.Nanoseconds()/int64(n))
}

func main() {
	n := 5_000_000

	benchmarkFastChanPut(16, n)
	benchmarkFastChanPut(64, n)
	benchmarkFastChanPut(256, n)
	benchmarkFastChanPut(1024, n)
	benchmarkFastChanPut(4096, n)
	benchmarkFastChanPut(16_384, n)
	benchmarkFastChanPut(65_536, n)
	benchmarkFastChanPut(262_144, n)
	benchmarkFastChanPut(1_048_576, n)

	benchmarkFastChanGet(16, n)
	benchmarkFastChanGet(64, n)
	benchmarkFastChanGet(256, n)
	benchmarkFastChanGet(1024, n)
	benchmarkFastChanGet(4096, n)
	benchmarkFastChanGet(16_384, n)
	benchmarkFastChanGet(65_536, n)
	benchmarkFastChanGet(262_144, n)
	benchmarkFastChanGet(1_048_576, n)
}
