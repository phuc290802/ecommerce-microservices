package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	svc *CategoryService
}

func NewHandlers(svc *CategoryService) *Handlers { return &Handlers{svc: svc} }

func (h *Handlers) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := h.svc.List()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	case http.MethodPost:
		var payload Category
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if payload.Name == "" || payload.Slug == "" {
			http.Error(w, "name and slug required", http.StatusBadRequest)
			return
		}
		created := h.svc.Create(payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handlers) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/categories/")
	if path == "" || path == "/" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(path, "slug/") {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	idStr := parts[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if len(parts) > 1 && parts[1] == "products" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		products, err := h.svc.GetProductsByCategory(id)
		if err != nil {
			http.Error(w, "product service error", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(products)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if c, ok := h.svc.GetByID(id); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(c)
			return
		}
		http.Error(w, "category not found", http.StatusNotFound)
	case http.MethodPut:
		var payload Category
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		payload.ID = id
		if !h.svc.Update(payload) {
			http.Error(w, "category not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if !h.svc.Delete(id) {
			http.Error(w, "category not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handlers) handleCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := strings.TrimPrefix(r.URL.Path, "/categories/slug/")
	if slug == "" {
		http.Error(w, "missing slug", http.StatusBadRequest)
		return
	}
	if c, ok := h.svc.GetBySlug(slug); ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(c)
		return
	}
	http.Error(w, "category not found", http.StatusNotFound)
}

func (h *Handlers) handleCategoryTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tree := h.svc.GetTree()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tree)
}

func (h *Handlers) handleRebuildTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tree := h.svc.RebuildTree()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tree)
}
