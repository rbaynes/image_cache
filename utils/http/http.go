/*
HTTP GET function.
*/

package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Arguments: hostname and URL
// Returns: status, response and file contents
func HTTP_GET(host string,
	URL string,
	headers map[string]string) (int, http.Header, []byte) {

	req, err := http.NewRequest("GET", "https://"+host+URL, nil)
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if nil != err {
		fmt.Println("Error:", err)
		return 0, resp.Header, []byte("")
	}

	defer resp.Body.Close() // close response after we have read all the data

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		fmt.Println("Error:", err)
		return 0, resp.Header, body
	}
	return resp.StatusCode, resp.Header, body
}
