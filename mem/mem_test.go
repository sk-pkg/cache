package mem

import (
	"strconv"
	"testing"
	"time"
)

func TestMemCache(t *testing.T) {
	c := Init()

	a, err := c.Get("a")
	if a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, err := c.Get("b")
	if a != nil {
		t.Error("Getting A found value that shouldn't exist:", b)
	}

	d, err := c.Get("d")
	if a != nil {
		t.Error("Getting A found value that shouldn't exist:", d)
	}

	err = c.Put("a", "a", 1)
	if err != nil {
		t.Error(err)
	}

	a, err = c.Get("a")
	if a == nil || a.(string) != "a" {
		t.Error("Got a failed value from mem cache:", a)
	}

	time.Sleep(2 * time.Second)
	a, err = c.Get("a")
	if err != nil {
		t.Error(err)
	}

	if a != nil {
		t.Error("Got a failed value from mem cache:", a)
	}

	err = c.Put("b", 1, 10)
	if err != nil {
		t.Error(err)
	}

	err = c.Add("b", 2, 10)
	if err != nil {
		t.Error(err)
	}

	b, err = c.Get("b")
	if b != nil && b.(int) != 1 {
		t.Error("Got a failed value from mem cache:", b)
	}

	pullB, err := c.Pull("b")
	if pullB != nil && pullB.(int) != 1 {
		t.Error("Got a failed value from mem cache:", pullB)
	}

	b, err = c.Get("b")
	if b != nil {
		t.Error("Got a failed value from mem cache:", b)
	}

	existed := c.Has("d")
	if existed {
		t.Error("Gets a key that should not exist")
	}

	err = c.Put("d", "D", 20)
	if err != nil {
		t.Error(err)
	}

	existed = c.Has("d")
	if !existed {
		t.Error("Gets a key that should exist")
	}

	_, _ = c.Forget("d")
	existed = c.Has("d")
	if existed {
		t.Error("Gets a key that should not exist")
	}

	increment, err := c.Increment("e", 1)
	if err != nil {
		t.Error(err)
	}

	if increment != 1 {
		t.Error("increment failed")
	}

	increment, err = c.Increment("e", 10)
	if err != nil {
		t.Error(err)
	}

	if increment != 11 {
		t.Error("increment failed")
	}

	decrement, err := c.Decrement("e", 2)
	if err != nil {
		return
	}

	if decrement != 9 {
		t.Error("decrement failed")
	}

	err = c.Flush()
	if err != nil {
		return
	}

	e, err := c.Get("e")
	if e != nil {
		t.Error("Got a failed value from mem cache:", e)
	}
}

func BenchmarkMemPut(b *testing.B) {
	c := Init()
	for i := 0; i < b.N; i++ {
		c.Put(strconv.Itoa(i), i, 10)
	}
}

func BenchmarkMemAdd(b *testing.B) {
	c := Init()
	for i := 0; i < b.N; i++ {
		c.Add(strconv.Itoa(i), i, 10)
	}
}

func BenchmarkGet(b *testing.B) {
	c := Init()
	for i := 0; i < b.N; i++ {
		c.Add(strconv.Itoa(i), i, 10)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Get(strconv.Itoa(i))
	}
}

func BenchmarkHas(b *testing.B) {
	c := Init()
	for i := 0; i < b.N; i++ {
		c.Add(strconv.Itoa(i), i, 10)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Has(strconv.Itoa(i))
	}
}

func BenchmarkPut(b *testing.B) {
	c := Init()
	for i := 0; i < b.N; i++ {
		c.Add(strconv.Itoa(i), i, 10)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Put(strconv.Itoa(i), i+1, 5)
	}
}
