package redis

import (
	"fmt"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	cfg := Config{
		Address:  "10.10.10.3:6379",
		Prefix:   "test",
		Password: "QCWuNFS3787on2",
	}
	c, _ := Init(WithRedisConfig(cfg))

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
	fmt.Println(b)
	if b != nil && b.(float64) != 1 {
		t.Error("Got a failed value from mem cache:", b)
	}

	pullB, err := c.Pull("b")
	if pullB != nil && pullB.(float64) != 1 {
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
