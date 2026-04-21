package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Handlers contains all HTTP handlers for product service
type Handlers struct {
	service *ProductService
}

// NewHandlers creates a new Handlers instance
func NewHandlers(service *ProductService) *Handlers {
	return &Handlers{service: service}
}

// HandleHealth returns service health status
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleListProducts retrieves all products
func (h *Handlers) HandleListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse common filters
	query := r.URL.Query()
	filter := map[string]string{}
	if v := query.Get("name"); v != "" {
		filter["name"] = v
	}
	if v := query.Get("attribute"); v != "" {
		filter["attribute"] = v
	}
	if v := query.Get("category"); v != "" {
		filter["category_id"] = v
	}

	// pagination & sorting
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	sortBy := query.Get("sort_by")
	sortOrder := query.Get("sort_order")

	// price range
	var minPrice, maxPrice *float64
	if v := query.Get("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			minPrice = &f
		}
	}
	if v := query.Get("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			maxPrice = &f
		}
	}

	products, err := h.service.GetProducts(filter, minPrice, maxPrice, page, pageSize, sortBy, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if products == nil {
		products = []*Product{}
	}
	responses := make([]*ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}
	_ = json.NewEncoder(w).Encode(responses)
}

// HandleBatchGetProducts returns multiple products by ids (ids=1,2,3)
func (h *Handlers) HandleBatchGetProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		http.Error(w, "missing ids param", http.StatusBadRequest)
		return
	}
	parts := strings.Split(idsStr, ",")
	var ids []int64
	for _, p := range parts {
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	products, err := h.service.GetProductsByIDs(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := make([]*ProductResponse, len(products))
	for i, p := range products {
		resp[i] = p.ToResponse()
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleUploadImage accepts base64 JSON or multipart file upload and returns a URL path
func (h *Handlers) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, 0755)

	// First, try multipart
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer file.Close()
		filename := time.Now().Format("20060102_150405_") + header.Filename
		dstPath := filepath.Join(uploadDir, filename)
		out, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "cannot save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "failed to save file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"url": "/uploads/" + filename})
		return
	}

	// Otherwise expect JSON with base64
	var body struct {
		Filename    string `json:"filename"`
		ImageBase64 string `json:"image_base64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ImageBase64 == "" || body.Filename == "" {
		http.Error(w, "filename and image_base64 required", http.StatusBadRequest)
		return
	}
	data, err := base64.StdEncoding.DecodeString(body.ImageBase64)
	if err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}
	filename := time.Now().Format("20060102_150405_") + filepath.Base(body.Filename)
	dstPath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": "/uploads/" + filename})
}

// HandleGetProductByID retrieves a single product by ID
func (h *Handlers) HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract product ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	if idStr == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProductByID(productID)
	if err != nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(product.ToResponse())
}

// HandleCreateProduct creates a new product
func (h *Handlers) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.service.CreateProduct(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(product.ToResponse())
}

// HandleUpdateProduct updates an existing product
func (h *Handlers) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.service.UpdateProduct(productID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(product.ToResponse())
}

// HandleDeleteProduct deletes a product
func (h *Handlers) HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProduct(productID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleProductByID delegates to GET/PUT/DELETE handlers based on method
func (h *Handlers) HandleProductByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.HandleGetProductByID(w, r)
	case http.MethodPut:
		h.HandleUpdateProduct(w, r)
	case http.MethodDelete:
		h.HandleDeleteProduct(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
