package ebidlocal

import (
	"net/http"
	"net/url"
)

//HTTPClient client interface for making requests.
type hTTPClient interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
}

var client hTTPClient

func init() {
	client = http.DefaultClient
}
