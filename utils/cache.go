/*
rbaynes Aug. 22, 2019
Go 'class' that implements a cache on top of a map (dict/hashtable)
*/

package cache

import (
    "fmt"
)

type cache_item struct {                  // one item in the cache
    headers     map[string]string   // header map of key: value
    file_bytes  []byte              // file bytes
}

type cache struct {
    keys        map[string]cache_item     // map of URL: item
}

// Construct a new cache and return it.
func New() cache {
    keys := make(map[string]cache_item)
    c := cache {keys}
    return c
}

// Print the cache contents.
func (c cache) Print() {
    fmt.Println("Cache:")
    for key, item := range c.keys {
        fmt.Printf("  %s\n", key)
        fmt.Println("    headers:", item.headers)
        fmt.Println("    file len:", len(item.file_bytes))
    }
}

// Set a header into the cache
func (c cache) SetHeader(key string, subkey string, value string) {
    if item, found := c.keys[key]; found {  // key exists, so add subkey
        item.headers[subkey] = value
    } else {                                // key not found, so add it
        headers := make(map[string]string)
        headers[subkey] = value
        item := cache_item {headers:headers}
        c.keys[key] = item
    }
}

// Set a file into the cache
func (c cache) SetFile(key string, value []byte) {
    if item, found := c.keys[key]; found {  // key exists, so set value
        item.file_bytes = value
        c.keys[key] = item                  // replace item with updated one
    } else {                                // key not found, so add it
        headers := make(map[string]string)
        item := cache_item {headers:headers, file_bytes:value}
        c.keys[key] = item
    }
}

func (c cache) GetHeader(key string, subkey string) (string) {
    return c.keys[key].headers[subkey]
}

func (c cache) GetFile(key string) ([]byte) {
    return c.keys[key].file_bytes
}



