package store

import (
	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
)

//Storer interface to represent a data store.
type Storer interface {
	UserStorer
}

//UserStorer store able to perform user store operations.
type UserStorer interface {
	SaveUser(u *model.User) error
	LoadUser(userID string) (*model.User, error)
	DeleteUser(userID string) error
}
