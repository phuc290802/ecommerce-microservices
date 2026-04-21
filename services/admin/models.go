package main

import "time"

// AdminRole represents the role type for administrators
type AdminRole string

const (
	RoleSuperAdmin     AdminRole = "super_admin"
	RoleProductManager AdminRole = "product_manager"
	RoleOrderManager   AdminRole = "order_manager"
)

// Administrator represents an admin user
type Administrator struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         AdminRole `json:"role"`
	Status       string    `json:"status"` // active, locked
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int64     `json:"id"`
	AdminID   int64     `json:"admin_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// LoginRequest represents admin login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string         `json:"token"`
	User  *Administrator `json:"user"`
}

// CreateAdminRequest represents create admin payload
type CreateAdminRequest struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     AdminRole `json:"role"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	NewUsersToday int     `json:"new_users_today"`
	PendingOrders int     `json:"pending_orders"`
	DailyRevenue  float64 `json:"daily_revenue"`
	SystemHealth  string  `json:"system_health"`
}
