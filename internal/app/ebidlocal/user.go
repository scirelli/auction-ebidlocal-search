package ebidlocal

import (
	b64 "encoding/base64"
	"strconv"

	"github.com/google/uuid"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/generator"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

var defaultIDGen = generator.NewAutoInc(0, 1)

type User struct {
	Username   string              `json:"userName"`
	ID         string              `json:"userID"`
	UserDir    string              `json:"userDir"`
	Watchlists map[string][]string `json:"watchlists"`
}

func NewUser(username string) User {
	id, _ := generateID()
	return User{
		Username:   username,
		ID:         id,
		Watchlists: make(map[string][]string),
	}
}

func generateID() (string, error) {
	var userID = uuid.New()
	bytes, err := userID.MarshalBinary()
	if err != nil {
		log.New("user").Error.Println(err)
		id := strconv.Itoa(defaultIDGen.NewID())
		return id, err
	}

	return b64.URLEncoding.EncodeToString(bytes), nil
}
