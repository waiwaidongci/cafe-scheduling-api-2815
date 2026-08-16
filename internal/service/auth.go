package service

import (
	"context"
	"errors"

	"cafe-scheduling-api/internal/model"
	"cafe-scheduling-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type AuthService struct {
	repo *repository.Repository
}

func NewAuthService(repo *repository.Repository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (model.Employee, error) {
	employee, err := s.repo.GetEmployeeByUsername(ctx, username)
	if err != nil || !employee.Active {
		return model.Employee{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(employee.PasswordHash), []byte(password)); err != nil {
		return model.Employee{}, ErrInvalidCredentials
	}
	employee.PasswordHash = ""
	return employee, nil
}
