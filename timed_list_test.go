package timedlist_test

import (
	"testing"
	"time"
	"timedlist"
)

//nolint:funlen
func TestTimedListAdd(t *testing.T) {
	// create a new timed list
	tl := timedlist.NewTimedList()
	// add an element to the list
	t.Log("Checking if element is returned after duration")
	tl.Add("test", time.Second*1)
	// check if the element returns after the duration
	afterFuncAddTimer := time.AfterFunc(time.Second*2, func() { tl.C <- false })
	if <-tl.C != "test" {
		t.Error("element not returned after duration")
	}
	if !afterFuncAddTimer.Stop() {
		<-afterFuncAddTimer.C
	}
	// check if timer is correctly updated
	t.Log("Checking if timer is correctly updated")
	firstEntry := tl.Add("first", time.Second*4)
	tl.Add("second", time.Second*1)
	afterFuncAddTimer = time.AfterFunc(time.Second*2, func() { tl.C <- false })
	if <-tl.C != "second" {
		t.Error("Timer not correctly updated")
	}
	if !afterFuncAddTimer.Stop() {
		<-afterFuncAddTimer.C
	}
	// remove the first entry
	t.Log("Checking if timer is correctly updated after removing first entry")
	tl.Remove(firstEntry)
	_ = time.AfterFunc(time.Second*1, func() { tl.C <- false })
	if <-tl.C != false {
		t.Error("Timer not correctly updated")
	}
	// check if list is empty after removing last entry
	t.Log("Checking if list is empty after removing last entry")
	if tl.Len() != 0 {
		t.Error("List not empty after removing last entry")
	}

	// check if robust to multiple adds
	t.Log("Checking if robust to multiple adds")
	tl.Add("first", time.Second*1)
	tl.Add("second", time.Second*1+time.Millisecond*500)
	tl.Add("third", time.Second*1+time.Millisecond*200)
	afterFuncAddTimer = time.AfterFunc(time.Second*4, func() { tl.C <- false })
	if <-tl.C != "first" {
		t.Error("List does not return first element")
	}
	if <-tl.C != "third" {
		t.Error("List does not return third element")
	}
	if <-tl.C != "second" {
		t.Error("List does not return second element")
	}
	if !afterFuncAddTimer.Stop() {
		<-afterFuncAddTimer.C
	}
	// check if robust to entries with same duration
	t.Log("Checking if robust to entries with same duration")
	tl.Add("first", time.Second*1)
	tl.Add("second", time.Second*1)
	tl.Add("third", time.Second*1)
	afterFuncAddTimer = time.AfterFunc(time.Second*4, func() { tl.C <- false })
	for i := 0; i < 3; i++ {
		if <-tl.C == false {
			t.Error("List does not return three elements after draining")
		}
	}
	if !afterFuncAddTimer.Stop() {
		<-afterFuncAddTimer.C
	}

	// check if drain works
	t.Log("Checking if drain works")
	tl.Add("first", time.Second*1)
	tl.Add("second", time.Second*1)
	tl.Add("third", time.Second*1)
	tl.Drain()
	afterFuncAddTimer = time.AfterFunc(time.Second*4, func() { tl.C <- false })
	for i := 0; i < 3; i++ {
		if <-tl.C == false {
			t.Error("List does not return three elements after draining")
		}
	}
	if tl.Len() != 0 {
		t.Error("List not empty after draining")
	}
	if !afterFuncAddTimer.Stop() {
		<-afterFuncAddTimer.C
	}

	// check if purge works
	t.Log("Checking if purge works")
	tl.Add("first", time.Second*2)
	tl.Add("second", time.Second*2)
	tl.Add("third", time.Second*2)
	tl.Purge()
	_ = time.AfterFunc(time.Second*4, func() { tl.C <- false })
	if tl.Len() != 0 {
		t.Error("List not empty after purging")
	}
	if <-tl.C != false {
		t.Error("List does not return false after purging")
	}
}
