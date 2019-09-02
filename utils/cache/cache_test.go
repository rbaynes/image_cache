// Test Package cache
package cache_test

import (
	"." // imports this current directory so we get the cache package
	"bytes"
	"testing"
)

// Testing for this package
func TestCache(t *testing.T) {

	size := 40
	verbose := true
	cache := cache.New(size, &verbose)
	if cache.UsedBytes() != 0 {
		cache.Print()
		t.Errorf("Expected 0, got %d", cache.UsedBytes())
	}

	key := "url1"
	header := "header1"
	value := "abcdefghij" // 10
	cache.SetHeader(key, header, value)
	if cache.UsedBytes() != 10 {
		cache.Print()
		t.Errorf("SetHeader 1, Expected 10, got %d", cache.UsedBytes())
	}

	if value != cache.GetHeader(key, header) {
		cache.Print()
		t.Errorf("GetHeader, Expected %s, got %s", value,
			cache.GetHeader(key, header))
	}

	// set same header again, should not use any more space
	cache.SetHeader(key, header, value)
	if cache.UsedBytes() != 10 {
		cache.Print()
		t.Errorf("SetHeader 2, Expected 10, got %d", cache.UsedBytes())
	}

	file_bytes := []byte("0123456789") // 10
	cache.SetFile(key, file_bytes)
	if cache.UsedBytes() != 20 {
		cache.Print()
		t.Errorf("SetFile 1, Expected 20, got %d", cache.UsedBytes())
	}

	if !bytes.Equal(file_bytes, cache.GetFile(key)) {
		cache.Print()
		t.Errorf("GetFile, Expected %s, got %s", file_bytes,
			cache.GetFile(key))
	}

	// set same file again, should not use any more space
	cache.SetFile(key, file_bytes)
	if cache.UsedBytes() != 20 {
		cache.Print()
		t.Errorf("SetFile 2, Expected 20, got %d", cache.UsedBytes())
	}

	// fill the cache to see if it handles eviction
	cache.SetHeader("url2", header, value)
	if cache.UsedBytes() != 30 {
		cache.Print()
		t.Errorf("Expected 30, got %d", cache.UsedBytes())
	}
	cache.SetHeader("url3", header, value)
	if cache.UsedBytes() != 20 {
		cache.Print()
		t.Errorf("Expected 20, got %d", cache.UsedBytes())
	}
}
