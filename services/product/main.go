package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	CategoryID int64   `json:"category_id"`
	CreatedAt  string  `json:"created_at"`
}

var sampleProducts = []Product{
	{ID: 1, Name: "Basic T-shirt", Price: 19.99, CategoryID: 1, CreatedAt: "2026-04-18T09:20:00Z"},
	{ID: 2, Name: "Sneakers", Price: 59.99, CategoryID: 2, CreatedAt: "2026-04-17T10:10:00Z"},
	{ID: 3, Name: "Coffee Mug", Price: 9.99, CategoryID: 3, CreatedAt: "2026-04-16T15:05:00Z"},
}

func main() {
	dsn := os.Getenv("DB_DSN")
	var db *sql.DB
	var err error
	if dsn != "" {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("failed to open mysql: %v", err)
		} else {
			db.SetConnMaxLifetime(3 * time.Minute)
			db.SetMaxOpenConns(5)
			db.SetMaxIdleConns(2)
			err = db.Ping()
			if err != nil {
				log.Printf("mysql ping failed: %v", err)
				db = nil
			} else {
				err = initializeSchema(db)
				if err != nil {
					log.Printf("failed to initialize schema: %v", err)
					db = nil
				}
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/products", handleProducts(db))
	mux.HandleFunc("/products/", handleProductByID(db))

	addr := ":8081"
	log.Printf("product service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func initializeSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL,
        price DECIMAL(10,2) NOT NULL,
        category_id BIGINT NOT NULL DEFAULT 1,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		return err
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		for _, p := range sampleProducts {
			_, err := db.Exec("INSERT INTO products (name, price, category_id) VALUES (?, ?, ?)", p.Name, p.Price, p.CategoryID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var products []Product
		if db != nil {
			rows, err := db.Query("SELECT id, name, price, category_id, created_at FROM products ORDER BY id")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var p Product
					var createdAt time.Time
					_ = rows.Scan(&p.ID, &p.Name, &p.Price, &p.CategoryID, &createdAt)
					p.CreatedAt = createdAt.Format(time.RFC3339)
					products = append(products, p)
				}
			}
		}
		if len(products) == 0 {
			products = sampleProducts
		}
		_ = json.NewEncoder(w).Encode(products)
	}
}

func handleProductByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/products/")
		if idStr == "" {
			http.Error(w, "missing product id", http.StatusBadRequest)
			return
		}
		var product Product
		if db != nil {
			err := db.QueryRow("SELECT id, name, price, category_id, created_at FROM products WHERE id = ?", idStr).Scan(&product.ID, &product.Name, &product.Price, &product.CategoryID, &product.CreatedAt)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(product)
				return
			}
		}
		for _, p := range sampleProducts {
			if fmt.Sprintf("%d", p.ID) == idStr {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(p)
				return
			}
		}
		http.Error(w, "product not found", http.StatusNotFound)
	}
}
