package mem

import (
	"strconv"
	"sync"
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
		t.Error(err)
	}

	if decrement != 9 {
		t.Error("decrement failed")
	}

	err = c.Flush()
	if err != nil {
		t.Error(err)
	}

	e, err := c.Get("e")
	if e != nil {
		t.Error("Got a failed value from mem cache:", e)
	}

	// 测试Forever方法
	// Test Forever method
	err = c.Forever("forever", "value")
	if err != nil {
		t.Error(err)
	}

	foreverVal, err := c.Get("forever")
	if err != nil || foreverVal == nil || foreverVal.(string) != "value" {
		t.Error("Forever method failed:", foreverVal, err)
	}

	// Check if the permanent cache still exists after waiting for a period of time
	time.Sleep(2 * time.Second)
	foreverVal, err = c.Get("forever")
	if err != nil || foreverVal == nil || foreverVal.(string) != "value" {
		t.Error("Forever value should persist:", foreverVal, err)
	}

	// Test error handling of Decrement method
	err = c.Put("string_val", "not_an_int", 10)
	if err != nil {
		t.Error(err)
	}

	_, err = c.Decrement("string_val", 1)
	if err == nil {
		t.Error("Decrement should fail with non-integer value")
	}

	// Test non-existent key
	_, err = c.Decrement("non_existent", 1)
	if err == nil {
		t.Error("Decrement should fail with non-existent key")
	}
}

// Test concurrent safety
func TestConcurrentAccess(t *testing.T) {
	c := Init()
	var wg sync.WaitGroup

	// Concurrent writing
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i)
			err := c.Put(key, i, 30)
			if err != nil {
				t.Error("Concurrent Put failed:", err)
			}
		}(i)
	}

	// Concurrent reading
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i)
			_, err := c.Get(key)
			if err != nil {
				t.Error("Concurrent Get failed:", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify data integrity
	for i := 0; i < 100; i++ {
		key := "key" + strconv.Itoa(i)
		val, err := c.Get(key)
		if err != nil || val == nil {
			t.Errorf("Expected value for key %s, got nil or error: %v", key, err)
		}
	}
}

// Test persistence of Forever method
func TestForeverPersistence(t *testing.T) {
	c := Init()

	err := c.Forever("persistent", 42)
	if err != nil {
		t.Error(err)
	}

	// Simulate time passing
	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		val, err := c.Get("persistent")
		if err != nil || val == nil || val.(int) != 42 {
			t.Errorf("Forever value should persist after %d checks: %v, %v", i, val, err)
		}
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

func BenchmarkDeleteMap(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
	b.ResetTimer()
	b.Logf("map count:%d", len(m))
	for i := 0; i < b.N; i++ {
		delete(m, i)
	}

	b.Logf("after clear map count:%d", len(m))
}

func BenchmarkClearMap(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < b.N; i++ {
		m[i] = i
	}

	b.ResetTimer()
	b.Logf("map count:%d", len(m))
	clear(m)
	b.Logf("after clear map count:%d", len(m))
}

// Modified to a more reasonable test data volume
func BenchmarkCache_Flush(b *testing.B) {
	c := Init()

	itemCount := 100000
	for i := 0; i < itemCount; i++ {
		c.Add(strconv.Itoa(i), i, 10)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Flush()
		// Refill data for the next test
		if i < b.N-1 {
			for j := 0; j < itemCount; j++ {
				c.Add(strconv.Itoa(j), j, 10)
			}
		}
	}
}

// Add benchmark test for Forever method
func BenchmarkForever(b *testing.B) {
	c := Init()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Forever(strconv.Itoa(i), i)
	}
}

// Add concurrent benchmark test
func BenchmarkConcurrentAccess(b *testing.B) {
	c := Init()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			key := strconv.Itoa(counter)
			c.Put(key, counter, 10)
			c.Get(key)
			counter++
		}
	})
}

// Test cache performance with different sizes
func BenchmarkLargeCache(b *testing.B) {
	c := Init()
	// Prefill with large amount of data
	for i := 0; i < 100000; i++ {
		c.Put(strconv.Itoa(i), i, 60)
	}

	b.ResetTimer()

	// Random access
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i % 100000)
		c.Get(key)
	}
}
