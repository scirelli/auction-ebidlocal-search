package server

import "fmt"

//User user data.
type User struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (u User) String() string {
	return fmt.Sprintf("User name: '%s' (%s)", u.Name, u.ID)
}

//IsValid validate user data.
func (u User) IsValid() bool {
	return u.Name != ""
}
