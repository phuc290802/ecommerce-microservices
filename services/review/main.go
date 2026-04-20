package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Review struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Author    string `json:"author"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

var reviews = []Review{
	{ID: 201, ProductID: 1, Author: "Lan", Rating: 5, Comment: "Chất lượng tốt", CreatedAt: "2026-04-20T14:30:00Z"},
	{ID: 202, ProductID: 1, Author: "Minh", Rating: 4, Comment: "Nhỏ gọn, ưng ý", CreatedAt: "2026-04-18T11:12:00Z"},
	{ID: 203, ProductID: 2, Author: "Hồ", Rating: 5, Comment: "Thoải mái, bền", CreatedAt: "2026-04-19T08:05:00Z"},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/reviews", handleReviews)

	addr := ":8086"
	log.Printf("review service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		http.Error(w, "product_id required", http.StatusBadRequest)
		return
	}
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product_id", http.StatusBadRequest)
		return
	}

	var filtered []Review
	for _, review := range reviews {
		if review.ProductID == productID {
			filtered = append(filtered, review)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filtered)
}
