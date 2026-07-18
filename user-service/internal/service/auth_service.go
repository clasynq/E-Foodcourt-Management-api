package service

import (
	"errors"
	"os"
	"strings"
	"time"

	"user-service/internal/model"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService outlines the sign-up and login logic requirements.
type AuthService interface {
	Register(req model.SignupRequest) (*model.User, error)
	Authenticate(req model.LoginRequest) (string, *model.User, error)
}

type authService struct {
	repo repository.UserRepository
}

// NewAuthService returns a new instance of AuthService.
func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

// Register hashes password, sanitizes inputs, and creates a user record.
func (s *authService) Register(req model.SignupRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role != "CUSTOMER" && role != "VENDOR" && role != "ADMIN" {
		role = "CUSTOMER"
	}

	user := &model.User{
		Name:     req.Name,
		Username: strings.ToLower(strings.TrimSpace(req.Username)),
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:    strings.TrimSpace(req.Phone),
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate verifies password and issues a JWT token.
func (s *authService) Authenticate(req model.LoginRequest) (string, *model.User, error) {
	user, err := s.repo.FindByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// generateToken cryptographically signs user claims to return a JWT string.
func (s *authService) generateToken(user *model.User) (string, error) {
	secretKey := os.Getenv("SECRECT_KEY")
	if secretKey == "" {
		secretKey = "default_fallback_jwt_secret"
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &model.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
