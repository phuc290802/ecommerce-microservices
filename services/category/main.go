package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var categories = []Category{
	{ID: 1, Name: "Clothing", Description: "Quần áo và phụ kiện"},
	{ID: 2, Name: "Footwear", Description: "Giày dép và đồ tập"},
	{ID: 3, Name: "Home", Description: "Đồ dùng gia đình"},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/categories", handleCategories)
	mux.HandleFunc("/categories/", handleCategoryByID)

	addr := ":8085"
	log.Printf("category service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

func handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/categories/"):]
	for _, c := range categories {
		if fmt.Sprintf("%d", c.ID) == idStr {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(c)
			return
		}
	}
	http.Error(w, "category not found", http.StatusNotFound)
}
