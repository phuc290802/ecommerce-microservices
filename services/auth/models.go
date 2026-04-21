package main

import "github.com/golang-jwt/jwt/v5"

// User represents a user in the system
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Email    string `json:"email"`
}

// RegisterRequest represents user registration payload
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// LoginRequest represents user login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// OTPRequest represents OTP request payload
type OTPRequest struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Purpose string `json:"purpose"`
}

// OTPValidateRequest represents OTP validation payload
type OTPValidateRequest struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Purpose string `json:"purpose"`
	Code    string `json:"code"`
}

// ForgotPasswordRequest represents forgot password payload
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents reset password payload
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// VerifyTokenRequest represents token verification payload
type VerifyTokenRequest struct {
	Token string `json:"token"`
}
