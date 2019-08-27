// Test Package cache
package cache_test

import (
	"." // imports this current directory so we get the cache 'class'
	"fmt"
	"testing"
)

// Testing for this package
func TestCache(t *testing.T) {
	fmt.Println("testing...")
	cache := cache.New(100)
	cache.Print()
	t.Errorf("hi")
	cached_etag := cache.GetHeader(URL, ETAG)
	cache.SetHeader(URL, ETAG, etag[0])
	cache.SetFile(URL, file_bytes)
	file_bytes = cache.GetFile(URL)
	//debugrob:
}
