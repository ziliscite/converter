package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/video-to-mp3/auth/internal/domain"
	"github.com/ziliscite/video-to-mp3/auth/internal/repository"
	"github.com/ziliscite/video-to-mp3/auth/pkg/validator"
)

var (
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateMail      = errors.New("email has been taken")
)

type UserService interface {
	SignIn(ctx context.Context, email, password string) (*domain.User, error)
	SignUp(ctx context.Context, v *validator.Validator, username, email, password string) (*domain.User, error)
}

type userServ struct {
	ur repository.UserRepository
}

func NewUserService(ur repository.UserRepository) UserService {
	return &userServ{ur}
}

func (u userServ) SignIn(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := u.ur.GetByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			return nil, ErrInvalidCredentials
		default:
			return nil, fmt.Errorf("something's wrong: %w", err)
		}
	}

	ok, err := user.Password.Matches(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (u userServ) SignUp(ctx context.Context, v *validator.Validator, username, email, password string) (*domain.User, error) {
	user := domain.RegisterUser(username, email)

	if err := user.Password.Set(password); err != nil {
		return nil, err
	}

	if user.Validate(v); !v.Valid() {
		return nil, ErrInvalidUser
	}

	if err := u.ur.Insert(ctx, user); err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEntry):
			return nil, ErrDuplicateMail
		default:
			return nil, fmt.Errorf("something's wrong: %w", err)
		}
	}

	return user, nil
}
