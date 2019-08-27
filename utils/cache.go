// Package cache implements a cache on top of a map (dict/hashtable)
// A Least Recently Used (LRU) list of keys is maintained, so that when
// the cache is about to exceed its maximum capacity, it can evict the
// LRU item and recover its space.
package cache

import (
	"container/list" // used to implement the LRU list, O(1)
	"fmt"
)

// A single item in the cache.
type cache_item struct {
	headers    map[string]string // header map of key: value
	file_bytes []byte            // file bytes
	item_size  int               // memory used by this item
}

// Members of the cache 'class'.
type cache struct {
	keys          map[string]cache_item // map of URL: item
	max_bytes     int
	current_bytes int
	LRU           *list.List // Least Recently Used list of keys
}

// Construct a new cache and return it.
func New(max_bytes int) cache {
	keys := make(map[string]cache_item)
	LRU := list.New()
	c := cache{keys, max_bytes, 0, LRU}
	return c
}

// Print the cache contents.
func (c *cache) Print() {
	multiple := "s"
	if 1 == c.LRU.Len() {
		multiple = ""
	}
	fmt.Printf("%d Cache Item%s:\n", c.LRU.Len(), multiple)
	for key, item := range c.keys {
		fmt.Printf("  %s\n", key)
		fmt.Println("    headers:", item.headers)
		fmt.Println("    file len:", len(item.file_bytes))
	}
	fmt.Println("LRU list (last key is LRU):")
	for e := c.LRU.Front(); e != nil; e = e.Next() {
		fmt.Println("  ", e.Value)
	}
	fmt.Println("Max cache size:", c.max_bytes, "bytes")
	fmt.Println("  Current size:", c.current_bytes, "bytes")
	fmt.Println("        Unused:", c.max_bytes-c.current_bytes, "bytes")
}

// Set a header into the cache
func (c *cache) SetHeader(key string, subkey string, value string) {
	c.checkSizeAndEvict(len(value))        // Do we need to make room?
	if item, found := c.keys[key]; found { // key exists, so add subkey
		item.headers[subkey] = value
		item.item_size += len(value)
	} else { // key not found, so add it
		headers := make(map[string]string)
		headers[subkey] = value
		item := cache_item{headers: headers, item_size: len(value)}
		c.keys[key] = item
	}
	c.current_bytes += len(value) // Total bytes in the cache
	// Put this key in the front (most recently used) spot in the LRU list.
	c.addHeadToLRU(key)
}

// Set a file into the cache
func (c *cache) SetFile(key string, value []byte) {
	c.checkSizeAndEvict(len(value))        // Do we need to make room?
	if item, found := c.keys[key]; found { // key exists, so set value
		item.file_bytes = value
		item.item_size += len(value)
		c.keys[key] = item // replace item with updated one
	} else { // key not found, so add it
		headers := make(map[string]string)
		item := cache_item{headers: headers, file_bytes: value,
			item_size: len(value)}
		c.keys[key] = item
	}
	c.current_bytes += len(value) // Total bytes in the cache
	// Put this key in the front (most recently used) spot in the LRU list.
	c.addHeadToLRU(key)
}

// Get a header by key from the cache.
func (c *cache) GetHeader(key string, subkey string) string {
	c.addHeadToLRU(key) // Move this key to the front of the LRU list.
	return c.keys[key].headers[subkey]
}

// Get a file by key from the cache.
func (c *cache) GetFile(key string) []byte {
	c.addHeadToLRU(key) // Move this key to the front of the LRU list.
	return c.keys[key].file_bytes
}

// If the key is in the list, remove it.
// This function is not exported, so like private, since it starts with a
// lower case letter.
func (c *cache) removeFromLRU(key string) {
	// (The linear search below will be slow if many items in the list.)
	for i := c.LRU.Front(); i != nil; i = i.Next() {
		if key == i.Value {
			c.LRU.Remove(i)
			return
		}
	}
}

// Put this key at the head of the list (make it the Most recently used).
func (c *cache) addHeadToLRU(key string) {
	// If the key is already in the list in another position, remove it.
	c.removeFromLRU(key)
	// Put the key at the front of the list.
	c.LRU.PushFront(key)
}

// Returns the key of the LRU item, or an empty string if the LRU list is empty.
func (c *cache) getLRU() string {
	if 0 == c.LRU.Len() {
		return ""
	}
	i := c.LRU.Back() // last item in list, so LRU
	ret := i.Value
	c.LRU.Remove(i)     // remove the LRU item
	return ret.(string) // convert List Element interface to string
}

// Will adding this value exceed our our capacity?
// If so, evict the LRU item to make room.
// Arguments: the size of the item to add.
func (c *cache) checkSizeAndEvict(value_size int) {
	if c.current_bytes+value_size < c.max_bytes {
		return // Enough space in the cache, so nothing to do.
	}

	// Capacity will be exceeded, so get and evict the LRU
	lru := c.getLRU()
	if "" == lru {
		return // The list/cache is empty
	}

	// Recover the bytes used by this item
	recovered_bytes := c.keys[lru].item_size
	c.current_bytes -= recovered_bytes
	if 0 > c.current_bytes {
		c.current_bytes = 0
	}

	// Remove the LRU from the cache dict
	delete(c.keys, lru) // delete is a built in, works on maps.

	// Remove the LRU from the list
	c.removeFromLRU(lru)
	fmt.Println("Evicted", lru, "from the cache and recovered", recovered_bytes,
		"bytes.")
}
