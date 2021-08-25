package notify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/notify/email"
)

var cssValidCharacters, startWithNumbers *regexp.Regexp

func init() {
	cssValidCharacters = regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	startWithNumbers = regexp.MustCompile(`^[-0-9]+`)
}

type EmailNotify struct {
	Logger       log.Logger
	MessageChan  <-chan NotificationMessage
	ServerUrl    string
	WatchlistDir string
}

func (en *EmailNotify) Send() error {
	for message := range en.MessageChan {
		if err := en.Notify(message); err != nil {
			en.Logger.Errorf("Error Sending email: '%s' for message %s", err, message)
		}
	}
	return nil
}

func (en *EmailNotify) Notify(message NotificationMessage) error {
	en.Logger.Infof("Message received trying to email... %s", message)

	var body, wllink string
	for wlname, wl := range message.User.Watchlists {
		for _, wlID := range strings.Split(wl, ",") {
			//body = fmt.Sprintf("%s/user/%s/watchlist/%s", en.ServerUrl, message.User.ID, wlID)
			wllink = fmt.Sprintf("%s/viewwatchlists.html?id=%s#%s", en.ServerUrl, url.QueryEscape(message.User.ID), startWithNumbers.ReplaceAllString(cssValidCharacters.ReplaceAllString(wlname+"_"+wlID, ""), ""))
			body = wllink
			if data, err := ioutil.ReadFile(path.Join(en.WatchlistDir, wlID, "index.html")); err == nil {
				//en.Logger.Debugf("Email TEMPLATE \n\n%s\n\n", string(data))
				body = strings.Replace(string(data), "<!--{{watchlistName}}-->", wlname, 1)
				body = strings.Replace(body, "__watchlistLink__", wllink, 1)
			} else {
				en.Logger.Errorf("Error retrieving wl '%s'", err)
				en.Logger.Debugf("wlname: '%s'; wl: '%s'; data: '%s'; body '%s'; error '%s'", wlname, wl, data, body, err)
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
