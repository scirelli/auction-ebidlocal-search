package ebidlocal

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	ebid "github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
)

type mockClient struct {
	postForm func(url string, data url.Values) (resp *http.Response, err error)
	get      func(url string) (resp *http.Response, err error)
}

func (m *mockClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return m.postForm(url, data)
}

func (m *mockClient) Get(url string) (resp *http.Response, err error) {
	return m.get(url)
}

func TestSearchAuction(t *testing.T) {
	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("hello world")),
				StatusCode: 200,
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return nil, nil
		},
	}
	result, err := SearchAuction("auction1", []string{"hi", "there"})
	expected := ""
	assert.Nilf(t, err, "Should not return an error", err)
	if result != expected {
		t.Errorf("'%v' not equal '%v'", result, expected)
	}

	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("hello world")),
				StatusCode: 200,
			}, errors.New("some error")
		},
		get: func(url string) (resp *http.Response, err error) {
			return nil, nil
		},
	}
	result, err = SearchAuction("auction1", []string{"hi", "there"})
	expected = ""
	assert.NotNilf(t, err, "Should not return an error", err)
	if result != expected {
		t.Errorf("'%v' not equal '%v'", result, expected)
	}

	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("hello world")),
				StatusCode: 404,
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return nil, nil
		},
	}
	result, err = SearchAuction("auction1", []string{"hi", "there"})
	expected = ""
	assert.Nilf(t, err, "Should not return an error", err)
	if result != expected {
		t.Errorf("'%v' not equal '%v'", result, expected)
	}

	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
<html>
	<body>
		<table id="DataTable">
			<tbody><tr><td>Some data</td></tr></tbody>
		</table>
	</body>
</html>
`)),
				StatusCode: 200,
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return nil, nil
		},
	}
	result, err = SearchAuction("auction1", []string{"hi", "there"})
	expected = "<tr><td>Some data</td></tr>"
	assert.Nilf(t, err, "Should not return an error", err)
	if result != expected {
		t.Errorf("'%v' not equal '%v'", result, expected)
	}
}

func TestSearchAuctions(t *testing.T) {
	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
<html>
	<body>
		<table id="DataTable">
			<tbody><tr><td>Some data</td></tr></tbody>
		</table>
	</body>
</html>
`)),
				StatusCode: 404,
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
<div class="widget_ebid_current_widget">
	<div class="widgetOuter">
		<a href="some link"></a>
	</div>
</div>`)),
				StatusCode: 200,
			}, nil
		},
	}
}
