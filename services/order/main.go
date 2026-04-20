package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Order struct {
	ID        int64   `json:"id"`
	Product   string  `json:"product"`
	Quantity  int     `json:"quantity"`
	TotalCost float64 `json:"total_cost"`
}

var sampleOrders = []Order{
	{ID: 101, Product: "Basic T-shirt", Quantity: 2, TotalCost: 39.98},
	{ID: 102, Product: "Sneakers", Quantity: 1, TotalCost: 59.99},
	{ID: 103, Product: "Coffee Mug", Quantity: 3, TotalCost: 29.97},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/orders", handleOrders)
	mux.HandleFunc("/orders/", handleOrderByID)

	addr := ":8082"
	log.Printf("order service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sampleOrders)
}

func handleOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/orders/"):]
	for _, o := range sampleOrders {
		if fmt.Sprintf("%d", o.ID) == idStr {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(o)
			return
		}
	}
	http.Error(w, "order not found", http.StatusNotFound)
}
