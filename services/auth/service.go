package main

import (
	"fmt"
	"log"
)

// AuthService contains business logic for authentication
type AuthService struct {
	repo       *Repository
	passUtils  *PasswordUtils
	tokenUtils *TokenUtils
	otpUtils   *OTPUtils
	resetUtils *ResetTokenUtils
	config     Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	repo *Repository,
	tokenUtils *TokenUtils,
	otpUtils *OTPUtils,
	resetUtils *ResetTokenUtils,
	config Config,
) *AuthService {
	return &AuthService{
		repo:       repo,
		passUtils:  &PasswordUtils{},
		tokenUtils: tokenUtils,
		otpUtils:   otpUtils,
		resetUtils: resetUtils,
		config:     config,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(req *RegisterRequest) (int64, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return 0, fmt.Errorf("username, email, password required")
	}

	exists, err := s.repo.UserExists(req.Email, req.Username)
	if err != nil {
		return 0, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return 0, fmt.Errorf("username or email already used")
	}

	passwordHash, err := s.passUtils.HashPassword(req.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := s.repo.CreateUser(req.Username, req.Email, req.Phone, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User registered: %s (%s)", req.Username, req.Email)
	return userID, nil
}

// LoginUser authenticates a user and returns tokens
func (s *AuthService) LoginUser(req *LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password required")
	}

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !s.passUtils.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	accessToken, err := s.tokenUtils.CreateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := s.tokenUtils.CreateRefreshToken(user.ID, s.config.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	log.Printf("User logged in: %s", req.Email)
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		Username:     user.Username,
	}, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (s *AuthService) RefreshAccessToken(refreshToken string) (*RefreshResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token required")
	}

	userID, err := s.tokenUtils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	newAccessToken, err := s.tokenUtils.CreateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	newRefreshToken, err := s.tokenUtils.RotateRefreshToken(refreshToken, user.ID, s.config.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	return &RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    900,
	}, nil
}

// LogoutUser invalidates a refresh token
func (s *AuthService) LogoutUser(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.tokenUtils.DeleteRefreshToken(refreshToken)
}

// VerifyToken validates a token and returns its claims
func (s *AuthService) VerifyToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token required")
	}
	return s.tokenUtils.ValidateToken(tokenString)
}

// RequestOTP sends an OTP code
func (s *AuthService) RequestOTP(req *OTPRequest) error {
	if req.Email == "" && req.Phone == "" {
		return fmt.Errorf("email or phone required")
	}

	if req.Purpose == "" {
		req.Purpose = "verification"
	}

	code := s.otpUtils.GenerateCode()
	if err := s.otpUtils.StoreOTP(req.Email, req.Phone, req.Purpose, code, s.config.OTPTokenTTL); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	log.Printf("OTP for %s/%s purpose=%s: %s", req.Email, req.Phone, req.Purpose, code)
	return nil
}

// ValidateOTP validates an OTP code
func (s *AuthService) ValidateOTP(req *OTPValidateRequest) error {
	if req.Email == "" && req.Phone == "" {
		return fmt.Errorf("email or phone required")
	}
	if req.Code == "" {
		return fmt.Errorf("OTP code required")
	}

	valid, err := s.otpUtils.ValidateOTP(req.Email, req.Phone, req.Purpose, req.Code)
	if err != nil {
		return fmt.Errorf("failed to validate OTP: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid OTP code")
	}

	return nil
}

// RequestPasswordReset initiates a password reset
func (s *AuthService) RequestPasswordReset(req *ForgotPasswordRequest) (string, error) {
	if req.Email == "" {
		return "", fmt.Errorf("email required")
	}

	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		return "", fmt.Errorf("email not found")
	}

	token, err := s.resetUtils.CreateResetToken(user.ID, s.config.ResetTokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to create reset token: %w", err)
	}

	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", token)
	log.Printf("Password reset link for %s: %s", user.Email, resetLink)

	return token, nil
}

// ResetPassword resets a user's password
func (s *AuthService) ResetPassword(req *ResetPasswordRequest) error {
	if req.Token == "" || req.Password == "" {
		return fmt.Errorf("token and password required")
	}

	userID, err := s.resetUtils.ValidateResetToken(req.Token)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	passwordHash, err := s.passUtils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(userID, passwordHash); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	_ = s.resetUtils.DeleteResetToken(req.Token)
	log.Printf("Password reset for user ID: %d", userID)

	return nil
}

// Response types

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	Username     string `json:"username"`
}

// RefreshResponse represents the refresh response
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}
