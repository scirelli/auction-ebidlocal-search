package fixtures

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
)

type MockClient struct {
	PostFormFunc func(url string, data url.Values) (resp *http.Response, err error)
	GetFunc      func(url string) (resp *http.Response, err error)
	DoFunc       func(req *http.Request) (*http.Response, error)
}

func NewMockClient(
	postForm func(url string, data url.Values) (resp *http.Response, err error),
	get func(url string) (resp *http.Response, err error),
) *MockClient {
	return &MockClient{
		PostFormFunc: postForm,
		GetFunc:      get,
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Not Implemented")
		},
	}
}

func (m *MockClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return m.PostFormFunc(url, data)
}

func (m *MockClient) Get(url string) (resp *http.Response, err error) {
	return m.GetFunc(url)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func OpenFile(t *testing.T, fileName string) io.ReadCloser {
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
