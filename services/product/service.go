package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ProductService contains business logic for product operations
type ProductService struct {
	repo     *Repository
	cache    *redis.Client
	cacheTTL time.Duration
}

// NewProductService creates a new ProductService instance
func NewProductService(repo *Repository, cache *redis.Client, cacheTTL time.Duration) *ProductService {
	return &ProductService{repo: repo, cache: cache, cacheTTL: cacheTTL}
}

// GetProducts retrieves products with optional filters/pagination/sort
func (s *ProductService) GetProducts(filter map[string]string, minPrice, maxPrice *float64, page, pageSize int, sortBy, sortOrder string) ([]*Product, error) {
	if s.repo == nil {
		return GetSampleProducts(), nil
	}
	return s.repo.GetProducts(filter, minPrice, maxPrice, page, pageSize, sortBy, sortOrder)
}

// GetProductByID retrieves a product by ID with Redis cache (TTL)
func (s *ProductService) GetProductByID(id int64) (*Product, error) {
	if s.repo == nil {
		return s.findSampleProduct(id), nil
	}

	key := fmt.Sprintf("product:%d", id)
	ctx := context.Background()
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, key).Result(); err == nil && val != "" {
			var p Product
			if err := json.Unmarshal([]byte(val), &p); err == nil {
				return &p, nil
			}
		}
	}

	p, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil && p != nil {
		b, _ := json.Marshal(p)
		_ = s.cache.Set(ctx, key, string(b), s.cacheTTL).Err()
	}

	return p, nil
}

// GetProductsByIDs returns multiple products
func (s *ProductService) GetProductsByIDs(ids []int64) ([]*Product, error) {
	if s.repo == nil {
		// fallback to sample selection
		var out []*Product
		for _, id := range ids {
			if p := s.findSampleProduct(id); p != nil {
				out = append(out, p)
			}
		}
		return out, nil
	}
	return s.repo.GetProductsByIDs(ids)
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(req *CreateProductRequest) (*Product, error) {
	if req.Name == "" || req.Price <= 0 {
		return nil, fmt.Errorf("name and positive price required")
	}
	if s.repo == nil {
		return nil, fmt.Errorf("database connection not available")
	}
	return s.repo.CreateProduct(req)
}

// UpdateProduct updates an existing product and invalidates cache
func (s *ProductService) UpdateProduct(id int64, req *UpdateProductRequest) (*Product, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database connection not available")
	}
	p, err := s.repo.UpdateProduct(id, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Del(context.Background(), fmt.Sprintf("product:%d", id)).Err()
	}
	return p, nil
}

// DeleteProduct deletes and invalidates cache
func (s *ProductService) DeleteProduct(id int64) error {
	if s.repo == nil {
		return fmt.Errorf("database connection not available")
	}
	if err := s.repo.DeleteProduct(id); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Del(context.Background(), fmt.Sprintf("product:%d", id)).Err()
	}
	return nil
}

// SearchProducts searches products
func (s *ProductService) SearchProducts(query string) ([]*Product, error) {
	if query == "" {
		return []*Product{}, nil
	}
	if s.repo == nil {
		return s.searchSampleProducts(query), nil
	}
	return s.repo.SearchProducts(query)
}

// HandleStockUpdate applies a stock delta and invalidates cache
func (s *ProductService) HandleStockUpdate(productID int64, delta int64) error {
	if s.repo == nil {
		return fmt.Errorf("database connection not available")
	}
	if err := s.repo.UpdateStock(productID, delta); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Del(context.Background(), fmt.Sprintf("product:%d", productID)).Err()
	}
	return nil
}

// Private helper methods

func (s *ProductService) findSampleProduct(id int64) *Product {
	for _, p := range GetSampleProducts() {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (s *ProductService) searchSampleProducts(query string) []*Product {
	var result []*Product
	for _, p := range GetSampleProducts() {
		if stringsContains(p.Name, query) || stringsContains(p.Description, query) {
			result = append(result, p)
		}
	}
	return result
}

func stringsContains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
