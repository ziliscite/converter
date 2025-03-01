package domain

import (
	"errors"
	"github.com/ziliscite/video-to-mp3/auth/pkg/validator"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	emailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type User struct {
	ID        int64
	Username  string
	Email     string
	Password  password
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
	IsAdmin   bool
}

func RegisterUser(username, email string) *User {
	return &User{
		Username: username,
		Email:    email,
	}
}

func (u *User) Hash() []byte {
	return u.Password.hash
}

func (u *User) Validate(v *validator.Validator) {
	if u.Username == "" {
		v.AddError("username", "must be provided")
	}

	if len(u.Username) >= 500 {
		v.AddError("username", "must not be more than 500 characters")
	}

	validateEmail(v, u.Email)

	pw := u.Password.plaintext
	if pw != nil {
		validatePasswordText(v, *pw)
	}

	if u.Hash() == nil {
		panic("missing password hash for user")
	}
}

func validateEmail(v *validator.Validator, email string) {
	if email == "" {
		v.AddError("email", "email is required")
	}

	if !emailRX.MatchString(email) {
		v.AddError("email", "invalid email")
	}
}

func validatePasswordText(v *validator.Validator, password string) {
	if password == "" {
		v.AddError("password", "password is required")
	}

	if len(password) < 8 {
		v.AddError("password", "password must be at least 8 characters")
	}

	if len(password) > 64 {
		v.AddError("password", "password must be less than 64 characters")
	}
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) InsertHash(h []byte) {
	p.hash = h
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
