package notify

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/stretchr/testify/assert"
)

type Test struct {
	Name      string
	Config    Config
	UserCount int64
	Users     []model.User
}

var config = Config{
	ContentPath:  "/tmp/email_test",
	DataFileName: "data.json",
	UserDir:      "/tmp/email_test/web/user",
}

func setupfixture(t *testing.T, config Config, userData model.User) (string, func()) {
	err := os.MkdirAll(config.UserDir, 0755)
	assert.Nil(t, err)
	dir, err := os.MkdirTemp(config.UserDir, "*")
	assert.Nil(t, err)
	userData.ID = filepath.Base(dir)

	file, err := json.Marshal(userData)
	assert.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, config.DataFileName), file, 0644)

	err = ioutil.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
	assert.Nil(t, err)

	return dir, func() {
		defer os.RemoveAll(config.ContentPath)
	}
}
func dataFixture() []Test {
	return []Test{
		{
			Name:   "finds one user",
			Config: config,
			Users: []model.User{
				model.User{
					Name: "fake1",
					ID:   "",
					Watchlists: map[string]string{
						"wlName": "id1",
					},
				},
			},
		},
		{
			Name:   "finds two user",
			Config: config,
			Users: []model.User{
				model.User{
					Name: "fake2",
					ID:   "",
					Watchlists: map[string]string{
						"wlName": "id1,id2,id3",
					},
				},
				model.User{
					Name: "fake3",
					ID:   "",
					Watchlists: map[string]string{
						"wlName": "id4",
					},
				},
			},
		},
	}
}

func Test_findAllUsersDataFiles(t *testing.T) {
	for _, test := range dataFixture() {
		t.Run(test.Name, func(t *testing.T) {
			var e WatchlistConvertData = WatchlistConvertData{
				logger: log.New("email_test"),
				config: config,
			}
			var expected []string

			for _, user := range test.Users {
				dir, tearDown := setupfixture(t, test.Config, user)
				defer tearDown()
				expected = append(expected, filepath.Join(dir, "data.json"))
			}
			actual := e.findAllUsersDataFiles()
			sort.Sort(sort.StringSlice(expected))
			sort.Sort(sort.StringSlice(actual))
			assert.Equalf(t, actual, expected, "Should search user directories")
		})
	}
}

func Test_allUsers(t *testing.T) {
	for _, test := range dataFixture() {
		t.Run(test.Name, func(t *testing.T) {
			var e WatchlistConvertData = WatchlistConvertData{
				logger: log.New("email_test"),
				config: config,
			}
			var expected []string

			for _, user := range test.Users {
				dir, tearDown := setupfixture(t, test.Config, user)
				defer tearDown()
				expected = append(expected, filepath.Base(dir))
			}
			actual := e.allUsers()
			sort.Sort(sort.StringSlice(expected))
			sort.Sort(sort.StringSlice(actual))
			assert.Equalf(t, actual, expected, "Should return a slice of user ids.")
		})
	}
}

func TestConvert(t *testing.T) {
	for _, test := range dataFixture() {
		t.Run(test.Name, func(t *testing.T) {
			var e WatchlistConvertData = WatchlistConvertData{
				logger: log.New("email_test"),
				config: config,
			}

			for _, user := range test.Users {
				_, tearDown := setupfixture(t, test.Config, user)
				defer tearDown()
			}

			var watchlistIDchan = make(chan string)
			var ch <-chan NotificationMessage = e.Convert(watchlistIDchan)
			for _, user := range test.Users {
				for _, id := range strings.Split(user.Watchlists["wlName"], ",") {
					go func(id string) {
						watchlistIDchan <- id
					}(id)
					msg := <-ch
					assert.Equalf(t, msg.User.Name, user.Name, "should create a notification message for the user.")
					assert.Equalf(t, msg.WatchlistID, id, "should create a notification message for the user with the watch list to notify about.")
				}
			}

			close(watchlistIDchan)
		})
	}
}
