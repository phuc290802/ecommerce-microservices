package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/redis/go-redis/v9"
)

const (
	defaultPort   = "8083"
	defaultRedis  = "redis:6379"
	summaryTTL    = 30 * time.Second
	graphqlMaxAge = 30 * time.Second
)

type ProductResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	CategoryID int64   `json:"category_id"`
	CreatedAt  string  `json:"created_at"`
}

type CategoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ReviewResponse struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Author    string `json:"author"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type StockResponse struct {
	ProductID int64 `json:"product_id"`
	Available bool  `json:"available"`
	Quantity  int   `json:"quantity"`
}

type SummaryItem struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	UnitPrice   float64 `json:"unit_price"`
	CategoryID  int64   `json:"category_id"`
	CreatedDate string  `json:"created_date"`
}

type SummaryCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Info string `json:"info"`
}

type SummaryReview struct {
	Reviewer string `json:"reviewer"`
	Stars    int    `json:"stars"`
	Text     string `json:"text"`
	Date     string `json:"date"`
}

type SummaryStock struct {
	Available     bool   `json:"available"`
	Quantity      int    `json:"quantity"`
	StatusMessage string `json:"status_message"`
}

type SummaryResponse struct {
	Item         SummaryItem     `json:"item"`
	Category     SummaryCategory `json:"category"`
	Reviews      []SummaryReview `json:"reviews"`
	StockStatus  SummaryStock    `json:"stock_status"`
	AggregatedAt string          `json:"aggregated_at"`
}

type summaryResolver struct {
	productURL  string
	categoryURL string
	reviewURL   string
	stockURL    string
	cache       *redis.Client
}

var (
	redisClient   *redis.Client
	graphqlSchema graphql.Schema
	client        = &http.Client{Timeout: 12 * time.Second}
)

func main() {
	redisAddr := getEnv("REDIS_ADDR", defaultRedis)
	port := getEnv("PORT", defaultPort)

	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

	if err := graphqlInit(); err != nil {
		log.Fatalf("graphql init: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/summary", handleSummary)
	mux.HandleFunc("/graphql", handleGraphQL)

	addr := ":" + port
	log.Printf("BFF service starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func graphqlInit() error {
	summaryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Summary",
		Fields: graphql.Fields{
			"item":          &graphql.Field{Type: itemType},
			"category":      &graphql.Field{Type: categoryType},
			"reviews":       &graphql.Field{Type: graphql.NewList(reviewType)},
			"stock_status":  &graphql.Field{Type: stockType},
			"aggregated_at": &graphql.Field{Type: graphql.String},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"summary": &graphql.Field{
				Type: summaryType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					productID := strconv.Itoa(p.Args["productId"].(int))
					return aggregateSummary(productID)
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: rootQuery})
	if err != nil {
		return err
	}
	graphqlSchema = schema
	return nil
}

var itemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Item",
	Fields: graphql.Fields{
		"id":           &graphql.Field{Type: graphql.Int},
		"title":        &graphql.Field{Type: graphql.String},
		"unit_price":   &graphql.Field{Type: graphql.Float},
		"category_id":  &graphql.Field{Type: graphql.Int},
		"created_date": &graphql.Field{Type: graphql.String},
	},
})

var categoryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Category",
	Fields: graphql.Fields{
		"id":   &graphql.Field{Type: graphql.Int},
		"name": &graphql.Field{Type: graphql.String},
		"info": &graphql.Field{Type: graphql.String},
	},
})

var reviewType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Review",
	Fields: graphql.Fields{
		"reviewer": &graphql.Field{Type: graphql.String},
		"stars":    &graphql.Field{Type: graphql.Int},
		"text":     &graphql.Field{Type: graphql.String},
		"date":     &graphql.Field{Type: graphql.String},
	},
})

var stockType = graphql.NewObject(graphql.ObjectConfig{
	Name: "StockStatus",
	Fields: graphql.Fields{
		"available":      &graphql.Field{Type: graphql.Boolean},
		"quantity":       &graphql.Field{Type: graphql.Int},
		"status_message": &graphql.Field{Type: graphql.String},
	},
})

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, "product_id required", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("bff:summary:%s", productID)
	if cached, err := redisClient.Get(context.Background(), cacheKey).Bytes(); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	summary, err := aggregateSummary(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		http.Error(w, "failed to marshal summary", http.StatusInternalServerError)
		return
	}
	_ = redisClient.Set(context.Background(), cacheKey, payload, summaryTTL).Err()

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func handleGraphQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result := graphql.Do(graphql.Params{Schema: graphqlSchema, RequestString: body.Query})
	if len(result.Errors) > 0 {
		http.Error(w, result.Errors[0].Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func aggregateSummary(productID string) (*SummaryResponse, error) {

	productURL := fmt.Sprintf("%s/products/%s", getEnv("PRODUCT_SERVICE_URL", "http://product-service:8081"), productID)
	product := &ProductResponse{}
	if err := fetchJSON(productURL, product); err != nil {
		return nil, fmt.Errorf("product service failed: %w", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 3)

	var category CategoryResponse
	var reviews []ReviewResponse
	var stock StockResponse

	wg.Add(3)

	go func() {
		defer wg.Done()
		categoryURL := fmt.Sprintf("%s/categories/%d", getEnv("CATEGORY_SERVICE_URL", "http://category-service:8085"), product.CategoryID)
		if err := fetchJSON(categoryURL, &category); err != nil {
			errs <- fmt.Errorf("category service failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		reviewURL := fmt.Sprintf("%s/reviews?product_id=%s", getEnv("REVIEW_SERVICE_URL", "http://review-service:8086"), productID)
		if err := fetchJSON(reviewURL, &reviews); err != nil {
			errs <- fmt.Errorf("review service failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		stockURL := fmt.Sprintf("%s/stock?product_id=%s", getEnv("STOCK_SERVICE_URL", "http://stock-service:8087"), productID)
		if err := fetchJSON(stockURL, &stock); err != nil {
			errs <- fmt.Errorf("stock service failed: %w", err)
		}
	}()

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return transformSummary(product, &category, reviews, &stock), nil
}

func transformSummary(product *ProductResponse, category *CategoryResponse, reviews []ReviewResponse, stock *StockResponse) *SummaryResponse {
	summaryReviews := make([]SummaryReview, 0, len(reviews))
	for _, review := range reviews {
		summaryReviews = append(summaryReviews, SummaryReview{
			Reviewer: review.Author,
			Stars:    review.Rating,
			Text:     review.Comment,
			Date:     formatDate(review.CreatedAt),
		})
	}

	statusMessage := "available"
	if !stock.Available {
		statusMessage = "out of stock"
	}

	return &SummaryResponse{
		Item: SummaryItem{
			ID:          product.ID,
			Title:       product.Name,
			UnitPrice:   product.Price,
			CategoryID:  product.CategoryID,
			CreatedDate: formatDate(product.CreatedAt),
		},
		Category: SummaryCategory{
			ID:   category.ID,
			Name: category.Name,
			Info: category.Description,
		},
		Reviews: summaryReviews,
		StockStatus: SummaryStock{
			Available:     stock.Available,
			Quantity:      stock.Quantity,
			StatusMessage: statusMessage,
		},
		AggregatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func fetchJSON(url string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-ID", generateRequestID())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func formatDate(value string) string {
	if value == "" {
		return ""
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts.Format("2006-01-02")
	}
	return value
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func generateRequestID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(999999999999))
	if err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req-%d", n.Int64())
}
