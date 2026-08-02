package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"staff-login/internal/model"
	"staff-login/internal/repository"
	"staff-login/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type StaffService interface {
	InitiateLogin(req model.InitiateLoginRequest) error
	VerifyLogin(req model.VerifyLoginRequest) (string, *model.StaffMember, error)
	BlacklistToken(tokenString string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
}

type staffService struct {
	repo repository.StaffRepository
	rdb  *redis.Client
}

func NewStaffService(repo repository.StaffRepository, rdb *redis.Client) StaffService {
	return &staffService{repo: repo, rdb: rdb}
}

func (s *staffService) InitiateLogin(req model.InitiateLoginRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// 1. Find staff member in DB
	member, err := s.repo.FindByEmail(email)
	if err != nil {
		return errors.New("invalid email or password")
	}

	if !member.IsActive {
		return errors.New("your staff account is deactivated. Contact admin.")
	}

	// 2. Validate password
	err = bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password))
	if err != nil {
		return errors.New("invalid email or password")
	}

	// 3. Generate secure 6-digit OTP
	nBig, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return fmt.Errorf("failed to generate secure security code")
	}
	otpCode := fmt.Sprintf("%06d", nBig.Int64()+100000)

	// Log OTP to server console (so they can verify/test without SMTP setup)
	log.Printf("-------------------------------------------")
	log.Printf("STAFF 2FA LOGIN: User=%s, Code=%s, Role=%s", email, otpCode, member.Role)
	log.Printf("-------------------------------------------")

	// 4. Save OTP in Redis with 5 minutes TTL
	ctx := context.Background()
	key := "staff:otp:" + email
	err = s.rdb.Set(ctx, key, otpCode, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to save verification code in cache: %v", err)
	}

	// 5. Send OTP via SMTP asynchronously
	go func(recipient, code, role string) {
		if err := utils.SendStaffOTPEmail(recipient, code, role); err != nil {
			log.Printf("Error: Failed to deliver background staff 2FA email to %s: %v", recipient, err)
		} else {
			log.Printf("Success: Background staff 2FA email delivered to %s", recipient)
		}
	}(email, otpCode, member.Role)

	return nil
}

func (s *staffService) VerifyLogin(req model.VerifyLoginRequest) (string, *model.StaffMember, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	ctx := context.Background()
	key := "staff:otp:" + email

	// 1. Fetch OTP from Redis
	storedOTP, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil, errors.New("verification code has expired or was not requested")
		}
		return "", nil, err
	}

	// 2. Validate OTP
	if storedOTP != req.OTP {
		return "", nil, errors.New("invalid verification code")
	}

	// Delete OTP from Redis so it is a single-use token
	s.rdb.Del(ctx, key)

	// 3. Get staff member details
	member, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}

	// 4. Generate signed JWT token
	token, err := s.generateToken(member)
	if err != nil {
		return "", nil, err
	}

	return token, member, nil
}

func (s *staffService) BlacklistToken(tokenString string) error {
	secretKey := os.Getenv("SECRECT_KEY")
	if secretKey == "" {
		log.Fatalf("Configure a SECRECT_KEY in the .env file")
	}

	claims := &model.Claims{}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return err
	}

	expTime, err := token.Claims.GetExpirationTime()
	if err != nil || expTime == nil {
		return errors.New("could not retrieve token expiration time")
	}

	ttl := time.Until(expTime.Time)
	if ttl <= 0 {
		return nil
	}

	ctx := context.Background()
	key := "blacklist:" + tokenString
	return s.rdb.Set(ctx, key, "true", ttl).Err()
}

func (s *staffService) IsTokenBlacklisted(tokenString string) (bool, error) {
	ctx := context.Background()
	key := "blacklist:" + tokenString
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *staffService) generateToken(member *model.StaffMember) (string, error) {
	secretKey := os.Getenv("SECRECT_KEY")
	if secretKey == "" {
		log.Fatalf("Configure a SECRECT_KEY in the .env file")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &model.Claims{
		UserID: member.ID,
		Role:   member.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
