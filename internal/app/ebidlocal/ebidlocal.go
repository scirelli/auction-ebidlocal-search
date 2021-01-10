package ebidlocal

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

func New(config Config) *Ebidlocal {
	if config.DataFileName == "" {
		config.DataFileName = "data.json"
	}

	return &Ebidlocal{
		config: config,
		logger: log.New("Ebidlocal"),
	}
}

type Ebidlocal struct {
	config Config
	logger *log.Logger
}

func (e *Ebidlocal) CreateUser(username string) (string, error) {
	var userID = uuid.New()
	bytes, err := userID.MarshalBinary()
	if err != nil {
		return "", err
	}

	userIDEnc := b64.StdEncoding.EncodeToString(bytes)
	e.createUserSpace(user{
		username: username,
		userID:   userIDEnc,
	})

	return userIDEnc, nil
}

func (e *Ebidlocal) createUserSpace(u user) {
	u.userDir = filepath.Join(e.config.UserDir, u.userID)
	os.MkdirAll(filepath.Join(u.userDir, "watchlists"), 0755)

	file, _ := json.MarshalIndent(u, "", "    ")
	ioutil.WriteFile(filepath.Join(u.userDir, e.config.DataFileName), file, 0644)

	os.Symlink(filepath.Join("./", "template"), filepath.Join(u.userDir, "web"))
}

type user struct {
	username string `json:"userName"`
	userID   string `json:"userID"`
	userDir  string `json:"userDir"`
}
