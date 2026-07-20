package service

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"user-service/internal/model"
	"user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// AuthService outlines the sign-up and login logic requirements.
type AuthService interface {
	Register(req model.SignupRequest) (*model.User, error)
	Authenticate(req model.LoginRequest) (string, *model.User, error)
	BlacklistToken(tokenString string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
}

type authService struct {
	repo repository.UserRepository
	rdb  *redis.Client
}

// IsTokenBlacklisted checks if the token key exists in Redis.
func (s *authService) IsTokenBlacklisted(tokenString string) (bool, error) {
	ctx := context.Background()
	key := "blacklist:" + tokenString
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// NewAuthService returns a new instance of AuthService.
func NewAuthService(repo repository.UserRepository, rdb *redis.Client) AuthService {
	return &authService{repo: repo, rdb: rdb}
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
		return "", nil, errors.New("invalid email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.New("invalid password please forget it if you didn't remember")
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
		log.Fatalf("Configure a secrect key in the .env file")
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

// BlacklistToken extracts the expirations time from the jwt and saves it to Redis.
func (s *authService) BlacklistToken(tokenString string) error {
	secretKey := os.Getenv("SECRECT_KEY")
	if secretKey == "" {
		log.Fatalf("Configure a secretkey in the .env file")
	}

	// Parse token to extract expiration time
	claims := &model.Claims{}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return err
	}

	expTime, err := token.Claims.GetExpirationTime()
	if err != nil || expTime == nil {
		return errors.New("could not retrive token expiration time")
	}

	// calculate remaining time left before token expires
	ttl := time.Until(expTime.Time)
	if ttl <= 0 {
		return nil //Already expired, no need to blacklist
	}

	// Store token in redis with the calculated TTL (time-to-live)
	ctx := context.Background()
	key := "blacklist:" + tokenString
	return s.rdb.Set(ctx, key, "true", ttl).Err()
}
