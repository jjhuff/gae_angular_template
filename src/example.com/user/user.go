package user

import (
	"errors"
	"strconv"
)

var (
	ErrEmailExists = errors.New("The email is taken")
	ErrUnknownUser = errors.New("Unknown user")
)

type UserId int64

const (
	InvalidUser UserId = 0
)

func (userId UserId) String() string {
	return strconv.FormatInt(int64(userId), 10)
}
func ParseUserId(s string) (UserId, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return UserId(v), nil
}

type User struct {
	Id           UserId `json:"id" datastore:"-"`
	Email        string `json:"email"`
	EmailOptin   bool   `json:"email_optin" datastore:",noindex"`
	PasswordHash []byte `json:"-" datastore:",noindex"`
}

type DB interface {
	Create(u *User) error
	GetByEmail(email string) (*User, error)
	GetById(id UserId) (*User, error)
	Update(u *User) error
}
