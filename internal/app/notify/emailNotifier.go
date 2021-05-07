package notify

import (
	"errors"
	"fmt"
	"strings"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/notify/email"
)

type EmailNotify struct {
	Logger      log.Logger
	MessageChan <-chan NotificationMessage
	ServerUrl   string
}

func (en *EmailNotify) Notify(message NotificationMessage) error {
	en.Logger.Infof("Message received trying to email... %s", message)
	for wlname, wl := range message.User.Watchlists {
		for _, wlID := range strings.Split(wl, ",") {
			if wlID == message.WatchlistID {
				return email.NewEmail(
					[]string{message.User.Email},
					fmt.Sprintf("Your watch list has updates '%s'", wlname),
					fmt.Sprintf("%s/user/%s/watchlist/%s", en.ServerUrl, message.User.ID, wlID),
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
