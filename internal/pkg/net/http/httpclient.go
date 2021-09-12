package http

import (
	gohttp "net/http"
	"net/url"
)

//HTTPClient client interface for making requests.
type HTTPClient interface {
	PostForm(url string, data url.Values) (resp *gohttp.Response, err error)
	Get(url string) (resp *gohttp.Response, err error)
	Do(req *gohttp.Request) (*gohttp.Response, error)
}
