package main

type Product struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	CategoryID int64   `json:"category_id"`
	Emoji      string  `json:"emoji"`
	CreatedAt  string  `json:"created_at"`
}
