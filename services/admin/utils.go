package main

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
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

// TokenUtils handles JWT operations
type TokenUtils struct {
	jwtSecret string
}

// NewTokenUtils creates a new TokenUtils instance
func NewTokenUtils(jwtSecret string) *TokenUtils {
	return &TokenUtils{jwtSecret: jwtSecret}
}

// CreateToken creates a new JWT token for admin
func (t *TokenUtils) CreateToken(adminID int64, role string, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl)
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", adminID),
		"role": role,
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.jwtSecret))
}

// ValidateToken validates a JWT token and returns claims
func (t *TokenUtils) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractAdminIDFromToken extracts admin ID from token claims
func ExtractAdminIDFromToken(claims jwt.MapClaims) (int64, error) {
	sub := claims["sub"]
	if sub == nil {
		return 0, fmt.Errorf("subject claim not found")
	}

	subStr, ok := sub.(string)
	if !ok {
		return 0, fmt.Errorf("invalid subject format")
	}

	var adminID int64
	_, err := fmt.Sscanf(subStr, "%d", &adminID)
	return adminID, err
}

// ExtractRoleFromToken extracts role from token claims
func ExtractRoleFromToken(claims jwt.MapClaims) (string, error) {
	role := claims["role"]
	if role == nil {
		return "", fmt.Errorf("role claim not found")
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", fmt.Errorf("invalid role format")
	}

	return roleStr, nil
}
