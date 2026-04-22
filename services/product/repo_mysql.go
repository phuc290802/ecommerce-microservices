package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLRepo struct {
	db *sql.DB
}

func NewMySQLRepo(dsn string) (*MySQLRepo, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping failed: %w", err)
	}

	if err := initializeSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &MySQLRepo{db: db}, nil
}

func initializeSchema(db *sql.DB) error {
	// Check if emoji column exists
	var columnCount int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'products' AND COLUMN_NAME = 'emoji'").Scan(&columnCount)
	if err != nil {
		return fmt.Errorf("failed to check emoji column: %w", err)
	}

	if columnCount == 0 {
		_, err := db.Exec(`ALTER TABLE products ADD COLUMN emoji VARCHAR(10) DEFAULT '📦'`)
		if err != nil {
			return fmt.Errorf("failed to add emoji column: %w", err)
		}
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL,
        price DECIMAL(10,2) NOT NULL,
        category_id BIGINT NOT NULL DEFAULT 1,
        emoji VARCHAR(10) DEFAULT '📦',
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	if err != nil {
		return err
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		sampleProducts := []Product{
			{Name: "Basic T-shirt", Price: 19.99, CategoryID: 1, Emoji: "👕"},
			{Name: "Sneakers", Price: 59.99, CategoryID: 2, Emoji: "👟"},
			{Name: "Coffee Mug", Price: 9.99, CategoryID: 3, Emoji: "☕"},
		}

		for _, p := range sampleProducts {
			_, err := db.Exec(
				"INSERT INTO products (name, price, category_id, emoji) VALUES (?, ?, ?, ?)",
				p.Name, p.Price, p.CategoryID, p.Emoji,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *MySQLRepo) List() ([]Product, error) {
	rows, err := r.db.Query("SELECT id, name, price, category_id, emoji, created_at FROM products ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CategoryID, &p.Emoji, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (r *MySQLRepo) GetByID(id int64) (Product, error) {
	var p Product
	var createdAt time.Time

	err := r.db.QueryRow(
		"SELECT id, name, price, category_id, emoji, created_at FROM products WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.Price, &p.CategoryID, &p.Emoji, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, fmt.Errorf("product not found")
		}
		return Product{}, fmt.Errorf("failed to query product: %w", err)
	}

	p.CreatedAt = createdAt.Format(time.RFC3339)
	return p, nil
}

func (r *MySQLRepo) Create(product Product) (Product, error) {
	result, err := r.db.Exec(
		"INSERT INTO products (name, price, category_id, emoji) VALUES (?, ?, ?, ?)",
		product.Name, product.Price, product.CategoryID, product.Emoji,
	)
	if err != nil {
		return Product{}, fmt.Errorf("failed to create product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Product{}, fmt.Errorf("failed to get last insert id: %w", err)
	}

	product.ID = id
	product.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	return product, nil
}

func (r *MySQLRepo) Update(product Product) (bool, error) {
	result, err := r.db.Exec(
		"UPDATE products SET name = ?, price = ?, category_id = ?, emoji = ? WHERE id = ?",
		product.Name, product.Price, product.CategoryID, product.Emoji, product.ID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *MySQLRepo) Delete(id int64) (bool, error) {
	result, err := r.db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *MySQLRepo) GetByCategory(categoryID int64) ([]Product, error) {
	rows, err := r.db.Query(
		"SELECT id, name, price, category_id, created_at FROM products WHERE category_id = ? ORDER BY id",
		categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by category: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CategoryID, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}
