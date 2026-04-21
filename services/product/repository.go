package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Repository handles all database operations for product service
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InitSchema creates the products table if it doesn't exist
func (r *Repository) InitSchema() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		stock_quantity BIGINT NOT NULL DEFAULT 0,
		sku VARCHAR(100),
		images JSON,
		attributes JSON,
		category_id BIGINT NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		return err
	}

	// Seed sample products if table is empty
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		for _, p := range GetSampleProducts() {
			imagesJson, _ := json.Marshal(p.Images)
			_, err := r.db.Exec(
				"INSERT INTO products (name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				p.Name, p.Description, p.Price, p.StockQuantity, p.SKU, imagesJson, p.Attributes, p.CategoryID, p.CreatedAt, p.UpdatedAt,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// helper to scan row into Product
func scanProductRow(rows ScannerRow) (*Product, error) {
	var (
		id          int64
		name        string
		description sql.NullString
		price       float64
		stock       int64
		sku         sql.NullString
		imagesRaw   sql.NullString
		attributes  sql.NullString
		categoryID  int64
		createdAt   time.Time
		updatedAt   time.Time
	)

	if err := rows.Scan(&id, &name, &description, &price, &stock, &sku, &imagesRaw, &attributes, &categoryID, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	var images []string
	if imagesRaw.Valid && imagesRaw.String != "" {
		_ = json.Unmarshal([]byte(imagesRaw.String), &images)
	}

	prod := &Product{
		ID:            id,
		Name:          name,
		Description:   description.String,
		Price:         price,
		StockQuantity: stock,
		SKU:           sku.String,
		Images:        images,
		Attributes:    attributes.String,
		CategoryID:    categoryID,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return prod, nil
}

// ScannerRow abstracts *sql.Row and *sql.Rows Scan signature
type ScannerRow interface {
	Scan(dest ...interface{}) error
}

// GetProductByID retrieves a product by ID
func (r *Repository) GetProductByID(id int64) (*Product, error) {
	row := r.db.QueryRow(
		"SELECT id, name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at FROM products WHERE id = ?",
		id,
	)

	return scanProductRow(row)
}

// GetProducts retrieves products with optional filters, pagination and sorting
func (r *Repository) GetProducts(filter map[string]string, minPrice, maxPrice *float64, page, pageSize int, sortBy, sortOrder string) ([]*Product, error) {
	var where []string
	var args []interface{}

	if q, ok := filter["name"]; ok && q != "" {
		where = append(where, "(name LIKE ? OR description LIKE ?)")
		term := "%" + q + "%"
		args = append(args, term, term)
	}
	if cat, ok := filter["category_id"]; ok && cat != "" {
		where = append(where, "category_id = ?")
		args = append(args, cat)
	}
	if attr, ok := filter["attribute"]; ok && attr != "" {
		// simple JSON contains search: attributes LIKE %"key":"value"%
		where = append(where, "attributes LIKE ?")
		args = append(args, "%"+attr+"%")
	}
	if minPrice != nil {
		where = append(where, "price >= ?")
		args = append(args, *minPrice)
	}
	if maxPrice != nil {
		where = append(where, "price <= ?")
		args = append(args, *maxPrice)
	}

	query := "SELECT id, name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at FROM products"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// sorting
	allowedSort := map[string]bool{"price": true, "name": true, "created_at": true}
	if !allowedSort[sortBy] {
		sortBy = "created_at"
	}
	if strings.ToLower(sortOrder) != "asc" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// pagination
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if page <= 0 {
		offset = 0
		page = 1
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProductsByIDs returns products for a slice of IDs
func (r *Repository) GetProductsByIDs(ids []int64) ([]*Product, error) {
	if len(ids) == 0 {
		return []*Product{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT id, name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at FROM products WHERE id IN (%s)", strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// CreateProduct creates a new product
func (r *Repository) CreateProduct(req *CreateProductRequest) (*Product, error) {
	imagesJson, _ := json.Marshal(req.Images)
	attrsJson, _ := json.Marshal(req.Attributes)
	res, err := r.db.Exec(
		"INSERT INTO products (name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Description, req.Price, req.StockQuantity, req.SKU, imagesJson, string(attrsJson), req.CategoryID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetProductByID(id)
}

// UpdateProduct updates an existing product
func (r *Repository) UpdateProduct(id int64, req *UpdateProductRequest) (*Product, error) {
	// fetch existing
	existing, err := r.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	// merge
	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}
	description := existing.Description
	if req.Description != "" {
		description = req.Description
	}
	price := existing.Price
	if req.Price > 0 {
		price = req.Price
	}
	stock := existing.StockQuantity
	if req.StockQuantity != nil {
		stock = *req.StockQuantity
	}
	sku := existing.SKU
	if req.SKU != "" {
		sku = req.SKU
	}
	images := existing.Images
	if len(req.Images) > 0 {
		images = req.Images
	}
	attrsMap := map[string]string{}
	if existing.Attributes != "" {
		_ = json.Unmarshal([]byte(existing.Attributes), &attrsMap)
	}
	for k, v := range req.Attributes {
		attrsMap[k] = v
	}
	attrsJson, _ := json.Marshal(attrsMap)
	categoryID := existing.CategoryID
	if req.CategoryID != 0 {
		categoryID = req.CategoryID
	}

	imagesJson, _ := json.Marshal(images)
	_, err = r.db.Exec(
		"UPDATE products SET name = ?, description = ?, price = ?, stock_quantity = ?, sku = ?, images = ?, attributes = ?, category_id = ?, updated_at = ? WHERE id = ?",
		name, description, price, stock, sku, imagesJson, string(attrsJson), categoryID, time.Now(), id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetProductByID(id)
}

// DeleteProduct deletes a product
func (r *Repository) DeleteProduct(id int64) error {
	_, err := r.db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

// UpdateStock updates product stock quantity (set or delta)
func (r *Repository) UpdateStock(id int64, delta int64) error {
	_, err := r.db.Exec("UPDATE products SET stock_quantity = stock_quantity + ?, updated_at = ? WHERE id = ?", delta, time.Now(), id)
	return err
}

// SearchProducts searches for products by name/description/attributes
func (r *Repository) SearchProducts(query string) ([]*Product, error) {
	if strings.TrimSpace(query) == "" {
		return []*Product{}, nil
	}
	term := "%" + query + "%"
	rows, err := r.db.Query("SELECT id, name, description, price, stock_quantity, sku, images, attributes, category_id, created_at, updated_at FROM products WHERE name LIKE ? OR description LIKE ? OR attributes LIKE ? ORDER BY id", term, term, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
