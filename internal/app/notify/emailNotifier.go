package notify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/notify/email"
)

type EmailNotify struct {
	Logger       log.Logger
	MessageChan  <-chan NotificationMessage
	ServerUrl    string
	WatchlistDir string
}

func (en *EmailNotify) Notify(message NotificationMessage) error {
	en.Logger.Infof("Message received trying to email... %s", message)
	var body string
	for wlname, wl := range message.User.Watchlists {
		for _, wlID := range strings.Split(wl, ",") {
			body = fmt.Sprintf("%s/user/%s/watchlist/%s", en.ServerUrl, message.User.ID, wlID)
			if data, err := ioutil.ReadFile(path.Join(en.WatchlistDir, wlID, "index.html")); err == nil {
				body = strings.Replace(string(data), "&lt;!--{{watchlistName}}--&gt;", wlname, 1)
			} else {
				en.Logger.Errorf("Error retrieving wl '%s'", err)
			}

			if wlID == message.WatchlistID {
				return email.NewEmail(
					[]string{message.User.Email},
					fmt.Sprintf("Your watch list has updates '%s'", wlname),
					body,
				).Send()
			}
		}
	}
	return errors.New("Failed to notify user watch list not found among user's watch lists.")
}

func (en *EmailNotify) Send() error {
	for message := range en.MessageChan {
		en.Notify(message)
	}
	return nil
}
