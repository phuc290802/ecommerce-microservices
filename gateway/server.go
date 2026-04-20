package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	defaultPort               = "8080"
	defaultJWTSecret          = "supersecret"
	defaultRateLimitIP        = 100
	defaultRateLimitUser      = 200
	defaultRateLimitWindowSec = 60
	failureThreshold          = 0.5
	circuitWindowSeconds      = 10
	openCircuitSeconds        = 10
	minRequestsForCircuit     = 5
)

type ServiceRoute struct {
	Name        string
	Prefix      string
	TargetURL   string
	TargetRoute string
}

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu          sync.Mutex
	state       CircuitState
	failures    int
	successes   int
	windowStart time.Time
	openUntil   time.Time
	lastChanged time.Time
}

type rateEntry struct {
	count int
	reset time.Time
}

type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	store  map[string]*rateEntry
}

var (
	jwtSecret     string
	rateLimitIP   = defaultRateLimitIP
	rateLimitUser = defaultRateLimitUser
	routes        = []ServiceRoute{}
	breakers      = map[string]*CircuitBreaker{}
	client        = &http.Client{Timeout: 15 * time.Second}
)

func main() {
	log.SetOutput(os.Stdout)
	jwtSecret = getEnv("JWT_SECRET", defaultJWTSecret)
	port := getEnv("PORT", defaultPort)

	if v := os.Getenv("RATE_LIMIT_IP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimitIP = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_USER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimitUser = n
		}
	}

	productURL := getEnv("PRODUCT_SERVICE_URL", "http://product-service:8081")
	orderURL := getEnv("ORDER_SERVICE_URL", "http://order-service:8082")
	bffURL := getEnv("BFF_SERVICE_URL", "http://bff-service:8083")
	authURL := getEnv("AUTH_SERVICE_URL", "http://auth-service:8084")
	adminURL := getEnv("ADMIN_SERVICE_URL", "http://admin-service:8088")

	routes = []ServiceRoute{
		{Name: "product", Prefix: "/api/products", TargetURL: productURL, TargetRoute: "/products"},
		{Name: "order", Prefix: "/api/orders", TargetURL: orderURL, TargetRoute: "/orders"},
		{Name: "bff", Prefix: "/api/bff", TargetURL: bffURL, TargetRoute: ""},
		{Name: "auth", Prefix: "/api/auth", TargetURL: authURL, TargetRoute: ""},
		{Name: "admin", Prefix: "/api/admin", TargetURL: adminURL, TargetRoute: ""},
	}

	for _, route := range routes {
		breakers[route.Name] = newCircuitBreaker()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/", handleAPI)

	handler := requestIDMiddleware(loggingMiddleware(rateLimitMiddleware(authMiddleware(mux))))

	log.Printf("GATEWAY VERSION 2 STARTING - SKIP AUTH ENABLED")
	log.Printf("API Gateway starting on :%s", port)
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("failed to start gateway: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleAPI: path=%s", r.URL.Path)
	path := r.URL.Path
	for _, route := range routes {
		if strings.HasPrefix(path, route.Prefix) {
			proxyRequest(w, r, route)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func proxyRequest(w http.ResponseWriter, r *http.Request, route ServiceRoute) {
	cb := breakers[route.Name]
	if !cb.AllowRequest() {
		sendFallback(w, route.Name)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	targetPath := strings.TrimPrefix(r.URL.Path, route.Prefix)
	downstreamURL := strings.TrimRight(route.TargetURL, "/") + route.TargetRoute + targetPath
	if r.URL.RawQuery != "" {
		downstreamURL += "?" + r.URL.RawQuery
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(r.Context(), r.Method, downstreamURL, io.NopCloser(strings.NewReader(string(bodyBytes))))
		if err != nil {
			lastErr = err
			break
		}
		copyHeaders(req.Header, r.Header)
		req.Header.Set("X-Request-ID", requestID)
		req.Header.Set("X-Forwarded-For", clientIP(r))
		req.Header.Del("Host")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			cb.RecordFailure()
			if attempt < 3 {
				time.Sleep(150 * time.Millisecond)
				continue
			}
			break
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			cb.RecordFailure()
			if attempt < 3 {
				time.Sleep(150 * time.Millisecond)
				continue
			}
		} else {
			cb.RecordSuccess()
		}

		copyResponse(w, resp)
		return
	}

	log.Printf("downstream request failed for %s: %v", route.Name, lastErr)
	cb.RecordFailure()
	http.Error(w, "service unavailable", http.StatusBadGateway)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log for debugging
		log.Printf("AuthMiddleware: %s %s", r.Method, r.URL.Path)

		// 1. Allow OPTIONS requests (CORS preflight)
		// 2. Allow /health
		// 3. Allow everything under /api/auth (login, register, forgot-password, etc.)
		if r.Method == http.MethodOptions || 
		   r.URL.Path == "/health" || 
		   strings.HasPrefix(r.URL.Path, "/api/auth") ||
		   strings.HasPrefix(r.URL.Path, "/api/admin/login") ||
		   strings.HasPrefix(r.URL.Path, "/auth") {
			log.Printf("AuthMiddleware: Skipping auth for %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		fmt.Printf("authMiddleware: checking auth for %s\n", r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing bearer token (v2)", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, userID, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	ipLimiter := newRateLimiter(rateLimitIP, defaultRateLimitWindowSec)
	userLimiter := newRateLimiter(rateLimitUser, defaultRateLimitWindowSec)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := clientIP(r)
		if !ipLimiter.Allow(clientIP) {
			http.Error(w, "rate limit exceeded (ip)", http.StatusTooManyRequests)
			return
		}

		if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
			if !userLimiter.Allow(userID) {
				http.Error(w, "rate limit exceeded (user)", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)
		duration := time.Since(start)
		log.Printf("method=%s path=%s status=%d duration=%s client_ip=%s request_id=%s",
			r.Method, r.URL.Path, rr.statusCode, duration.String(), clientIP(r), r.Header.Get("X-Request-ID"))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			r.Header.Set("X-Request-ID", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func validateJWT(tokenString string) (*jwt.RegisteredClaims, string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, "", err
	}
	if !token.Valid {
		return nil, "", fmt.Errorf("invalid token")
	}

	userID := claims.Subject
	if userID == "" {
		userID = "anonymous"
	}
	return claims, userID, nil
}

func newCircuitBreaker() *CircuitBreaker {
	now := time.Now()
	return &CircuitBreaker{state: StateClosed, windowStart: now, lastChanged: now}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == StateOpen {
		if now.After(cb.openUntil) {
			cb.state = StateHalfOpen
			cb.failures = 0
			cb.successes = 0
			cb.windowStart = now
			cb.lastChanged = now
			return true
		}
		return false
	}
	if now.Sub(cb.windowStart) > time.Second*circuitWindowSeconds {
		cb.windowStart = now
		cb.failures = 0
		cb.successes = 0
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if now.Sub(cb.windowStart) > time.Second*circuitWindowSeconds {
		cb.windowStart = now
		cb.failures = 0
		cb.successes = 0
	}
	cb.successes++

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
		cb.successes = 0
		cb.lastChanged = now
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if now.Sub(cb.windowStart) > time.Second*circuitWindowSeconds {
		cb.windowStart = now
		cb.failures = 0
		cb.successes = 0
	}
	cb.failures++

	total := cb.failures + cb.successes
	failureRate := 0.0
	if total > 0 {
		failureRate = float64(cb.failures) / float64(total)
	}

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.openUntil = now.Add(time.Second * openCircuitSeconds)
		cb.lastChanged = now
		return
	}

	if total >= minRequestsForCircuit && failureRate >= failureThreshold {
		cb.state = StateOpen
		cb.openUntil = now.Add(time.Second * openCircuitSeconds)
		cb.lastChanged = now
	}
}

func sendFallback(w http.ResponseWriter, serviceName string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": fmt.Sprintf("%s service temporarily unavailable", serviceName),
	})
}

func newRateLimiter(limit, windowSeconds int) *RateLimiter {
	return &RateLimiter{limit: limit, window: time.Duration(windowSeconds) * time.Second, store: map[string]*rateEntry{}}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.store[key]
	now := time.Now()
	if !ok || now.After(entry.reset) {
		rl.store[key] = &rateEntry{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	return true
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func generateRequestID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(999999999999))
	if err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req-%d", n.Int64())
}
