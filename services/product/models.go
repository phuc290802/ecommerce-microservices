package main

import (
	"encoding/json"
	"time"
)

// Product represents a product in the system
type Product struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int64     `json:"stock_quantity"`
	SKU           string    `json:"sku"`
	Images        []string  `json:"images"`
	Attributes    string    `json:"attributes"` // JSON string
	CategoryID    int64     `json:"category_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateProductRequest represents create product payload
type CreateProductRequest struct {
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Price         float64           `json:"price"`
	StockQuantity int64             `json:"stock_quantity"`
	SKU           string            `json:"sku,omitempty"`
	Images        []string          `json:"images,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	CategoryID    int64             `json:"category_id"`
}

// UpdateProductRequest represents update product payload
type UpdateProductRequest struct {
	Name          string            `json:"name,omitempty"`
	Description   string            `json:"description,omitempty"`
	Price         float64           `json:"price,omitempty"`
	StockQuantity *int64            `json:"stock_quantity,omitempty"`
	SKU           string            `json:"sku,omitempty"`
	Images        []string          `json:"images,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	CategoryID    int64             `json:"category_id,omitempty"`
}

// ProductResponse represents the product in API response
type ProductResponse struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Price         float64           `json:"price"`
	StockQuantity int64             `json:"stock_quantity"`
	SKU           string            `json:"sku"`
	Images        []string          `json:"images"`
	Attributes    map[string]string `json:"attributes"`
	CategoryID    int64             `json:"category_id"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() *ProductResponse {
	var attrs map[string]string
	if p.Attributes != "" {
		_ = json.Unmarshal([]byte(p.Attributes), &attrs)
	} else {
		attrs = map[string]string{}
	}

	return &ProductResponse{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		StockQuantity: p.StockQuantity,
		SKU:           p.SKU,
		Images:        p.Images,
		Attributes:    attrs,
		CategoryID:    p.CategoryID,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

// GetSampleProducts returns sample products for fallback
func GetSampleProducts() []*Product {
	attrs, _ := json.Marshal(map[string]string{"color": "red", "size": "M"})
	now := time.Now()
	return []*Product{
		{
			ID:            1,
			Name:          "Basic T-shirt",
			Description:   "Comfortable cotton t-shirt",
			Price:         19.99,
			StockQuantity: 100,
			SKU:           "TSHIRT-001",
			Images:        []string{"/uploads/tshirt.jpg"},
			Attributes:    string(attrs),
			CategoryID:    1,
			CreatedAt:     now.Add(-72 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            2,
			Name:          "Sneakers",
			Description:   "Running sneakers",
			Price:         59.99,
			StockQuantity: 50,
			SKU:           "SNK-001",
			Images:        []string{"/uploads/sneakers.jpg"},
			Attributes:    string(attrs),
			CategoryID:    2,
			CreatedAt:     now.Add(-96 * time.Hour),
			UpdatedAt:     now.Add(-24 * time.Hour),
		},
	}
}
