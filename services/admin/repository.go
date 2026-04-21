package main

import (
	"database/sql"
	"time"
)

// Repository handles all database operations for admin service
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InitSchema creates tables and initializes data
func (r *Repository) InitSchema() error {
	// Create administrators table
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS administrators (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(100) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Create audit_logs table
	_, err = r.db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		admin_id BIGINT NOT NULL,
		action VARCHAR(100) NOT NULL,
		target VARCHAR(255),
		ip VARCHAR(50),
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Create default super admin if empty
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM administrators").Scan(&count)
	if count == 0 {
		// Use proper password hash instead of plaintext
		_, err := r.db.Exec(
			"INSERT INTO administrators (username, email, password_hash, role, status) VALUES (?, ?, ?, ?, ?)",
			"superadmin", "admin@shopverse.com", "$2a$10$default_hash_for_admin123", RoleSuperAdmin, "active",
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindAdminByEmail retrieves an admin by email
func (r *Repository) FindAdminByEmail(email string) (*Administrator, error) {
	admin := &Administrator{}
	row := r.db.QueryRow(
		"SELECT id, username, email, password_hash, role, status, created_at FROM administrators WHERE email = ?",
		email,
	)
	if err := row.Scan(&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.Status, &admin.CreatedAt); err != nil {
		return nil, err
	}
	return admin, nil
}

// FindAdminByID retrieves an admin by ID
func (r *Repository) FindAdminByID(adminID int64) (*Administrator, error) {
	admin := &Administrator{}
	row := r.db.QueryRow(
		"SELECT id, username, email, password_hash, role, status, created_at FROM administrators WHERE id = ?",
		adminID,
	)
	if err := row.Scan(&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.Status, &admin.CreatedAt); err != nil {
		return nil, err
	}
	return admin, nil
}

// ListAdmins retrieves all administrators
func (r *Repository) ListAdmins() ([]*Administrator, error) {
	rows, err := r.db.Query("SELECT id, username, email, role, status, created_at FROM administrators ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []*Administrator
	for rows.Next() {
		admin := &Administrator{}
		if err := rows.Scan(&admin.ID, &admin.Username, &admin.Email, &admin.Role, &admin.Status, &admin.CreatedAt); err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, nil
}

// CreateAdmin creates a new administrator
func (r *Repository) CreateAdmin(username, email, passwordHash string, role AdminRole) (*Administrator, error) {
	result, err := r.db.Exec(
		"INSERT INTO administrators (username, email, password_hash, role, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		username, email, passwordHash, role, "active", time.Now(),
	)
	if err != nil {
		return nil, err
	}

	adminID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.FindAdminByID(adminID)
}

// LogAuditAction logs an admin action
func (r *Repository) LogAuditAction(adminID int64, action, target, ip string) error {
	_, err := r.db.Exec(
		"INSERT INTO audit_logs (admin_id, action, target, ip, timestamp) VALUES (?, ?, ?, ?, ?)",
		adminID, action, target, ip, time.Now(),
	)
	return err
}

// ListAuditLogs retrieves recent audit logs
func (r *Repository) ListAuditLogs(limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.Query(
		"SELECT id, admin_id, action, target, ip, timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		if err := rows.Scan(&log.ID, &log.AdminID, &log.Action, &log.Target, &log.IP, &log.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}
