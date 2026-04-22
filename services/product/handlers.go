package main

import (
"encoding/json"
"log"
"net/http"
"strconv"
"strings"
)

type Handlers struct {
svc *ProductService
}

func NewHandlers(svc *ProductService) *Handlers {
return &Handlers{svc: svc}
}

func (h *Handlers) handleHealth(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) handleProducts(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
products, err := h.svc.List()
if err != nil {
http.Error(w, "failed to list products", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(products); err != nil {
http.Error(w, "failed to encode response", http.StatusInternalServerError)
}

case http.MethodPost:
var payload Product
if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
http.Error(w, "invalid request body", http.StatusBadRequest)
return
}

if payload.Name == "" || payload.Price <= 0 {
http.Error(w, "name and price are required and price must be positive", http.StatusBadRequest)
return
}

created, err := h.svc.Create(payload)
if err != nil {
http.Error(w, "failed to create product", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_ = json.NewEncoder(w).Encode(created)

default:
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

func (h *Handlers) handleProductByID(w http.ResponseWriter, r *http.Request) {
idStr := strings.TrimPrefix(r.URL.Path, "/products/")
if idStr == "" || idStr == "/" {
log.Printf("handleProductByID: missing product id, path=%s", r.URL.Path)
http.Error(w, "missing product id", http.StatusBadRequest)
return
}

id, err := strconv.ParseInt(idStr, 10, 64)
if err != nil {
log.Printf("handleProductByID: invalid product id %s, path=%s", idStr, r.URL.Path)
http.Error(w, "invalid product id", http.StatusBadRequest)
return
}

log.Printf("handleProductByID: method=%s, id=%d, path=%s", r.Method, id, r.URL.Path)

switch r.Method {
case http.MethodGet:
log.Printf("handleProductByID: getting product id=%s", idStr)
product, err := h.svc.GetByID(id)
if err != nil {
log.Printf("handleProductByID: product not found, id=%d", id)
http.Error(w, "product not found", http.StatusNotFound)
return
}

log.Printf("handleProductByID: product found, id=%d, name=%s", id, product.Name)
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(product)

case http.MethodPut:
var payload Product
if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
http.Error(w, "invalid request body", http.StatusBadRequest)
return
}

payload.ID = id
if ok, err := h.svc.Update(payload); !ok || err != nil {
http.Error(w, "failed to update product", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(payload)

case http.MethodDelete:
if ok, err := h.svc.Delete(id); !ok || err != nil {
http.Error(w, "failed to delete product", http.StatusInternalServerError)
return
}

w.WriteHeader(http.StatusNoContent)

default:
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}
