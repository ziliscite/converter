package token

import (
	"github.com/ziliscite/video-to-mp3/auth/internal/domain"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndValidate_Success(t *testing.T) {
	secret := "test-secret"
	user := &domain.User{
		ID:       int64(123),
		Username: "someone",
		Email:    "user@test.com",
		IsAdmin:  true,
	}

	tokenStr, expAt, err := Create(user, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	now := time.Now()
	expectedExp := now.Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExp, expAt, time.Second)

	parsedUser, err := Validate(tokenStr, secret)
	require.NoError(t, err)
	assert.Equal(t, user.ID, parsedUser.ID)
	assert.Equal(t, user.Email, parsedUser.Email)
	assert.Equal(t, true, parsedUser.IsAdmin)
}

func TestValidate_InvalidSecret(t *testing.T) {
	secret := "correct-secret"
	user := &domain.User{
		ID:       int64(123),
		Username: "someone",
		Email:    "user@test.com",
		IsAdmin:  false,
	}

	tokenStr, _, err := Create(user, secret)
	require.NoError(t, err)

	_, err = Validate(tokenStr, "wrong-secret")
	require.Error(t, err)
	assert.ErrorContains(t, err, "signature is invalid")
}

func TestValidate_ExpiredToken(t *testing.T) {
	secret := "secret"
	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Subject:   "123",
		},
		Id:       123,
		Username: "akasd",
		Email:    "expired@test.com",
		IsAdmin:  true,
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = Validate(tokenStr, secret)
	require.Error(t, err)
	assert.ErrorContains(t, err, "token is expired")
}

func TestValidate_TamperedToken(t *testing.T) {
	secret := "secret"
	user := &domain.User{
		ID:       int64(123),
		Username: "someone",
		Email:    "user@test.com",
		IsAdmin:  false,
	}

	tokenStr, _, err := Create(user, secret)
	require.NoError(t, err)

	parts := strings.Split(tokenStr, ".")
	require.Len(t, parts, 3)
	tamperedToken := parts[0] + "." + parts[1] + "x" + "." + parts[2]

	_, err = Validate(tamperedToken, secret)
	require.Error(t, err)
	assert.ErrorContains(t, err, "token is malformed")
}

func TestCreate_EmptyFields(t *testing.T) {
	secret := "secret"

	user := &domain.User{
		ID:       0,
		Username: "",
		Email:    "",
		IsAdmin:  false,
	}

	tokenStr, _, err := Create(user, secret)
	require.NoError(t, err)

	parsedUser, err := Validate(tokenStr, secret)
	require.NoError(t, err)
	assert.Equal(t, user.ID, parsedUser.ID)
	assert.Equal(t, user.Email, parsedUser.Email)
	assert.Equal(t, user.IsAdmin, parsedUser.IsAdmin)
}
