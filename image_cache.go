/*
Application to fetch files over HTTP and keep a local cache of them.
rbaynes Aug. 22, 2019
My first Go application.
*/

package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "crypto/md5"
    "./utils"
)

const (
    // Application defaults
    HOST = "static.rbxcdn.com"
    URL  = "/images/landing/Rollercoaster/whatsroblox_12072017.jpg"

    // HTTP headers 
    IF_NONE_MATCH       = "If-None-Match"
    IF_MODIFIED_SINCE   = "If-Modified-Since"
    ETAG                = "Etag"
    LAST_MODIFIED       = "Last-Modified"

    // Cache keys
    FILE_BYTES          = "file_bytes"
    FILE_HASH           = "file_hash"
)

// Arguments: hostname and URL
// Returns: status, response and file contents
func http_get(host string,
              URL string,
              headers map[string]string) (int, http.Header, []byte) {

    req, err := http.NewRequest("GET", "https://" + HOST + URL, nil)
    for key, value := range headers {
        req.Header.Add(key, value)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if nil != err {
        fmt.Println("Error:", err)
        return 0, resp.Header, []byte("")
    }

    defer resp.Body.Close()  // close response after we have read all the data

    body, err := ioutil.ReadAll(resp.Body)
    if nil != err {
        fmt.Println("Error:", err)
        return 0, resp.Header, body
    }
    return resp.StatusCode, resp.Header, body
}

func main() {
    // Our header and file content cache.
    cache := cache.New()

    // Make sure we have the same file contents.
    fetched_file_hash := ""
    cached_file_hash := ""

    // Do this two times, first to get the file, second to see if we can use
    // our cached version.
    for i := 0; i < 2; i++ {

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
        status, headers, file_bytes := http_get(HOST, URL, req_headers)

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
            fmt.Println("Fetched and cached the file!")
            bytes := md5.Sum(file_bytes)
            fetched_file_hash = string(bytes[:])
        } else if 304 == status {
            // Server says use our cached version
            file_bytes = cache.GetFile(URL)
            fmt.Println("Got the file from the cache!")
            bytes := md5.Sum(file_bytes)
            cached_file_hash = string(bytes[:])
        } else {
            fmt.Println("Error: unhandled status code:", status)
            return
        }

    }
    if fetched_file_hash != cached_file_hash {
        fmt.Println("Error: the file we fetched does not match the cached file")
    }
}




