package token

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndValidate_Success(t *testing.T) {
	secret := "test-secret"
	id := int64(123)
	email := "user@test.com"

	tokenStr, expAt, err := Create(id, true, email, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	now := time.Now()
	expectedExp := now.Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExp, expAt, time.Second)

	parsedID, parsedEmail, parsedStatus, err := Validate(tokenStr, secret)
	require.NoError(t, err)
	assert.Equal(t, id, parsedID)
	assert.Equal(t, email, parsedEmail)
	assert.Equal(t, true, parsedStatus)
}

func TestValidate_InvalidSecret(t *testing.T) {
	secret := "correct-secret"
	id := int64(123)
	email := "user@test.com"

	tokenStr, _, err := Create(id, true, email, secret)
	require.NoError(t, err)

	_, _, _, err = Validate(tokenStr, "wrong-secret")
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
		Id:     123,
		Email:  "expired@test.com",
		Active: false,
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)

	_, _, status, err := Validate(tokenStr, secret)
	require.Error(t, err)
	assert.Equal(t, false, status)
	assert.ErrorContains(t, err, "token is expired")
}

func TestValidate_TamperedToken(t *testing.T) {
	secret := "secret"
	id := int64(123)
	email := "user@test.com"

	tokenStr, _, err := Create(id, true, email, secret)
	require.NoError(t, err)

	parts := strings.Split(tokenStr, ".")
	require.Len(t, parts, 3)
	tamperedToken := parts[0] + "." + parts[1] + "x" + "." + parts[2]

	_, _, _, err = Validate(tamperedToken, secret)
	require.Error(t, err)
	assert.ErrorContains(t, err, "token is malformed")
}

func TestCreate_EmptyFields(t *testing.T) {
	secret := "secret"
	id := int64(0)
	email := ""

	tokenStr, _, err := Create(id, false, email, secret)
	require.NoError(t, err)

	parsedID, parsedEmail, parsedStatus, err := Validate(tokenStr, secret)
	require.NoError(t, err)
	assert.Equal(t, id, parsedID)
	assert.Equal(t, email, parsedEmail)
	assert.Equal(t, false, parsedStatus)
}
