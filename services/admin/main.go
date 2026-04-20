package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/go-sql-driver/mysql"
)

type AdminRole string

const (
	RoleSuperAdmin     AdminRole = "super_admin"
	RoleProductManager AdminRole = "product_manager"
	RoleOrderManager   AdminRole = "order_manager"
)

type Administrator struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Role      AdminRole `json:"role"`
	Status    string    `json:"status"` // active, locked
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	AdminID   int64     `json:"admin_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DB_DSN")
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)

	if err := initSchema(); err != nil {
		log.Printf("Warning: Schema init failed: %v", err)
	}

	mux := http.NewServeMux()
	
	// Public (Login)
	mux.HandleFunc("/login", handleLogin)

	// Protected Admin Routes
	mux.HandleFunc("/users", authMiddleware(RoleSuperAdmin, handleListAdmins))
	mux.HandleFunc("/users/create", authMiddleware(RoleSuperAdmin, handleCreateAdmin))
	mux.HandleFunc("/audit-logs", authMiddleware(RoleSuperAdmin, handleListAuditLogs))
	
	// Stats & Dashboard
	mux.HandleFunc("/dashboard/stats", authMiddleware("", handleDashboardStats))

	log.Printf("Admin Service starting on :8088")
	log.Fatal(http.ListenAndServe(":8088", mux))
}

func initSchema() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS administrators (
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		admin_id BIGINT NOT NULL,
		action VARCHAR(100) NOT NULL,
		target VARCHAR(255),
		ip VARCHAR(50),
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	
	// Create default super admin if empty
	var count int
	db.QueryRow("SELECT COUNT(*) FROM administrators").Scan(&count)
	if count == 0 {
		db.Exec("INSERT INTO administrators (username, email, password_hash, role) VALUES (?, ?, ?, ?)", 
			"superadmin", "admin@shopverse.com", "admin123", RoleSuperAdmin)
	}
	
	return err
}

// Middleware & Auth
func authMiddleware(requiredRole AdminRole, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real microservice, we'd validate JWT here or trust Gateway
		// For this implementation, we assume Gateway passed user_id/role in headers
		adminID := r.Header.Get("X-Admin-ID")
		adminRole := AdminRole(r.Header.Get("X-Admin-Role"))

		if adminID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if requiredRole != "" && adminRole != RoleSuperAdmin && adminRole != requiredRole {
			http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func logAction(adminID int64, action, target, ip string) {
	_, err := db.Exec("INSERT INTO audit_logs (admin_id, action, target, ip) VALUES (?, ?, ?, ?)", 
		adminID, action, target, ip)
	if err != nil {
		log.Printf("Failed to log action: %v", err)
	}
}

// Handlers
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&payload)

	var admin Administrator
	var storedPassword string
	err := db.QueryRow("SELECT id, username, email, password_hash, role FROM administrators WHERE email = ?", 
		payload.Email).Scan(&admin.ID, &admin.Username, &admin.Email, &storedPassword, &admin.Role)

	if err != nil || storedPassword != payload.Password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := createToken(admin.ID, string(admin.Role))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user":  admin,
	})
}

func createToken(adminID int64, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecret"
	}
	
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", adminID),
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func handleListAdmins(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, username, email, role, status, created_at FROM administrators")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []Administrator = []Administrator{}
	for rows.Next() {
		var a Administrator
		rows.Scan(&a.ID, &a.Username, &a.Email, &a.Role, &a.Status, &a.CreatedAt)
		list = append(list, a)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	var a Administrator
	json.NewDecoder(r.Body).Decode(&a)
	
	_, err := db.Exec("INSERT INTO administrators (username, email, password_hash, role) VALUES (?, ?, ?, ?)",
		a.Username, a.Email, a.Password, a.Role)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT id, admin_id, action, target, ip, timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT 100")
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		rows.Scan(&l.ID, &l.AdminID, &l.Action, &l.Target, &l.IP, &l.Timestamp)
		logs = append(logs, l)
	}
	json.NewEncoder(w).Encode(logs)
}

func handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Simple mock stats for demonstration
	stats := map[string]any{
		"new_users_today": 42,
		"pending_orders": 15,
		"daily_revenue": 1250.50,
		"system_health": "stable",
	}
	json.NewEncoder(w).Encode(stats)
}
