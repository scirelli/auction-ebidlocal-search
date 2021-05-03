package notify

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/store"
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/store/fs"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

type WatchlistIDToNotificationMessageConverter interface {
	Convert(<-chan string) <-chan NotificationMessage
}

func NewWatchlistConvertData(config Config) *WatchlistConvertData {
	DefaultConfig(&config)

	return &WatchlistConvertData{
		logger: log.New("WatchlistConvertData"),
		config: config,
	}
}

type WatchlistConvertData struct {
	logger *log.Logger
	config Config
}

func (e *WatchlistConvertData) Convert(watchlistIDChan <-chan string) <-chan NotificationMessage {
	var messageChan = make(chan NotificationMessage)

	go func() {
		var watchlistToUsers map[string][]*model.User = e.createUserCache()
		ticker := time.NewTicker(350 * time.Second)

		for wlid := range watchlistIDChan {
			select {
			case <-ticker.C:
				e.logger.Info.Println("Updating user list")
				watchlistToUsers = e.createUserCache()
			default:
			}

			for _, user := range watchlistToUsers[wlid] {
				e.logger.Info.Printf("Sending notification message: %s, about watch list '%s'", user, wlid)
				messageChan <- NotificationMessage{
					User:        user,
					WatchlistID: wlid,
				}
			}
		}
		close(messageChan)
		ticker.Stop()
	}()

	return messageChan
}

func (e *WatchlistConvertData) findAllUsersDataFiles() []string {
	e.logger.Info.Printf("Searching for users")
	var userPaths []string

	if err := filepath.Walk(e.config.UserDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			e.logger.Info.Printf("Failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.Name() == e.config.DataFileName {
			e.logger.Info.Printf("Found file: %q\n", path)
			userPaths = append(userPaths, path)
		}

		return nil
	}); err != nil {
		e.logger.Error.Printf("Error walking the path %q: %v\n", e.config.UserDir, err)
	}

	return userPaths
}

func (e *WatchlistConvertData) allUsers() (userIds []string) {
	var files []os.DirEntry
	var err error

	e.logger.Info.Printf("Searching for users")
	if files, err = os.ReadDir(e.config.UserDir); err != nil {
		e.logger.Error.Printf("Error walking the path %q: %v\n", e.config.UserDir, err)
		return userIds
	}
	for _, file := range files {
		if file.IsDir() {
			userIds = append(userIds, file.Name())
		}
	}

	return userIds
}

func (e *WatchlistConvertData) createUserCache() map[string][]*model.User {
	var userStore store.UserStorer = fs.NewUserStore(e.config.UserDir, e.config.DataFileName, e.logger)
	var watchlistToUsers = make(map[string][]*model.User)

	for _, userID := range e.allUsers() {
		user, err := userStore.LoadUser(context.Background(), userID)
		if err != nil {
			e.logger.Warn.Printf("Skipping user '%s'", userID)
			continue
		}
		for _, listIDs := range user.Watchlists {
			for _, listID := range strings.Split(listIDs, ",") {
				watchlistToUsers[listID] = append(watchlistToUsers[listID], user)
			}
		}
	}

	return watchlistToUsers
}
