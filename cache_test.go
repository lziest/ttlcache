package ttlcache

import (
	"fmt"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	c, err := NewLRU(10, 1*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

	// check internal structure
	elem, ok := c.table[k]
	if !ok {
		t.Fatal("Set didn't insert value")
	}
	item := elem.Value.(*entry)
	if item.key != "key" || item.value != "value" {
		t.Fatal("Bad key value stored")
	}
}

func TestDoubleSet(t *testing.T) {
	c, err := NewLRU(10, 1*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

	created = c.Set(k, v, 0)
	if created {
		t.Fatal("Set returns wrong value")
	}

}

func TestSetGet(t *testing.T) {
	c, err := NewLRU(10, 1*time.Hour, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

	got, stale := c.Get(k)
	if stale {
		t.Fatal("bad stale value returned")
	}

	if got != "value" {
		t.Fatal("bad value returned")
	}
}

func TestGetStale(t *testing.T) {
	c, err := NewLRU(10, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

	time.Sleep(100 * time.Millisecond)
	got, stale := c.Get(k)
	if !stale {
		t.Fatal("bad stale value returned")
	}

	if got != "value" {
		t.Fatal("bad value returned")
	}
}

func TestGetMiss(t *testing.T) {
	c, err := NewLRU(10, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

	got, stale := c.Get("badkey")
	if stale {
		t.Fatal("bad stale value returned")
	}

	if got != nil {
		t.Fatal("bad value returned")
	}
}

func TestSetRemoveGet(t *testing.T) {
	c, err := NewLRU(10, 50*time.Millisecond, nil)
	if err != nil {
		t.Fatal(err)
	}

	k := "key"
	v := "value"
	created := c.Set(k, v, 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}
	found := c.Remove("key")
	if !found {
		t.Fatal("bad delete")
	}

	got, stale := c.Get("key")
	if stale {
		t.Fatal("bad stale value returned")
	}

	if got != nil {
		t.Fatal("bad value returned")
	}
}

func TestLRU1(t *testing.T) {
	MustEvictKey0 := func(key string, value interface{}) {
		if key != "key-0" {
			t.Fatal("LRU failure: not the last element being evicted", key)
		}
	}

	num := 10
	c, err := NewLRU(num, 1*time.Hour, MustEvictKey0)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		k := fmt.Sprintf("key-%d", i)
		v := fmt.Sprintf("val-%d", i)
		created := c.Set(k, v, 0)
		if !created {
			t.Fatal("Set returns wrong value")
		}
	}

	created := c.Set("new-key", "new-value", 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

}

func TestLRU2(t *testing.T) {
	num := 10

	MustEvictKey := func(key string, value interface{}) {
		lastK := fmt.Sprintf("key-%d", num-1)
		if key != lastK {
			t.Fatal("LRU failure: not the desired element being evicted", key)
		}
	}

	c, err := NewLRU(num, 1*time.Hour, MustEvictKey)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		k := fmt.Sprintf("key-%d", i)
		v := fmt.Sprintf("val-%d", i)
		created := c.Set(k, v, 0)
		if !created {
			t.Fatal("Set returns wrong value")
		}
	}

	// reference all cached value except the last one
	for i := 0; i < num-1; i++ {
		k := fmt.Sprintf("key-%d", i)
		c.Get(k)
	}

	// add a new one to trigger eviction
	created := c.Set("new-key", "new-value", 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}

}

func TestStaleLRU(t *testing.T) {
	MustEvictKey0 := func(key string, value interface{}) {
		if key != "key-0" {
			t.Fatal("LRU failure: not the desired element being evicted", key)
		}
	}

	num := 10
	c, err := NewLRU(num, 100*time.Second, MustEvictKey0)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		k := fmt.Sprintf("key-%d", i)
		v := fmt.Sprintf("val-%d", i)
		var ttl time.Duration
		if i == 0 {
			ttl = 50 * time.Millisecond
		}
		created := c.Set(k, v, ttl)
		if !created {
			t.Fatal("Set returns wrong value")
		}
	}

	time.Sleep(100 * time.Millisecond)
	// Get a stale cached value will send it to the back of the list
	_, stale := c.Get("key-0")
	if !stale {
		t.Fatal("bad stale value")
	}

	created := c.Set("new-key", "new-value", 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}
}

func TestLRUTimer(t *testing.T) {
	num := 10
	MustEvictKeyNumMinusOne := func(key string, value interface{}) {
		lastK := fmt.Sprintf("key-%d", num-1)
		if key != lastK {
			t.Fatal("LRU failure: not the desired element being evicted", key)
		}
	}

	c, err := NewLRU(num, 100*time.Second, MustEvictKeyNumMinusOne)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		k := fmt.Sprintf("key-%d", i)
		v := fmt.Sprintf("val-%d", i)
		var ttl time.Duration
		if i == num-1 {
			ttl = 50 * time.Millisecond
		}
		created := c.Set(k, v, ttl)
		if !created {
			t.Fatal("Set returns wrong value")
		}
	}

	time.Sleep(100 * time.Millisecond)
	// Without touching the cache, a timer should move "key-9" to the end of the list for eviction.

	created := c.Set("new-key", "new-value", 0)
	if !created {
		t.Fatal("Set returns wrong value")
	}
}