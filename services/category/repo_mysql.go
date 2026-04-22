package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type MySQLRepo struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewMySQLRepo(dsn string) (*MySQLRepo, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	r := &MySQLRepo{db: db}
	if err := r.ensureSchema(); err != nil {
		return nil, err
	}
	return r, nil
}

// Close closes the underlying DB connection.
func (r *MySQLRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// RunMigrate connects to the given DSN and ensures the schema exists.
func RunMigrate(dsn string) error {
	repo, err := NewMySQLRepo(dsn)
	if err != nil {
		return err
	}
	return repo.Close()
}

func (r *MySQLRepo) SetRedisClient(c *redis.Client) { r.redisClient = c }

func (r *MySQLRepo) ensureSchema() error {
	q := `CREATE TABLE IF NOT EXISTS categories (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL,
        slug VARCHAR(255) NOT NULL UNIQUE,
        parent_id BIGINT NOT NULL DEFAULT 0,
        description TEXT,
        created_at DATETIME NOT NULL
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	_, err := r.db.Exec(q)
	return err
}

func (r *MySQLRepo) List() []Category {
	rows, err := r.db.Query("SELECT id, name, slug, parent_id, description, created_at FROM categories ORDER BY id ASC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Category
	for rows.Next() {
		var c Category
		var t time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description, &t); err != nil {
			continue
		}
		c.CreatedAt = t.UTC().Format(time.RFC3339)
		out = append(out, c)
	}
	return out
}

func (r *MySQLRepo) GetByID(id int64) (Category, bool) {
	var c Category
	var t time.Time
	err := r.db.QueryRow("SELECT id, name, slug, parent_id, description, created_at FROM categories WHERE id=?", id).Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description, &t)
	if err != nil {
		return Category{}, false
	}
	c.CreatedAt = t.UTC().Format(time.RFC3339)
	return c, true
}

func (r *MySQLRepo) Create(c Category) Category {
	now := time.Now().UTC()
	res, err := r.db.Exec("INSERT INTO categories (name, slug, parent_id, description, created_at) VALUES (?, ?, ?, ?, ?)", c.Name, c.Slug, c.ParentID, c.Description, now)
	if err != nil {
		return c
	}
	id, _ := res.LastInsertId()
	c.ID = id
	c.CreatedAt = now.Format(time.RFC3339)
	r.invalidateCacheLocked()
	return c
}

func (r *MySQLRepo) Update(c Category) bool {
	_, err := r.db.Exec("UPDATE categories SET name=?, slug=?, parent_id=?, description=? WHERE id=?", c.Name, c.Slug, c.ParentID, c.Description, c.ID)
	if err != nil {
		return false
	}
	r.invalidateCacheLocked()
	return true
}

func (r *MySQLRepo) Delete(id int64) bool {
	_, err := r.db.Exec("DELETE FROM categories WHERE id=?", id)
	if err != nil {
		return false
	}
	r.invalidateCacheLocked()
	return true
}

func (r *MySQLRepo) GetBySlug(slug string) (Category, bool) {
	var c Category
	var t time.Time
	err := r.db.QueryRow("SELECT id, name, slug, parent_id, description, created_at FROM categories WHERE slug=?", slug).Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description, &t)
	if err != nil {
		return Category{}, false
	}
	c.CreatedAt = t.UTC().Format(time.RFC3339)
	return c, true
}

func (r *MySQLRepo) BuildTree() []CategoryNode {
	// try cache first
	if r.redisClient != nil {
		if b, err := r.redisClient.Get(context.Background(), "categories:tree").Bytes(); err == nil {
			var nodes []CategoryNode
			if err := json.Unmarshal(b, &nodes); err == nil {
				return nodes
			}
		}
	}
	rows, err := r.db.Query("SELECT id, name, slug, parent_id, description, created_at FROM categories ORDER BY id ASC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	m := make(map[int64]*CategoryNode)
	var ids []int64
	for rows.Next() {
		var c Category
		var t time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.Description, &t); err != nil {
			continue
		}
		c.CreatedAt = t.UTC().Format(time.RFC3339)
		node := CategoryNode{Category: c}
		m[c.ID] = &node
		ids = append(ids, c.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var roots []CategoryNode
	for _, id := range ids {
		node := m[id]
		if node.ParentID == 0 {
			roots = append(roots, *node)
		} else if parent, ok := m[node.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}
	if r.redisClient != nil {
		if b, err := json.Marshal(roots); err == nil {
			_ = r.redisClient.Set(context.Background(), "categories:tree", b, time.Hour).Err()
		}
	}
	return roots
}

func (r *MySQLRepo) InvalidateCache() { r.invalidateCacheLocked() }

func (r *MySQLRepo) invalidateCacheLocked() {
	if r.redisClient != nil {
		_ = r.redisClient.Del(context.Background(), "categories:tree").Err()
	}
}
