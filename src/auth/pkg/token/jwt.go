package token

import (
	"errors"
	"fmt"
	"github.com/ziliscite/video-to-mp3/auth/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

func Create(user *domain.User, secretKey string) (string, time.Time, error) {
	now := time.Now()
	expAt := now.Add(time.Hour * time.Duration(24))

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-service",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
		Id:      user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}

	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return "", expAt, errors.New("failed to create token")
	}

	return tokenStr, expAt, nil
}

// Validate will validate token and return user
func Validate(tokenStr, secretKey string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Issuer != "auth-service" {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invalid token")
	}

	v := jwt.NewValidator()
	if err = v.Validate(claims); err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	return &domain.User{
		ID:       claims.Id,
		Username: claims.Username,
		Email:    claims.Email,
		IsAdmin:  claims.IsAdmin,
	}, nil
}
