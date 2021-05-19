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

type TestCase struct {
	Body       string
	StatusCode int
	Error      error
	Auction    string
	Keywords   []string
	Expected   string
}

func TestSearchAuction(t *testing.T) {
	var tests map[string]TestCase = map[string]TestCase{
		"Search for key words that return no results": TestCase{
			Body:       "hello world",
			StatusCode: 200,
			Error:      nil,
			Auction:    "auction1",
			Keywords:   []string{"hi", "there"},
			Expected:   "",
		},
		"PostForm returns an error": TestCase{
			Body:       "hello world",
			StatusCode: 200,
			Error:      errors.New("some error"),
			Auction:    "auction1",
			Keywords:   []string{"hi", "there"},
			Expected:   "",
		},
		"PostForm gets a 404 response": TestCase{
			Body:       "hello world",
			StatusCode: 404,
			Error:      nil,
			Auction:    "auction1",
			Keywords:   []string{"hi", "there"},
			Expected:   "",
		},
		"Postform returns results": TestCase{
			Body: `
	<html>
		<body>
			<table id="DataTable">
				<tbody><tr><td>Some data</td></tr></tbody>
			</table>
		</body>
	</html>
	`,
			StatusCode: 200,
			Error:      nil,
			Auction:    "auction1",
			Keywords:   []string{"hi", "there"},
			Expected:   "<tr><td>Some data</td></tr>",
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			ebid.Client = &mockClient{
				postForm: func(url string, data url.Values) (resp *http.Response, err error) {
					return &http.Response{
						Body:       ioutil.NopCloser(strings.NewReader(test.Body)),
						StatusCode: test.StatusCode,
						Request:    &http.Request{},
					}, test.Error
				},
				get: func(url string) (resp *http.Response, err error) {
					return nil, nil
				},
			}
			result, err := SearchAuction(test.Auction, test.Keywords)
			expected := test.Expected
			assert.Equal(t, test.Error, err)
			assert.Equalf(t, expected, result, "'%v' not equal '%v'", result, expected)
		})
	}
}

func Skip_TestSearchAuctionRemovesDataThatChanges(t *testing.T) {
	expected := `
				<tr class="DataRow" id="270" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples556/270">270</a><br/><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples556/270">more<br/>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples556/270"><img border="0" src="https://ebid15.com/photos/202104-D1 Warehouse-1356/smalls/270.JPG" alt="smalls/270.JPG"/></a></td><td class="description"><b>Category</b>: DECORATIVE<br/><b>Item</b>: PROP TV MADE IN BRAZIL, METAL 48X28&#34;.  MISSING ONE BUTTON.<br/><b>Location</b>: FRONT RIGHT<br/></td>
					<td align="right" class="bids"><a href="/cgi-bin/mmhistory.cgi?staples556/270"><span id="270_bids">2</span></a></td>
					<td align="right" class="highbidder"></td>
					<td align="right" class="currentamount"></td>
					<td align="right" class="nextbidrequired"></td>
					<td align="center" class="yourbid"></td>
					<td align="center" class="yourmaximum"></td>
				</tr>
				<tr class="DataRow" id="1435" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples571/1435">1435</a><br/><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples571/1435">more<br/>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples571/1435"><img border="0" src="https://ebid15.com/photos/202104-HM-Kelleher-1371/smalls/1435.JPG" alt="smalls/1435.JPG"/></a></td><td class="description"><b>Category</b>: HOUSEHOLD<br/><b>Item</b>: SHOEHORNS, DRESSER VALET, ELECTRIC SHAVERS, TIMEX ALARM CLOCK, SHOESHINE BRUSHES, SONY TV WEATHER/AM/FM PERSONAL RADIO<br/><b>Location</b>: DINING ROOM<br/></td>
					<td align="right" class="bids"><span id="1435_bids"> </span></td>
					<td align="right" class="highbidder"></td>
					<td align="right" class="currentamount"></td>
					<td align="right" class="nextbidrequired"></td>
					<td align="center" class="yourbid"></td>
					<td align="center" class="yourmaximum"></td>
				</tr>
			`
	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
    <body bgcolor="#FFFFFF">
        <table id="DataTable" class="listbody" border="0" width="100%" align="center" cellpadding="3" cellspacing="1">
            <thead>
                <tr bgcolor="#073C68" valign="bottom">
                <th align="center" width="40"><font color="#FFFFFF"><strong>Item</strong></font></th>
                <th id="DataTablePhoto" align="center"><font color="#FFFFFF"><strong>Photo</strong></font></th>
                <th id="DataTableDesc" align="center"><font color="#FFFFFF"><strong>Description</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Bids</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>High<br>Bidder</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Current<br>Amount</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Next Bid<br>Required</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Your<br>Bid</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Your<br>Maximum</strong></font></th>
                </tr>
            </thead>

            <tbody>
				<tr class="DataRow" id="270" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples556/270">270</a><br><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples556/270">more<br>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples556/270"><img border="0" src="https://ebid15.com/photos/202104-D1 Warehouse-1356/smalls/270.JPG" alt="smalls/270.JPG"></a></td><td class="description"><b>Category</b>: DECORATIVE<br><b>Item</b>: PROP TV MADE IN BRAZIL, METAL 48X28".  MISSING ONE BUTTON.<br><b>Location</b>: FRONT RIGHT<br></td>
					<td align="right" class="bids"><a href="/cgi-bin/mmhistory.cgi?staples556/270"><span id="270_bids">2</span></a></td>
					<td align="right" class="highbidder"><span id="270_highbidder">21493</span></td>
					<td align="right" class="currentamount"><span id="270_currentprice">1.49</span></td>
					<td align="right" class="nextbidrequired"><span id="270_nextrequired"><a href="javascript:subfillform('270','1.99')">1.99</a></span></td>
					<td align="center" class="yourbid"><span id="270_yourbid"><input type="text" name="270" size="8" placeholder="your bid"></span></td>
					<td align="center" class="yourmaximum"><span id="270_yourmax"><input type="text" name="m270" size="8" placeholder="your max"> <br><i><a href="javascript:subbnpw()">submit bid</a></i></span><br><span id="270_endtime"></span><br><span id="270_status"></span></td>
				</tr>
				<tr class="DataRow" id="1435" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples571/1435">1435</a><br><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples571/1435">more<br>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples571/1435"><img border="0" src="https://ebid15.com/photos/202104-HM-Kelleher-1371/smalls/1435.JPG" alt="smalls/1435.JPG"></a></td><td class="description"><b>Category</b>: HOUSEHOLD<br><b>Item</b>: SHOEHORNS, DRESSER VALET, ELECTRIC SHAVERS, TIMEX ALARM CLOCK, SHOESHINE BRUSHES, SONY TV WEATHER/AM/FM PERSONAL RADIO<br><b>Location</b>: DINING ROOM<br></td>
					<td align="right" class="bids"><span id="1435_bids">&nbsp;</span></td>
					<td align="right" class="highbidder"><span id="1435_highbidder">&nbsp;</span></td>
					<td align="right" class="currentamount"><span id="1435_currentprice">&nbsp;</span></td>
					<td align="right" class="nextbidrequired"><span id="1435_nextrequired"><a href="javascript:subfillform('1435','0.99')">0.99</a></span></td>
					<td align="center" class="yourbid"><span id="1435_yourbid"><input type="text" name="1435" size="8" placeholder="your bid"></span></td>
					<td align="center" class="yourmaximum"><span id="1435_yourmax"><input type="text" name="m1435" size="8" placeholder="your max"> <br><i><a href="javascript:subbnpw()">submit bid</a></i></span><br><span id="1435_endtime"></span><br><span id="1435_status"></span></td>
				</tr>
			</tbody>
        </table>
	</body>
</html>
`)),
				StatusCode: 200,
				Request:    &http.Request{},
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("")),
				StatusCode: 200,
				Request:    &http.Request{},
			}, nil
		},
	}
	result, err := SearchAuction("auction1", []string{"hi", "there"})
	assert.Nil(t, err)
	assert.Equalf(t, expected, result, "%s", result)
}

func TestSearchAuctionRemovesDataThatChanges(t *testing.T) {
	expected := `
				<tr class="DataRow" id="270" valign="top">
					<td class="item"><a href="http://ebidlocal.cirelli.org/cgi-bin/mmlist.cgi?staples556/270">270</a><br/><div class="morepics"><a href="http://ebidlocal.cirelli.org/cgi-bin/mmlist.cgi?staples556/270">more<br/>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples556/270"><img border="0" src="https://ebid15.com/photos/202104-D1 Warehouse-1356/smalls/270.JPG" alt="smalls/270.JPG"/></a></td><td class="description"><b>Category</b>: DECORATIVE<br/><b>Item</b>: PROP TV MADE IN BRAZIL, METAL 48X28&#34;.  MISSING ONE BUTTON.<br/><b>Location</b>: FRONT RIGHT<br/></td>
					<td align="right" class="bids"><a href="/cgi-bin/mmhistory.cgi?staples556/270"><span id="270_bids">2</span></a></td>
					<td align="right" class="highbidder"><span id="270_highbidder">21493</span></td>
					<td align="right" class="currentamount"><span id="270_currentprice">1.49</span></td>
					<td align="right" class="nextbidrequired"><span id="270_nextrequired"><a href="javascript:subfillform(&#39;270&#39;,&#39;1.99&#39;)">1.99</a></span></td>
					<td align="center" class="yourbid"><span id="270_yourbid"><input type="text" name="270" size="8" placeholder="your bid"/></span></td>
					<td align="center" class="yourmaximum"><span id="270_yourmax"><input type="text" name="m270" size="8" placeholder="your max"/> <br/><i><a href="javascript:subbnpw()">submit bid</a></i></span><br/><span id="270_endtime"></span><br/><span id="270_status"></span></td>
				</tr>
				<tr class="DataRow" id="1435" valign="top">
					<td class="item"><a href="http://ebidlocal.cirelli.org/cgi-bin/mmlist.cgi?staples571/1435">1435</a><br/><div class="morepics"><a href="http://ebidlocal.cirelli.org/cgi-bin/mmlist.cgi?staples571/1435">more<br/>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples571/1435"><img border="0" src="https://ebid15.com/photos/202104-HM-Kelleher-1371/smalls/1435.JPG" alt="smalls/1435.JPG"/></a></td><td class="description"><b>Category</b>: HOUSEHOLD<br/><b>Item</b>: SHOEHORNS, DRESSER VALET, ELECTRIC SHAVERS, TIMEX ALARM CLOCK, SHOESHINE BRUSHES, SONY TV WEATHER/AM/FM PERSONAL RADIO<br/><b>Location</b>: DINING ROOM<br/></td>
					<td align="right" class="bids"><span id="1435_bids"></span></td>
					<td align="right" class="highbidder"><span id="1435_highbidder"></span></td>
					<td align="right" class="currentamount"><span id="1435_currentprice"></span></td>
					<td align="right" class="nextbidrequired"><span id="1435_nextrequired"><a href="javascript:subfillform(&#39;1435&#39;,&#39;0.99&#39;)">0.99</a></span></td>
					<td align="center" class="yourbid"><span id="1435_yourbid"><input type="text" name="1435" size="8" placeholder="your bid"/></span></td>
					<td align="center" class="yourmaximum"><span id="1435_yourmax"><input type="text" name="m1435" size="8" placeholder="your max"/> <br/><i><a href="javascript:subbnpw()">submit bid</a></i></span><br/><span id="1435_endtime"></span><br/><span id="1435_status"></span></td>
				</tr>
			`
	ebid.Client = &mockClient{
		postForm: func(url string, data url.Values) (resp *http.Response, err error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
    <body bgcolor="#FFFFFF">
        <table id="DataTable" class="listbody" border="0" width="100%" align="center" cellpadding="3" cellspacing="1">
            <thead>
                <tr bgcolor="#073C68" valign="bottom">
                <th align="center" width="40"><font color="#FFFFFF"><strong>Item</strong></font></th>
                <th id="DataTablePhoto" align="center"><font color="#FFFFFF"><strong>Photo</strong></font></th>
                <th id="DataTableDesc" align="center"><font color="#FFFFFF"><strong>Description</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Bids</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>High<br>Bidder</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Current<br>Amount</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Next Bid<br>Required</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Your<br>Bid</strong></font></th>
                <th align="center"><font color="#FFFFFF"><strong>Your<br>Maximum</strong></font></th>
                </tr>
            </thead>

            <tbody>
				<tr class="DataRow" id="270" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples556/270">270</a><br><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples556/270">more<br>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples556/270"><img border="0" src="https://ebid15.com/photos/202104-D1 Warehouse-1356/smalls/270.JPG" alt="smalls/270.JPG"></a></td><td class="description"><b>Category</b>: DECORATIVE<br><b>Item</b>: PROP TV MADE IN BRAZIL, METAL 48X28".  MISSING ONE BUTTON.<br><b>Location</b>: FRONT RIGHT<br></td>
					<td align="right" class="bids"><a href="/cgi-bin/mmhistory.cgi?staples556/270"><span id="270_bids">2</span></a></td>
					<td align="right" class="highbidder"><span id="270_highbidder">21493</span></td>
					<td align="right" class="currentamount"><span id="270_currentprice">1.49</span></td>
					<td align="right" class="nextbidrequired"><span id="270_nextrequired"><a href="javascript:subfillform('270','1.99')">1.99</a></span></td>
					<td align="center" class="yourbid"><span id="270_yourbid"><input type="text" name="270" size="8" placeholder="your bid"></span></td>
					<td align="center" class="yourmaximum"><span id="270_yourmax"><input type="text" name="m270" size="8" placeholder="your max"> <br><i><a href="javascript:subbnpw()">submit bid</a></i></span><br><span id="270_endtime"></span><br><span id="270_status"></span></td>
				</tr>
				<tr class="DataRow" id="1435" valign="top">
					<td class="item"><a href="/cgi-bin/mmlist.cgi?staples571/1435">1435</a><br><div class="morepics"><a href="/cgi-bin/mmlist.cgi?staples571/1435">more<br>pics</a></div></td>
					<td align="center" class="photo"><a href="/cgi-bin/mmlist.cgi?staples571/1435"><img border="0" src="https://ebid15.com/photos/202104-HM-Kelleher-1371/smalls/1435.JPG" alt="smalls/1435.JPG"></a></td><td class="description"><b>Category</b>: HOUSEHOLD<br><b>Item</b>: SHOEHORNS, DRESSER VALET, ELECTRIC SHAVERS, TIMEX ALARM CLOCK, SHOESHINE BRUSHES, SONY TV WEATHER/AM/FM PERSONAL RADIO<br><b>Location</b>: DINING ROOM<br></td>
					<td align="right" class="bids"><span id="1435_bids"></span></td>
					<td align="right" class="highbidder"><span id="1435_highbidder"></span></td>
					<td align="right" class="currentamount"><span id="1435_currentprice"></span></td>
					<td align="right" class="nextbidrequired"><span id="1435_nextrequired"><a href="javascript:subfillform('1435','0.99')">0.99</a></span></td>
					<td align="center" class="yourbid"><span id="1435_yourbid"><input type="text" name="1435" size="8" placeholder="your bid"></span></td>
					<td align="center" class="yourmaximum"><span id="1435_yourmax"><input type="text" name="m1435" size="8" placeholder="your max"> <br><i><a href="javascript:subbnpw()">submit bid</a></i></span><br><span id="1435_endtime"></span><br><span id="1435_status"></span></td>
				</tr>
			</tbody>
        </table>
	</body>
</html>
`)),
				StatusCode: 200,
				Request:    &http.Request{},
			}, nil
		},
		get: func(url string) (resp *http.Response, err error) {
			return &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("No value")),
				StatusCode: 200,
				Request:    &http.Request{},
			}, nil
		},
	}
	result, err := SearchAuction("auction1", []string{"hi", "there"})
	assert.Nil(t, err)
	assert.Equalf(t, expected, result, "%s", result)
}
