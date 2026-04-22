package main

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CategoryRepo interface {
	List() []Category
	GetByID(int64) (Category, bool)
	Create(Category) Category
	Update(Category) bool
	Delete(int64) bool
	GetBySlug(string) (Category, bool)
	BuildTree() []CategoryNode
	InvalidateCache()
	SetRedisClient(*redis.Client)
}

type InMemoryRepo struct {
	mu          sync.RWMutex
	categories  []Category
	redisClient *redis.Client
}

func NewInMemoryRepo(initial []Category) *InMemoryRepo {
	cats := make([]Category, len(initial))
	copy(cats, initial)
	return &InMemoryRepo{categories: cats}
}

func (r *InMemoryRepo) SetRedisClient(c *redis.Client) {
	r.redisClient = c
}

func (r *InMemoryRepo) List() []Category {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Category, len(r.categories))
	copy(out, r.categories)
	return out
}

func (r *InMemoryRepo) GetByID(id int64) (Category, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.categories {
		if c.ID == id {
			return c, true
		}
	}
	return Category{}, false
}

func (r *InMemoryRepo) Create(c Category) Category {
	r.mu.Lock()
	defer r.mu.Unlock()
	var maxID int64
	for _, x := range r.categories {
		if x.ID > maxID {
			maxID = x.ID
		}
	}
	c.ID = maxID + 1
	r.categories = append(r.categories, c)
	r.invalidateCacheLocked()
	return c
}

func (r *InMemoryRepo) Update(c Category) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.categories {
		if r.categories[i].ID == c.ID {
			// update fields if provided
			if c.Name != "" {
				r.categories[i].Name = c.Name
			}
			if c.Slug != "" {
				r.categories[i].Slug = c.Slug
			}
			r.categories[i].ParentID = c.ParentID
			r.categories[i].Description = c.Description
			r.invalidateCacheLocked()
			return true
		}
	}
	return false
}

func (r *InMemoryRepo) Delete(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := -1
	for i, c := range r.categories {
		if c.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	r.categories = append(r.categories[:idx], r.categories[idx+1:]...)
	r.invalidateCacheLocked()
	return true
}

func (r *InMemoryRepo) GetBySlug(slug string) (Category, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.categories {
		if c.Slug == slug {
			return c, true
		}
	}
	return Category{}, false
}

func (r *InMemoryRepo) BuildTree() []CategoryNode {
	// try cache first
	if r.redisClient != nil {
		if b, err := r.redisClient.Get(context.Background(), "categories:tree").Bytes(); err == nil {
			var nodes []CategoryNode
			if err := json.Unmarshal(b, &nodes); err == nil {
				return nodes
			}
		}
	}
	r.mu.RLock()
	// create map
	m := make(map[int64]*CategoryNode)
	var ids []int64
	for _, c := range r.categories {
		node := CategoryNode{Category: c}
		m[c.ID] = &node
		ids = append(ids, c.ID)
	}
	r.mu.RUnlock()
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

	// cache
	if r.redisClient != nil {
		if b, err := json.Marshal(roots); err == nil {
			_ = r.redisClient.Set(context.Background(), "categories:tree", b, time.Hour).Err()
		}
	}
	return roots
}

func (r *InMemoryRepo) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invalidateCacheLocked()
}

func (r *InMemoryRepo) invalidateCacheLocked() {
	if r.redisClient != nil {
		_ = r.redisClient.Del(context.Background(), "categories:tree").Err()
	}
}
