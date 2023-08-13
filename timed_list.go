/*The MIT License (MIT)

Copyright (c) 2023 Valentin Lemaire

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.*/

package timedlist

import (
	"container/list"
	"context"
	"github.com/petar/GoLLRB/llrb"
	"sync"
	"time"
)

type TimedEntry struct {
	v interface{}
	d time.Time
}

func (t TimedEntry) Less(than llrb.Item) bool {
	return t.d.Before(than.(*TimedEntry).d)
}

// Value returns the value of the entry
func (t TimedEntry) Value() interface{} {
	return t.v
}

type TimedList struct {
	C       chan interface{}
	l       list.List
	entries *llrb.LLRB
	t       *time.Timer
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
}

// NewTimedList creates a new TimedList
func NewTimedList() *TimedList {
	return &TimedList{
		C:       make(chan interface{}),
		l:       list.List{},
		entries: llrb.New(),
		t:       nil,
		ctx:     nil,
		cancel:  nil,
		mu:      sync.Mutex{},
	}
}

func (t *TimedList) run() {
	for {
		if t.t == nil {
			return
		}
		select {
		case <-t.ctx.Done():
			return
		case <-t.t.C:
			t.mu.Lock()
			min := t.entries.DeleteMin()
			t.mu.Unlock()
			go func() { t.C <- min.(*TimedEntry).v }()
			if t.entries.Len() == 0 {
				t.t = nil
				return
			}
			t.t.Reset(time.Until(t.entries.Min().(*TimedEntry).d))
		}
	}
}

// stop stops the TimedList
func (t *TimedList) stop() {
	if t.t != nil {
		if !t.t.Stop() {
			<-t.t.C
		}
		t.cancel()
		t.t = nil
	}
}

// Add adds v to the list with duration d
//
// It refreshes the timer if the new element will expire before the current timer
func (t *TimedList) Add(v interface{}, d time.Duration) (item *TimedEntry) {
	item = &TimedEntry{v, time.Now().Add(d)}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.t == nil || t.entries.Len() == 0 {
		t.t = time.NewTimer(d)
		t.entries.InsertNoReplace(item)
		t.ctx, t.cancel = context.WithCancel(context.Background())
		go t.run()
		return
	}
	if t.entries.Min().(*TimedEntry).d.After(time.Now().Add(d)) {
		t.t.Stop()
		t.t.Reset(d)
	}
	t.entries.InsertNoReplace(item)
	return
}

// Remove removes te from the list.
//
// It calls stop if the list is empty after removing the element.
//
// It returns the value of the element removed.
func (t *TimedList) Remove(te *TimedEntry) (v interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry := t.entries.Delete(te).(*TimedEntry)
	v = entry.v
	if t.entries.Len() == 0 {
		t.stop()
		return
	}
	minEntry := t.entries.Min().(*TimedEntry)
	if minEntry.d.After(entry.d) {
		t.t.Stop()
		t.t.Reset(time.Until(minEntry.d))
	}
	return
}

// Drain drains the list into the channel
//
// It calls stop after draining the list
func (t *TimedList) Drain() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stop()
	for t.entries.Len() > 0 {
		min := t.entries.DeleteMin()
		go func() {
			t.C <- min.(*TimedEntry).v
		}()
	}
}

// Purge removes all elements from the list and then calls stop
func (t *TimedList) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stop()
	for t.entries.Len() > 0 {
		t.entries.DeleteMin()
	}
}

// Len returns the number of elements in the list
func (t *TimedList) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	i := t.entries.Len()
	return i
}
