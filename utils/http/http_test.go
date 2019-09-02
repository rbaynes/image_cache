// Test Package
package http_test

import (
	"." // imports the current directory so we get the http package
	"testing"
)

func TestHttp(t *testing.T) {
	host := "github.com"
	URL := "/rbaynes/image_cache/blob/master/README.md"
	req_headers := make(map[string]string)
	status, headers, file_bytes := http.HTTP_GET(host, URL, req_headers)
	if status != 200 {
		t.Errorf("Expected status 200, got %d", status)
	}
	if 0 == len(file_bytes) {
		t.Errorf("Expected file size")
	}
	if 0 == len(headers) {
		t.Errorf("Expected headers")
	}
}
