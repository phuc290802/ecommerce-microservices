package main

import (
"encoding/json"
"log"
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
log.Printf("handleCategories: method=%s, path=%s", r.Method, r.URL.Path)

switch r.Method {
case http.MethodGet:
list := h.svc.List()
log.Printf("handleCategories: returning %d categories", len(list))
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(list)
case http.MethodPost:
var payload Category
if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
log.Printf("handleCategories: invalid request body: %v", err)
http.Error(w, "invalid request body", http.StatusBadRequest)
return
}
if payload.Name == "" || payload.Slug == "" {
log.Printf("handleCategories: missing name or slug")
http.Error(w, "name and slug required", http.StatusBadRequest)
return
}
created := h.svc.Create(payload)
log.Printf("handleCategories: created category id=%d, name=%s", created.ID, created.Name)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_ = json.NewEncoder(w).Encode(created)
default:
log.Printf("handleCategories: method not allowed: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

func (h *Handlers) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
path := strings.TrimPrefix(r.URL.Path, "/categories/")
if path == "" || path == "/" {
log.Printf("handleCategoryByID: missing id, path=%s", r.URL.Path)
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
log.Printf("handleCategoryByID: invalid id %s, path=%s", idStr, r.URL.Path)
http.Error(w, "invalid id", http.StatusBadRequest)
return
}

log.Printf("handleCategoryByID: method=%s, id=%d, path=%s", r.Method, id, r.URL.Path)

if len(parts) > 1 && parts[1] == "products" {
if r.Method != http.MethodGet {
log.Printf("handleCategoryByID: method not allowed for products: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
log.Printf("handleCategoryByID: getting products for category id=%d", id)
products, err := h.svc.GetProductsByCategory(id)
if err != nil {
log.Printf("handleCategoryByID: product service error: %v", err)
http.Error(w, "product service error", http.StatusBadGateway)
return
}
log.Printf("handleCategoryByID: returning %d products for category id=%d", len(products), id)
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(products)
return
}
switch r.Method {
case http.MethodGet:
if c, ok := h.svc.GetByID(id); ok {
log.Printf("handleCategoryByID: category found, id=%d, name=%s", id, c.Name)
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(c)
return
}
log.Printf("handleCategoryByID: category not found, id=%d", id)
http.Error(w, "category not found", http.StatusNotFound)
case http.MethodPut:
var payload Category
if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
log.Printf("handleCategoryByID: invalid request body: %v", err)
http.Error(w, "invalid request body", http.StatusBadRequest)
return
}
payload.ID = id
if !h.svc.Update(payload) {
log.Printf("handleCategoryByID: category not found for update, id=%d", id)
http.Error(w, "category not found", http.StatusNotFound)
return
}
log.Printf("handleCategoryByID: updated category id=%d", id)
w.WriteHeader(http.StatusNoContent)
case http.MethodDelete:
if !h.svc.Delete(id) {
log.Printf("handleCategoryByID: category not found for delete, id=%d", id)
http.Error(w, "category not found", http.StatusNotFound)
return
}
log.Printf("handleCategoryByID: deleted category id=%d", id)
w.WriteHeader(http.StatusNoContent)
default:
log.Printf("handleCategoryByID: method not allowed: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

func (h *Handlers) handleCategoryTree(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
log.Printf("handleCategoryTree: method not allowed: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
log.Printf("handleCategoryTree: getting category tree")
tree := h.svc.GetTree()
log.Printf("handleCategoryTree: returning tree with %d root categories", len(tree))
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(tree)
}

func (h *Handlers) handleRebuildTree(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
log.Printf("handleRebuildTree: method not allowed: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
log.Printf("handleRebuildTree: rebuilding category tree")
h.svc.RebuildTree()
log.Printf("handleRebuildTree: tree rebuilt successfully")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) handleCategoryBySlug(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
log.Printf("handleCategoryBySlug: method not allowed: %s", r.Method)
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}
slug := strings.TrimPrefix(r.URL.Path, "/categories/slug/")
if slug == "" {
log.Printf("handleCategoryBySlug: missing slug, path=%s", r.URL.Path)
http.Error(w, "missing slug", http.StatusBadRequest)
return
}
log.Printf("handleCategoryBySlug: getting category by slug=%s", slug)
if c, ok := h.svc.GetBySlug(slug); ok {
log.Printf("handleCategoryBySlug: category found, slug=%s, name=%s", slug, c.Name)
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(c)
return
}
log.Printf("handleCategoryBySlug: category not found, slug=%s", slug)
http.NotFound(w, r)
}
