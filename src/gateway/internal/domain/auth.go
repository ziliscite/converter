package domain

import "time"

type Auth struct {
	AccessToken string    `json:"access_token"`
	Exp         time.Time `json:"exp"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}
