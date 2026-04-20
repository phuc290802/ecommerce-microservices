package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultPort            = "8084"
	defaultRedis           = "redis:6379"
	defaultJWTSecret       = "supersecret"
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
	defaultResetTokenTTL   = 15 * time.Minute
	defaultOTPTokenTTL     = 5 * time.Minute
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Email    string `json:"email"`
}

var (
	db          *sql.DB
	redisClient *redis.Client
	jwtSecret   string
)

func main() {
	jwtSecret = getEnv("JWT_SECRET", defaultJWTSecret)
	redisAddr := getEnv("REDIS_ADDR", defaultRedis)
	dbDsn := getEnv("DB_DSN", "user:password@tcp(mysql:3306)/ecommerce?charset=utf8mb4&parseTime=true")
	port := getEnv("PORT", defaultPort)

	var err error
	db, err = sql.Open("mysql", dbDsn)
	if err != nil {
		log.Fatalf("failed connect mysql: %v", err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if err = db.Ping(); err != nil {
		log.Fatalf("mysql ping failed: %v", err)
	}
	if err = initSchema(db); err != nil {
		log.Fatalf("init schema failed: %v", err)
	}

	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/refresh", handleRefresh)
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/internal/verify", handleVerifyToken)
	mux.HandleFunc("/otp/request", handleOTPRequest)
	mux.HandleFunc("/otp/validate", handleOTPValidate)
	mux.HandleFunc("/forgot-password", handleForgotPassword)
	mux.HandleFunc("/reset-password", handleResetPassword)

	addr := ":" + port
	log.Printf("auth service starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        username VARCHAR(100) NOT NULL UNIQUE,
        email VARCHAR(255) NOT NULL UNIQUE,
        phone VARCHAR(30) DEFAULT '',
        password_hash VARCHAR(255) NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	return err
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.Username == "" || payload.Email == "" || payload.Password == "" {
		http.Error(w, "username, email, password required", http.StatusBadRequest)
		return
	}

	if exists, _ := userExists(payload.Email, payload.Username); exists {
		http.Error(w, "username or email already used", http.StatusConflict)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	userID, err := createUser(payload.Username, payload.Email, payload.Phone, string(passwordHash))
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": userID, "username": payload.Username, "email": payload.Email})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := findUserByEmail(payload.Email)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)) != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := createAccessToken(user)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := createRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "failed to create refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(defaultRefreshTokenTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"access_token": token, "expires_in": 900})
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "refresh_token missing", http.StatusUnauthorized)
		return
	}

	userID, err := validateRefreshToken(cookie.Value)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	user, err := findUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := createAccessToken(user)
	if err != nil {
		http.Error(w, "failed to create access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := rotateRefreshToken(cookie.Value, user.ID)
	if err != nil {
		http.Error(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(defaultRefreshTokenTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"access_token": newAccessToken, "expires_in": 900})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = deleteRefreshToken(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", HttpOnly: true, Path: "/", MaxAge: -1, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	claims, err := validateToken(payload.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(claims)
}

func handleOTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Purpose string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Email == "" && payload.Phone == "" {
		http.Error(w, "email or phone required", http.StatusBadRequest)
		return
	}
	if payload.Purpose == "" {
		payload.Purpose = "verification"
	}
	code := generateOTPCode()
	key := fmt.Sprintf("otp:%s:%s", payload.Purpose, strings.ToLower(payload.Email+payload.Phone))
	if err := redisClient.Set(context.Background(), key, code, defaultOTPTokenTTL).Err(); err != nil {
		http.Error(w, "failed to store otp", http.StatusInternalServerError)
		return
	}
	log.Printf("OTP for %s/%s purpose=%s: %s", payload.Email, payload.Phone, payload.Purpose, code)
	w.WriteHeader(http.StatusNoContent)
}

func handleOTPValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Purpose string `json:"purpose"`
		Code    string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("otp:%s:%s", payload.Purpose, strings.ToLower(payload.Email+payload.Phone))
	stored, err := redisClient.Get(context.Background(), key).Result()
	if err != nil || stored != payload.Code {
		http.Error(w, "invalid otp", http.StatusUnauthorized)
		return
	}
	_ = redisClient.Del(context.Background(), key).Err()
	w.WriteHeader(http.StatusNoContent)
}

func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := findUserByEmail(payload.Email)
	if err != nil {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}
	token := generateRandomToken()
	key := fmt.Sprintf("reset:%s", token)
	if err := redisClient.Set(context.Background(), key, fmt.Sprint(user.ID), defaultResetTokenTTL).Err(); err != nil {
		http.Error(w, "failed to create reset token", http.StatusInternalServerError)
		return
	}
	log.Printf("Password reset link for %s: http://localhost:5173/reset-password?token=%s", user.Email, token)
	w.WriteHeader(http.StatusNoContent)
}

func handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Token == "" || payload.Password == "" {
		http.Error(w, "token and password required", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("reset:%s", payload.Token)
	userIDStr, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid token payload", http.StatusInternalServerError)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	if err := updatePassword(userID, string(passwordHash)); err != nil {
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}
	_ = redisClient.Del(context.Background(), key).Err()
	w.WriteHeader(http.StatusNoContent)
}

func userExists(email, username string) (bool, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", email, username)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func createUser(username, email, phone, passwordHash string) (int64, error) {
	result, err := db.Exec("INSERT INTO users (username, email, phone, password_hash) VALUES (?, ?, ?, ?)", username, email, phone, passwordHash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func findUserByEmail(email string) (*User, error) {
	user := &User{}
	row := db.QueryRow("SELECT id, username, email, phone, password_hash, created_at FROM users WHERE email = ?", email)
	var createdAt time.Time
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)
	return user, nil
}

func findUserByID(userID int64) (*User, error) {
	user := &User{}
	row := db.QueryRow("SELECT id, username, email, phone, password_hash, created_at FROM users WHERE id = ?", userID)
	var createdAt time.Time
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = createdAt.Format(time.RFC3339)
	return user, nil
}

func updatePassword(userID int64, passwordHash string) error {
	_, err := db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, userID)
	return err
}

func createAccessToken(user *User) (string, error) {
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
	return token.SignedString([]byte(jwtSecret))
}

func createRefreshToken(userID int64) (string, error) {
	token := generateRandomToken()
	key := fmt.Sprintf("refresh:%s", token)
	if err := redisClient.Set(context.Background(), key, fmt.Sprint(userID), defaultRefreshTokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func validateRefreshToken(token string) (int64, error) {
	key := fmt.Sprintf("refresh:%s", token)
	userIDStr, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(userIDStr, 10, 64)
}

func rotateRefreshToken(oldToken string, userID int64) (string, error) {
	if err := deleteRefreshToken(oldToken); err != nil {
		return "", err
	}
	return createRefreshToken(userID)
}

func deleteRefreshToken(token string) error {
	return redisClient.Del(context.Background(), fmt.Sprintf("refresh:%s", token)).Err()
}

func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func generateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func generateRandomToken() string {
	raw := make([]byte, 32)
	_, err := rand.Read(raw)
	if err != nil {
		return fmt.Sprintf("rt-%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
