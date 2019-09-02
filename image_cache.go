/*
Application to fetch files over HTTP and keep a local cache of them.
rbaynes Aug. 22, 2019
My first Go application.
*/

package main

import (
	"./utils/cache"
	"./utils/http"
	"crypto/md5"
	"flag"
	"fmt"
)

const (
	// Application defaults
	HOST = "static.rbxcdn.com"
	URL1 = "/images/landing/Rollercoaster/whatsroblox_12072017.jpg"
	URL2 = "/images/landing/Rollercoaster/gameimage3_12072017.jpg"
	URL3 = "/images/landing/Rollercoaster/devices_people_12072017.png"

	// HTTP headers
	IF_NONE_MATCH     = "If-None-Match"
	IF_MODIFIED_SINCE = "If-Modified-Since"
	ETAG              = "Etag"
	LAST_MODIFIED     = "Last-Modified"

	// Cache keys
	FILE_BYTES = "file_bytes"
	FILE_HASH  = "file_hash"
)

func main() {
	// Command line args
	var pverbose = flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Our header and file content cache.
	cache := cache.New(200*1024, pverbose)

	// The list of files we fetch / cache.
	files := []string{URL1, URL2, URL3}
	fetched_file_hash := [3]string{"", "", ""}
	cached_file_hash := [3]string{"", "", ""}

	for f := 0; f < len(files); f++ { // Fetch each file

		URL := files[f]

		// Do this two times, first to get the file, second to see if we can use
		// our cached version.
		for i := 0; i < 2; i++ { // Fetch each two times, to test cache.

			// Headers we send with our request.
			req_headers := make(map[string]string)

			// First check our cache for the headers from the URL,
			// if we find them, add headers to our request
			cached_etag := cache.GetHeader(URL, ETAG)
			cached_last_mod := cache.GetHeader(URL, LAST_MODIFIED)
			if 0 < len(cached_etag) && 0 < len(cached_last_mod) {
				req_headers[IF_NONE_MATCH] = cached_etag
				req_headers[IF_MODIFIED_SINCE] = cached_last_mod
			}

			// Get the file
			status, headers, file_bytes := http.HTTP_GET(HOST, URL, req_headers)

			// Get and cache the headers we care about from the response
			etag := headers[ETAG]
			last_mod := headers[LAST_MODIFIED]
			if 0 < len(etag) {
				cache.SetHeader(URL, ETAG, etag[0])
			}
			if 0 < len(last_mod) {
				cache.SetHeader(URL, LAST_MODIFIED, last_mod[0])
			}

			if 200 == status {
				// We got the file!
				cache.SetFile(URL, file_bytes)
				fmt.Println("Fetched and cached:", URL)
				bytes := md5.Sum(file_bytes)
				fetched_file_hash[f] = string(bytes[:])
				if *pverbose {
					cache.Print()
				}

			} else if 304 == status {
				// Server says use our cached version
				file_bytes = cache.GetFile(URL)
				fmt.Println("Cache hit for:", URL)
				bytes := md5.Sum(file_bytes)
				cached_file_hash[f] = string(bytes[:])
				if *pverbose {
					cache.Print()
				}

			} else {
				fmt.Println("Error: unhandled status code:", status)
				return
			}
		}
		fmt.Println()
	}

	// Validate
	for f := 0; f < len(files); f++ {
		if fetched_file_hash[f] != cached_file_hash[f] {
			fmt.Println("Error: file hashes do not match")
		}
	}
	if *pverbose {
		cache.Print()
	}
}
