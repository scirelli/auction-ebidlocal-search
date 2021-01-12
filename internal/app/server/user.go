package server

import "fmt"

//User user data.
type User struct {
	UserName string `json:"username"`
}

func (u User) String() string {
	return fmt.Sprintf("User name: '%s'", u.UserName)
}

//IsValid validate user data.
func (u User) IsValid() bool {
	return u.UserName != ""
}
