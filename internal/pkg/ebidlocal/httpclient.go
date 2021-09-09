package ebidlocal

import (
	"net/http"
	"net/url"
)

//HTTPClient client interface for making requests.
type HTTPClient interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

var Client HTTPClient

func init() {
	Client = http.DefaultClient
}
