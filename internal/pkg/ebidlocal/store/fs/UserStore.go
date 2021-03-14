package fs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/log"
)

//NewUserStore constructor for the UserStore
func NewUserStore(baseUserDir string, dataFileName string, log *log.Logger) *UserStore {
	return &UserStore{
		baseUserDir:  baseUserDir,
		dataFileName: dataFileName,
		logger:       log,
	}
}

type UserStore struct {
	baseUserDir  string
	dataFileName string
	logger       *log.Logger
}

func (s *UserStore) SaveUser(ctx context.Context, u *User) (string, error) {
	var userDir string = filepath.Join(s.baseUserDir, u.ID)

	s.logger.Info.Printf("Creating user '%s' at '%s'\n", u.ID, userDir)
	if err := os.MkdirAll(userDir, 0775); err != nil {
		s.logger.Error.Println(err)
		return "", err
	}

	file, err := json.Marshal(u)
	if err != nil {
		s.logger.Error.Println(err)
		return "", err
	}
	return u.ID, ioutil.WriteFile(filepath.Join(userDir, s.dataFileName), file, 0644)
}

func (s *UserStore) LoadUser(ctx context.Context, userID string) (*User, error) {
	var userDataFile string = filepath.Join(s.baseUserDir, userID, s.dataFileName)

	if _, err := os.Stat(userDataFile); os.IsNotExist(err) {
		s.logger.Info.Println("User does not exist")
		return nil, err
	}

	var usr User
	jsonFile, err := os.Open(userDataFile)
	if err != nil {
		s.logger.Error.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&usr); err != nil {
		s.logger.Error.Println(err)
		return nil, err
	}

	return &usr, nil
}

func (s *UserStore) DeleteUser(ctx context.Context, userID string) error {
	return os.RemoveAll(filepath.Join(s.baseUserDir, userID))
}
