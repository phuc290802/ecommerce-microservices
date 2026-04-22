package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	dsn := getEnv("DB_DSN", "user:password@tcp(mysql:3306)/ecommerce?parseTime=true")

	var repo ProductRepo
	var repoErr error
	const maxRetries = 10

	for i := 0; i < maxRetries; i++ {
		repo, repoErr = NewMySQLRepo(dsn)
		if repoErr == nil {
			break
		}
		log.Printf("failed to connect to db (attempt %d/%d): %v", i+1, maxRetries, repoErr)
		time.Sleep(2 * time.Second)
	}

	if repoErr != nil {
		log.Fatalf("failed to connect to db: %v", repoErr)
	}

	svc := NewProductService(repo)
	handlers := NewHandlers(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.handleHealth)
	mux.HandleFunc("/products", handlers.handleProducts)
	mux.HandleFunc("/products/", handlers.handleProductByID)

	addr := ":8081"
	log.Printf("product service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
