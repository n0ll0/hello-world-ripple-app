package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/model"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/repository"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

func NewUserService(db *repository.DB) *UserService {
	return &UserService{db: db}
}

type UserService struct {
	db *repository.DB
}

func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{Username: username, PasswordHash: string(hashed)}
	if err := s.db.CreateUser(ctx, user); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	users, err := s.db.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}
