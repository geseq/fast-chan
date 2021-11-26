package fastchan

import (
	"runtime"
	"sync/atomic"

	"golang.org/x/sys/cpu"
)

// Copyright 2012 Darren Elwood <darren@textnode.com> http://www.textnode.com @textnode
// Copyright 2021 E Sequeira
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// A fast chan replacement based on textnode/gringo
//
// N.B. To see the performance benefits of gringo versus Go's channels, you must have multiple goroutines
// and GOMAXPROCS > 1.

// Known Limitations:
//
// *) At most (2^64)-2 items can be written to the queue.
// *) The size of the queue must be a power of 2.
//
// Suggestions:
//
// *) If you have enough cores you can change from runtime.Gosched() to a busy loop.
//

// template type FastChan(CacheItem)
type CacheItem interface{}

// FastChan is a minimalist queue replacement for channels with higher throughput
type FastChan struct {
	// The padding members 1 to 5 below are here to ensure each item is on a separate cache line.
	// This prevents false sharing and hence improves performance.
	_                  cpu.CacheLinePad
	indexMask          uint64
	_                  cpu.CacheLinePad
	lastCommittedIndex uint64
	_                  cpu.CacheLinePad
	nextFreeIndex      uint64
	_                  cpu.CacheLinePad
	readerIndex        uint64
	_                  cpu.CacheLinePad
	contents           []CacheItem
	_                  cpu.CacheLinePad
}

// NewFastChan creates a new channel
func NewFastChan(size uint64) *FastChan {
	size = roundUpNextPowerOfTwo(size)
	return &FastChan{
		lastCommittedIndex: 0,
		nextFreeIndex:      0,
		readerIndex:        0,
		indexMask:          size - 1,
		contents:           make([]CacheItem, size),
	}
}

// Put writes a CacheItem to the front of the channel
func (c *FastChan) Put(value CacheItem) {
	var myIndex = atomic.AddUint64(&c.nextFreeIndex, 1)
	//Wait for reader to catch up, so we don't clobber a slot which it is (or will be) reading
	for myIndex > (atomic.LoadUint64(&c.readerIndex) + c.indexMask) {
		runtime.Gosched()
	}

	//Write the item into it's slot
	c.contents[myIndex&c.indexMask] = value

	//Increment the lastCommittedIndex so the item is available for reading
	for !atomic.CompareAndSwapUint64(&c.lastCommittedIndex, myIndex-1, myIndex) {
		runtime.Gosched()
	}
}

// Read reads and removes a CacheItem from the back of the channel
func (c *FastChan) Read() CacheItem {
	var myIndex = atomic.AddUint64(&c.readerIndex, 1)
	//If reader has out-run writer, wait for a value to be committed
	for myIndex > atomic.LoadUint64(&c.lastCommittedIndex) {
		runtime.Gosched()
	}
	return c.contents[myIndex&c.indexMask]
}

// Empty the channel
func (c *FastChan) Empty() {
	c.lastCommittedIndex = 0
	c.nextFreeIndex = 0
	c.readerIndex = 0
}

// Size gets the size of the contents in the channel buffer
func (c *FastChan) Size() uint64 {
	return atomic.LoadUint64(&c.lastCommittedIndex) - atomic.LoadUint64(&c.readerIndex)
}

// IsEmpty checks if the channel is empty
func (c *FastChan) IsEmpty() bool {
	return atomic.LoadUint64(&c.readerIndex) >= atomic.LoadUint64(&c.lastCommittedIndex)
}

// IsFull checks if the channel is full
func (c *FastChan) IsFull() bool {
	return atomic.LoadUint64(&c.nextFreeIndex) >= (atomic.LoadUint64(&c.readerIndex) + c.indexMask)
}

func roundUpNextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
