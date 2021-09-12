package notify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/notify/email"
)

var cssValidCharacters, startWithNumbers *regexp.Regexp

func init() {
	cssValidCharacters = regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	startWithNumbers = regexp.MustCompile(`^[-0-9]+`)
}

type EmailNotify struct {
	Logger      log.Logger
	MessageChan <-chan NotificationMessage
	template    *template.Template
	store       store.WatchlistContentStorer
	config      Config
}

func NewEmailNotify(config Config, store store.WatchlistContentStorer, messageChan <-chan NotificationMessage) *EmailNotify {
	var logger = log.New("EmailNotify", log.DEFAULT_LOG_LEVEL)

	t, err := template.New("email.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
		"String": func(item fmt.Stringer) string {
			return item.String()
		},
	}).ParseFiles(config.TemplateFile)
	if err != nil {
		logger.Fatal(err)
	}

	return &EmailNotify{
		MessageChan: messageChan,
		Logger:      logger,
		config:      config,
		template:    t,
		store:       store,
	}
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

	var wllink string
	var err error
	var content *model.WatchlistContent
	for wlname, wl := range message.User.Watchlists {
		for _, wlID := range strings.Split(wl, ",") {
			var emailBody *bytes.Buffer = &bytes.Buffer{}
			wllink = fmt.Sprintf("%s/viewwatchlists.html?id=%s#%s", en.config.ServerUrl, url.QueryEscape(message.User.ID), startWithNumbers.ReplaceAllString(cssValidCharacters.ReplaceAllString(wlname+"_"+wlID, ""), ""))

			if content, err = en.store.LoadWatchlistContent(context.Background(), wlID); err != nil {
				en.Logger.Errorf("Error retrieving wl '%s'", err)
				en.Logger.Debugf("wlname: '%s'; wl: '%s'; wllink '%s'; error '%s'", wlname, wl, wllink, err)
				content = &model.WatchlistContent{}
			}

			if err := en.template.Execute(emailBody, struct {
				ServerURL     string
				Rows          []model.AuctionItem
				WatchlistLink string
				WatchlistName string
				Timestamp     string
				TimestampEpoc string
			}{
				ServerURL:     en.config.ServerUrl,
				Rows:          content.AuctionItems,
				WatchlistLink: wllink,
				WatchlistName: wlname,
				Timestamp:     time.Now().Format(time.UnixDate),
				TimestampEpoc: fmt.Sprintf("%d", time.Now().Unix()),
			}); err != nil {
				en.Logger.Errorf("Error retrieving wl '%s'", err)
				return err
			}

			if wlID == message.WatchlistID {
				eb := emailBody.String()
				if os.Getenv("DEBUG") != "" {
					f, _ := ioutil.TempFile("/tmp", fmt.Sprintf("doc_%s_", "EmailBody"))
					f.WriteString(eb)
					f.Close()
				}
				return email.NewEmail(
					[]string{message.User.Email},
					fmt.Sprintf("Your watch list has updates '%s'", wlname),
					eb,
				).Send()
			}
		}
	}
	return errors.New("Failed to notify user watch list not found among user's watch lists.")
}
