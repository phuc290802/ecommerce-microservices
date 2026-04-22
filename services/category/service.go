package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CategoryService struct {
	repo       CategoryRepo
	httpClient *http.Client
}

func NewCategoryService(repo CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (s *CategoryService) List() []Category {
	return s.repo.List()
}

func (s *CategoryService) Create(payload Category) Category {
	if payload.CreatedAt == "" {
		payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return s.repo.Create(payload)
}

func (s *CategoryService) GetByID(id int64) (Category, bool)      { return s.repo.GetByID(id) }
func (s *CategoryService) Update(c Category) bool                 { return s.repo.Update(c) }
func (s *CategoryService) Delete(id int64) bool                   { return s.repo.Delete(id) }
func (s *CategoryService) GetBySlug(slug string) (Category, bool) { return s.repo.GetBySlug(slug) }
func (s *CategoryService) GetTree() []CategoryNode                { return s.repo.BuildTree() }
func (s *CategoryService) RebuildTree() []CategoryNode {
	s.repo.InvalidateCache()
	return s.repo.BuildTree()
}

func (s *CategoryService) GetProductsByCategory(categoryID int64) ([]interface{}, error) {
	productURL := fmt.Sprintf("%s/products", getEnv("PRODUCT_SERVICE_URL", "http://product-service:8081"))
	req, err := http.NewRequest(http.MethodGet, productURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("product service error: %d", resp.StatusCode)
	}
	var products []struct {
		ID         int64   `json:"id"`
		Name       string  `json:"name"`
		Price      float64 `json:"price"`
		CategoryID int64   `json:"category_id"`
		CreatedAt  string  `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}
	var filtered []interface{}
	for _, p := range products {
		if p.CategoryID == categoryID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}
