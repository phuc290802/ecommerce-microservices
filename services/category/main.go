package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := getEnv("REDIS_ADDR", "redis:6379")
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("warning: redis ping failed: %v (cache disabled)", err)
		redisClient = nil
	}

	// database-backed repo (replace in-memory seed)
	// default DSN matches docker-compose credentials: user/password on host `mysql`
	dsn := getEnv("DB_DSN", "user:password@tcp(mysql:3306)/ecommerce?parseTime=true")

	// retry loop: wait for DB to be ready (useful on container startup)
	var repoCategoryRepoErr error
	var repoRetryAttempts = 10
	var repoErr error
	var repoIface *MySQLRepo
	for i := 0; i < repoRetryAttempts; i++ {
		repoIface, repoErr = NewMySQLRepo(dsn)
		if repoErr == nil {
			break
		}
		log.Printf("failed to connect to db (attempt %d/%d): %v", i+1, repoRetryAttempts, repoErr)
		time.Sleep(2 * time.Second)
	}
	repoCategoryRepoErr = repoErr
	if repoCategoryRepoErr != nil {
		log.Fatalf("failed to connect to db: %v", repoCategoryRepoErr)
	}
	// assign to CategoryRepo interface
	var repo CategoryRepo = repoIface
	repo.SetRedisClient(redisClient)

	// CLI: `migrate` ensures DB schema, `seed` inserts initial categories
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			// prefer DB_DSN env if set, otherwise default to host-mapped port 3307
			migrateDSN := getEnv("DB_DSN", "root:root@tcp(localhost:3307)/ecommerce?parseTime=true")
			if err := RunMigrate(migrateDSN); err != nil {
				log.Fatalf("migrate failed: %v", err)
			}
			log.Printf("migration completed against %s", migrateDSN)
			return
		case "seed":
			if err := RunSeedCLI(repo); err != nil {
				log.Fatalf("seed failed: %v", err)
			}
			return
		}
	}
	svc := NewCategoryService(repo)
	handlers := NewHandlers(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.handleHealth)
	mux.HandleFunc("/categories", handlers.handleCategories)
	mux.HandleFunc("/categories/tree", handlers.handleCategoryTree)
	mux.HandleFunc("/categories/rebuild", handlers.handleRebuildTree)
	mux.HandleFunc("/categories/slug/", handlers.handleCategoryBySlug)
	mux.HandleFunc("/categories/", handlers.handleCategoryByID)

	addr := getEnv("PORT", ":8085")
	log.Printf("category service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
