package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Stock struct {
	ProductID int64 `json:"product_id"`
	Available bool  `json:"available"`
	Quantity  int   `json:"quantity"`
}

var stockData = []Stock{
	{ProductID: 1, Available: true, Quantity: 24},
	{ProductID: 2, Available: true, Quantity: 12},
	{ProductID: 3, Available: false, Quantity: 0},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/stock", handleStock)

	addr := ":8087"
	log.Printf("stock service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStock(w http.ResponseWriter, r *http.Request) {
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

	for _, item := range stockData {
		if item.ProductID == productID {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(item)
			return
		}
	}

	http.Error(w, "stock info not found", http.StatusNotFound)
}
