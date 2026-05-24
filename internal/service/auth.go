package service

import (
	"context"
	"errors"
	"fmt"
	"ssubench/internal/config"
	"ssubench/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   domain.UserRepository
	config *config.Config
}

func NewAuthService(repo domain.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		repo:   repo,
		config: config,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string, role domain.UserRole) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.config.PasswordCost)

	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: user %s: generating password: %w", username, err)
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		Status:       domain.StatusActive,
		Balance:      s.config.StartingBalance,
	}

	err = s.repo.Create(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: user %s: %w", username, err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := s.repo.GetByUsername(ctx, username)

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", fmt.Errorf("AuthService.Login: %w", domain.ErrInvalidCredentials)
		}

		return "", fmt.Errorf("AuthService.Login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("AuthService.Login: %w", domain.ErrInvalidCredentials)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(s.config.JWTTTL).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))

	if err != nil {
		return "", fmt.Errorf("AuthService.Login: %w", err)
	}

	return tokenString, nil
}
