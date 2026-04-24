package main

import (
	"fmt"
	"log"
)

// AdminService contains business logic for admin operations
type AdminService struct {
	repo       *Repository
	passUtils  *PasswordUtils
	tokenUtils *TokenUtils
	config     Config
}

// NewAdminService creates a new AdminService instance
func NewAdminService(repo *Repository, tokenUtils *TokenUtils, config Config) *AdminService {
	return &AdminService{
		repo:       repo,
		passUtils:  &PasswordUtils{},
		tokenUtils: tokenUtils,
		config:     config,
	}
}

// LoginAdmin authenticates an administrator
func (s *AdminService) LoginAdmin(req *LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password required")
	}

	admin, err := s.repo.FindAdminByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if admin.Status != "active" {
		return nil, fmt.Errorf("admin account is locked")
	}

	if !s.passUtils.VerifyPassword(admin.PasswordHash, req.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := s.tokenUtils.CreateToken(admin.ID, string(admin.Role), s.config.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Hide password hash in response
	admin.PasswordHash = ""

	log.Printf("Admin logged in: %s (ID: %d)", admin.Email, admin.ID)
	return &LoginResponse{
		Token: token,
		User:  admin,
	}, nil
}

// ListAdmins retrieves all administrators
func (s *AdminService) ListAdmins() ([]*Administrator, error) {
	admins, err := s.repo.ListAdmins()
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}
	return admins, nil
}

// CreateAdmin creates a new administrator
func (s *AdminService) CreateAdmin(req *CreateAdminRequest) (*Administrator, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("username, email, and password required")
	}

	if req.Role != RoleSuperAdmin && req.Role != RoleProductManager && req.Role != RoleOrderManager {
		return nil, fmt.Errorf("invalid role")
	}

	// Hash password securely
	passwordHash, err := s.passUtils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin, err := s.repo.CreateAdmin(req.Username, req.Email, passwordHash, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	log.Printf("New admin created: %s (role: %s)", req.Email, req.Role)
	return admin, nil
}

// LogAction logs an admin action
func (s *AdminService) LogAction(adminID int64, action, target, ip string) {
	if err := s.repo.LogAuditAction(adminID, action, target, ip); err != nil {
		log.Printf("Failed to log action: %v", err)
	}
}

// GetAuditLogs retrieves audit logs
func (s *AdminService) GetAuditLogs(limit int) ([]*AuditLog, error) {
	logs, err := s.repo.ListAuditLogs(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	return logs, nil
}

// GetDashboardStats returns dashboard statistics
func (s *AdminService) GetDashboardStats() *DashboardStats {
	// TODO: Implement real stats from other services
	return &DashboardStats{
		NewUsersToday: 42,
		PendingOrders: 15,
		DailyRevenue:  1250.50,
		SystemHealth:  "stable",
	}
}

// ValidateAdminAccess checks if admin has required role
func (s *AdminService) ValidateAdminAccess(adminID int64, requiredRole AdminRole) (bool, error) {
	admin, err := s.repo.FindAdminByID(adminID)
	if err != nil {
		return false, fmt.Errorf("admin not found")
	}

	if requiredRole != "" && admin.Role != RoleSuperAdmin && admin.Role != requiredRole {
		return false, fmt.Errorf("insufficient permissions")
	}

	return true, nil
}
