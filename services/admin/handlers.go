package main

import (
	"encoding/json"
	"net/http"
)

// Handlers contains all HTTP handlers for admin service
type Handlers struct {
	service *AdminService
}

// NewHandlers creates a new Handlers instance
func NewHandlers(service *AdminService) *Handlers {
	return &Handlers{service: service}
}

// HandleLogin authenticates an admin
func (h *Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.LoginAdmin(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleListAdmins retrieves all administrators (SuperAdmin only)
func (h *Handlers) HandleListAdmins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	admins, err := h.service.ListAdmins()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if admins == nil {
		admins = []*Administrator{}
	}
	_ = json.NewEncoder(w).Encode(admins)
}

// HandleCreateAdmin creates a new administrator (SuperAdmin only)
func (h *Handlers) HandleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	admin, err := h.service.CreateAdmin(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(admin)
}

// HandleListAuditLogs retrieves audit logs (SuperAdmin only)
func (h *Handlers) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := h.service.GetAuditLogs(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if logs == nil {
		logs = []*AuditLog{}
	}
	_ = json.NewEncoder(w).Encode(logs)
}

// HandleDashboardStats returns dashboard statistics (Protected)
func (h *Handlers) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.service.GetDashboardStats()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
