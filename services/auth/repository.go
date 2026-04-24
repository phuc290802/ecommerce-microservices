package main

import (
	"database/sql"
	"time"
)

// Repository handles all database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UserExists checks if a user exists by email or username
func (r *Repository) UserExists(email, username string) (bool, error) {
	var count int
	row := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", email, username)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(username, email, phone, passwordHash string) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO users (username, email, phone, password_hash) VALUES (?, ?, ?, ?)",
		username, email, phone, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// FindUserByEmail retrieves a user by email
func (r *Repository) FindUserByEmail(email string) (*User, error) {
	user := &User{}
	row := r.db.QueryRow(
		"SELECT id, username, email, phone, password_hash, created_at FROM users WHERE email = ?",
		email,
	)
	var createdAt time.Time
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)
	return user, nil
}

// FindUserByID retrieves a user by ID
func (r *Repository) FindUserByID(userID int64) (*User, error) {
	user := &User{}
	row := r.db.QueryRow(
		"SELECT id, username, email, phone, password_hash, created_at FROM users WHERE id = ?",
		userID,
	)
	var createdAt time.Time
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)
	return user, nil
}

// UpdatePassword updates a user's password
func (r *Repository) UpdatePassword(userID int64, passwordHash string) error {
	_, err := r.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, userID)
	return err
}

// InitSchema creates the users table if it doesn't exist
func (r *Repository) InitSchema() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        username VARCHAR(100) NOT NULL UNIQUE,
        email VARCHAR(255) NOT NULL UNIQUE,
        phone VARCHAR(30) DEFAULT '',
        password_hash VARCHAR(255) NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	return err
}
