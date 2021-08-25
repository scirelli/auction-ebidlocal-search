package model

import (
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/generator"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

var defaultIDGen = generator.NewAutoInc(0, 1)

func NewUser(username string) User {
	id, _ := generateID()
	return User{
		Name:       username,
		ID:         id,
		Watchlists: make(map[string]string),
	}
}

//User user data.
type User struct {
	Name string `json:"name"`
	ID   string `json:"id"`

	Email        string    `json:"email"`
	Verified     bool      `json:"verified"`
	LastVerified time.Time `json:"lastVerified"`
	VerifyToken  uuid.UUID `json:"verifyToken"`
	UserDir      string    `json:"-"`
	IsAdmin      bool      `json:"isAdmin"`

	//Wathclist names to ids
	Watchlists map[string]string `json:"watchlists"`
}

func (u User) String() string {
	return fmt.Sprintf("'%s' (%s) %s %s", u.Name, u.ID, u.UserDir, u.Email)
}

//IsValid validate user data.
func (u User) IsValid() bool {
	return u.Name != "" && u.Email != ""
}

func generateID() (string, error) {
	var userID = uuid.New()
	bytes, err := userID.MarshalBinary()
	if err != nil {
		log.New("user", log.DEFAULT_LOG_LEVEL).Error(err)
		id := strconv.Itoa(defaultIDGen.NewID())
		return id, err
	}

	return b64.URLEncoding.EncodeToString(bytes), nil
}
