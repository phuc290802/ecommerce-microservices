package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type CartService struct {
	db         *sql.DB
	httpc      *http.Client
	productSvc string
}

func main() {
	// DB connection (env vars)
	dsn := os.Getenv("DB_DSN") // e.g. "user:pass@tcp(mysql:3306)/ecommerce"
	if dsn == "" {
		log.Fatal("DB_DSN required")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	productSvc := os.Getenv("PRODUCT_SERVICE_URL")
	if productSvc == "" {
		productSvc = "http://product:8083" // default docker
	}
	service := &CartService{
		db:         db,
		httpc:      &http.Client{Timeout: 5 * time.Second},
		productSvc: productSvc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.handleHealth)
	mux.HandleFunc("/cart", service.handleCart)           // GET list/POST add
	mux.HandleFunc("/cart/item/", service.handleCartItem) // PUT update/DELETE remove
	mux.HandleFunc("/cart/merge", service.handleMerge)
	mux.HandleFunc("/cart/clear", service.handleClear)

	addr := ":8090"
	log.Printf("cart service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *CartService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "db": "connected"})
}

type CartItem struct {
	ID        int64       `json:"id"`
	CartKey   string      `json:"cart_key"`
	ProductID int64       `json:"product_id"`
	Variant   interface{} `json:"variant,omitempty"`
	Quantity  int         `json:"quantity"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

type AddCartRequest struct {
	CartKey   string      `json:"cart_key"`
	ProductID int64       `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Variant   interface{} `json:"variant,omitempty"`
}

type CartResponse struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}

type ProductPrice struct {
	Price float64 `json:"price"`
}

// getPrice calls product service for price
func (s *CartService) getPrice(productID int64) (float64, error) {
	resp, err := s.httpc.Get(fmt.Sprintf("%s/products/%d", s.productSvc, productID))

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var prod struct {
		Price float64 `json:"price"`
	}
	if err := json.Unmarshal(body, &prod); err != nil {
		return 0, err
	}
	return prod.Price, nil
}

func (s *CartService) handleCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		cartKey := r.URL.Query().Get("cart_key")
		if cartKey == "" {
			http.Error(w, "cart_key required", http.StatusBadRequest)
			return
		}
		rows, err := s.db.Query("SELECT id, cart_key, product_id, variant, quantity, created_at, updated_at FROM cart_items WHERE cart_key = ?", cartKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []CartItem
		for rows.Next() {
			var item CartItem
			var variantStr string
			if err := rows.Scan(&item.ID, &item.CartKey, &item.ProductID, &variantStr, &item.Quantity, &item.CreatedAt, &item.UpdatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if variantStr != "" {
				json.Unmarshal([]byte(variantStr), &item.Variant)
			}
			items = append(items, item)
		}
		// Calculate total
		var total float64
		for _, item := range items {
			price, err := s.getPrice(item.ProductID)
			if err != nil {
				log.Printf("price fetch failed for %d: %v", item.ProductID, err)
				continue
			}
			total += price * float64(item.Quantity)
		}
		json.NewEncoder(w).Encode(CartResponse{Items: items, Total: total})

	case http.MethodPost:
		var req AddCartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.CartKey == "" || req.ProductID == 0 || req.Quantity <= 0 {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		// Upsert logic: REPLACE INTO for simplicity (mysql)
		variantBytes, _ := json.Marshal(req.Variant)
		_, err := s.db.Exec("REPLACE INTO cart_items (cart_key, product_id, variant, quantity) VALUES (?, ?, ?, ?)", req.CartKey, req.ProductID, string(variantBytes), req.Quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *CartService) handleCartItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/cart/item/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodDelete:
		_, err := s.db.Exec("DELETE FROM cart_items WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPut:
		var update struct {
			Quantity int `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if update.Quantity <= 0 {
			http.Error(w, "quantity > 0", http.StatusBadRequest)
			return
		}
		_, err := s.db.Exec("UPDATE cart_items SET quantity = ? WHERE id = ?", update.Quantity, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *CartService) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	type MergeReq struct {
		GuestKey string `json:"guest_key"`
		UserID   int64  `json:"user_id"`
	}
	var req MergeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.GuestKey == "" || req.UserID == 0 {
		http.Error(w, "guest_key and user_id required", http.StatusBadRequest)
		return
	}

	// Move guest items to user cart_key (user_id as str)
	userKey := strconv.FormatInt(req.UserID, 10)
	_, err := s.db.Exec("UPDATE cart_items SET cart_key = ?, updated_at = CURRENT_TIMESTAMP WHERE cart_key = ?", userKey, req.GuestKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "merged"})
}

func (s *CartService) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	cartKey := r.URL.Query().Get("cart_key")
	if cartKey == "" {
		var body struct {
			CartKey string `json:"cart_key"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		cartKey = body.CartKey
	}
	if cartKey == "" {
		http.Error(w, "cart_key required", http.StatusBadRequest)
		return
	}
	_, err := s.db.Exec("DELETE FROM cart_items WHERE cart_key = ?", cartKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
