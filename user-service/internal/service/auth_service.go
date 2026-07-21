package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"user-service/internal/model"
	"user-service/internal/repository"
	"user-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// CachedUser represents internal user data stored in Redis (including hashed password)
type CachedUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthService outlines the sign-up and login logic requirements.
type AuthService interface {
	Register(req model.SignupRequest) (*model.User, error)
	Authenticate(req model.LoginRequest) (string, *model.User, error)
	BlacklistToken(tokenString string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
	SendOTP(email string) error
	VerifyOTP(email string, otp string) (bool, error)
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

// Register hashes password, sanitizes inputs, creates a user record, and caches in Redis.
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

	// Write-Through: Cache user in Redis immediately upon registration
	s.cacheUserInRedis(user)

	return user, nil
}

// Authenticate verifies password and issues a JWT token (checking Redis cache first).
func (s *authService) Authenticate(req model.LoginRequest) (string, *model.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var user *model.User

	// 1. Read-Through: Check Redis Cache first to bypass PostgreSQL query
	cachedUser, err := s.getUserFromRedis(email)
	if err == nil && cachedUser != nil {
		log.Println("[Cache HIT] User details retrieved from Redis for login:", email)
		user = cachedUser
	} else {
		log.Println("[Cache MISS] Fetching user details from PostgreSQL database for login:", email)
		user, err = s.repo.FindByEmail(email)
		if err != nil {
			return "", nil, errors.New("invalid email or password")
		}
		// Cache retrieved user in Redis for subsequent fast logins
		s.cacheUserInRedis(user)
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

// Helper: Cache user details in Redis under user:email:<email> and user:id:<id>
func (s *authService) cacheUserInRedis(user *model.User) {
	ctx := context.Background()
	cached := CachedUser{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Password:  user.Password,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		log.Printf("Warning: Failed to marshal user for Redis cache: %v", err)
		return
	}

	emailKey := "user:email:" + strings.ToLower(user.Email)
	idKey := "user:id:" + user.ID
	ttl := 7 * 24 * time.Hour // Cache for 7 days

	if err := s.rdb.Set(ctx, emailKey, data, ttl).Err(); err != nil {
		log.Printf("Warning: Failed to cache user by email in Redis: %v", err)
	}
	if err := s.rdb.Set(ctx, idKey, data, ttl).Err(); err != nil {
		log.Printf("Warning: Failed to cache user by ID in Redis: %v", err)
	}
}

// Helper: Retrieve cached user details from Redis
func (s *authService) getUserFromRedis(email string) (*model.User, error) {
	ctx := context.Background()
	key := "user:email:" + strings.ToLower(email)

	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cached CachedUser
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, err
	}

	user := &model.User{
		ID:        cached.ID,
		Name:      cached.Name,
		Username:  cached.Username,
		Email:     cached.Email,
		Phone:     cached.Phone,
		Password:  cached.Password,
		Role:      cached.Role,
		CreatedAt: cached.CreatedAt,
	}

	return user, nil
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

// SendOtp generated a 6-digit numeric code, saves it to redis (5 min TTL), and sends via Mail
func (s *authService) SendOTP(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Generates cryptographically secre 6-digit random number
	nBig, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return fmt.Errorf("failed to generate random otp")
	}
	otpCode := fmt.Sprintf("%06d", nBig.Int64()+100000)

	// save the otp with the 5 minutes expiration time
	ctx := context.Background()
	key := "otp:email:" + email
	err = s.rdb.Set(ctx, key, otpCode, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to save the otp in the redis: %v", err)
	}

	// Dispatch the otp via SMTP
	return utils.SendOTPEmail(email, otpCode)
}

func (s *authService) VerifyOTP(email string, otp string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	ctx := context.Background()
	key := "otp:email:" + email

	// Fetch stored otp from redis
	storedOTP, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, errors.New("OTP has expired or as not requested")
		}
		return false, err
	}
	if storedOTP != otp {
		return false, errors.New("Invalid OTP code")
	}

	// Deleted OTP key from Redis upon successfully veriication so it cannot be reused

	s.rdb.Del(ctx, key)
	return true, nil
}
