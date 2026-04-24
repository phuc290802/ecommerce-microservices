package main

import (
	"context"
	"net/http"
	"strings"
)

// AdminAuthMiddleware validates JWT token and checks role-based access
func AdminAuthMiddleware(tokenUtils *TokenUtils, service *AdminService, requiredRole AdminRole, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Fallback to X-Admin-Token header for internal service calls
			authHeader = r.Header.Get("X-Admin-Token")
			if authHeader == "" {
				http.Error(w, "authorization header missing", http.StatusUnauthorized)
				return
			}
		}

		// Extract bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		var tokenString string
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		} else {
			tokenString = authHeader
		}

		// Validate token
		claims, err := tokenUtils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Extract admin ID and role
		adminID, err := ExtractAdminIDFromToken(claims)
		if err != nil {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		role, err := ExtractRoleFromToken(claims)
		if err != nil {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		// Check role-based access
		if requiredRole != "" && AdminRole(role) != RoleSuperAdmin && AdminRole(role) != requiredRole {
			http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		// Store admin info in request headers for handler use
		r.Header.Set("X-Admin-ID", string(rune(adminID)))
		r.Header.Set("X-Admin-Role", role)

		// Also store in context for later retrieval
		ctx := context.WithValue(r.Context(), "admin_id", adminID)
		ctx = context.WithValue(ctx, "admin_role", role)

		next(w, r.WithContext(ctx))
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
