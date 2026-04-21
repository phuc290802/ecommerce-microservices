package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// PasswordUtils handles password-related operations
type PasswordUtils struct{}

// HashPassword hashes a password using bcrypt
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordUtils) VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// TokenUtils handles token-related operations
type TokenUtils struct {
	jwtSecret   string
	redisClient *redis.Client
}

// NewTokenUtils creates a new TokenUtils instance
func NewTokenUtils(jwtSecret string, redisClient *redis.Client) *TokenUtils {
	return &TokenUtils{jwtSecret: jwtSecret, redisClient: redisClient}
}

// CreateAccessToken creates a new JWT access token
func (t *TokenUtils) CreateAccessToken(user *User) (string, error) {
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: user.Username,
		Email:    user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.jwtSecret))
}

// CreateRefreshToken creates a new refresh token stored in Redis
func (t *TokenUtils) CreateRefreshToken(userID int64, ttl time.Duration) (string, error) {
	token := generateRandomToken()
	key := fmt.Sprintf("refresh:%s", token)
	if err := t.redisClient.Set(context.Background(), key, fmt.Sprint(userID), ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (t *TokenUtils) ValidateRefreshToken(token string) (int64, error) {
	key := fmt.Sprintf("refresh:%s", token)
	userIDStr, err := t.redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userIDStr, 10, 64)
}

// RotateRefreshToken rotates a refresh token
func (t *TokenUtils) RotateRefreshToken(oldToken string, userID int64, ttl time.Duration) (string, error) {
	if err := t.DeleteRefreshToken(oldToken); err != nil {
		return "", err
	}
	return t.CreateRefreshToken(userID, ttl)
}

// DeleteRefreshToken deletes a refresh token
func (t *TokenUtils) DeleteRefreshToken(token string) error {
	return t.redisClient.Del(context.Background(), fmt.Sprintf("refresh:%s", token)).Err()
}

// ValidateToken validates a JWT token and returns its claims
func (t *TokenUtils) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(t.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// OTPUtils handles OTP-related operations
type OTPUtils struct {
	redisClient *redis.Client
}

// NewOTPUtils creates a new OTPUtils instance
func NewOTPUtils(redisClient *redis.Client) *OTPUtils {
	return &OTPUtils{redisClient: redisClient}
}

// GenerateCode generates a 6-digit OTP code
func (o *OTPUtils) GenerateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// StoreOTP stores an OTP code in Redis
func (o *OTPUtils) StoreOTP(email, phone, purpose, code string, ttl time.Duration) error {
	key := formatOTPKey(email, phone, purpose)
	return o.redisClient.Set(context.Background(), key, code, ttl).Err()
}

// ValidateOTP validates an OTP code and deletes it
func (o *OTPUtils) ValidateOTP(email, phone, purpose, code string) (bool, error) {
	key := formatOTPKey(email, phone, purpose)
	stored, err := o.redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	if stored != code {
		return false, nil
	}
	_ = o.redisClient.Del(context.Background(), key).Err()
	return true, nil
}

// ResetTokenUtils handles password reset token operations
type ResetTokenUtils struct {
	redisClient *redis.Client
}

// NewResetTokenUtils creates a new ResetTokenUtils instance
func NewResetTokenUtils(redisClient *redis.Client) *ResetTokenUtils {
	return &ResetTokenUtils{redisClient: redisClient}
}

// CreateResetToken creates a password reset token
func (r *ResetTokenUtils) CreateResetToken(userID int64, ttl time.Duration) (string, error) {
	token := generateRandomToken()
	key := fmt.Sprintf("reset:%s", token)
	if err := r.redisClient.Set(context.Background(), key, fmt.Sprint(userID), ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ValidateResetToken validates a reset token and returns the user ID
func (r *ResetTokenUtils) ValidateResetToken(token string) (int64, error) {
	key := fmt.Sprintf("reset:%s", token)
	userIDStr, err := r.redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteResetToken deletes a reset token
func (r *ResetTokenUtils) DeleteResetToken(token string) error {
	return r.redisClient.Del(context.Background(), fmt.Sprintf("reset:%s", token)).Err()
}

// Utility functions

// generateRandomToken generates a random base64 token
func generateRandomToken() string {
	raw := make([]byte, 32)
	_, err := rand.Read(raw)
	if err != nil {
		return fmt.Sprintf("rt-%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// formatOTPKey formats an OTP storage key
func formatOTPKey(email, phone, purpose string) string {
	identifier := email
	if identifier == "" {
		identifier = phone
	}
	return fmt.Sprintf("otp:%s:%s", purpose, identifier)
}
