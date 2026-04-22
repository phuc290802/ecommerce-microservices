package main

import (
	"fmt"
)

// SeedInitialCategories inserts sample categories if they do not already exist.
func SeedInitialCategories(repo CategoryRepo) (int, error) {
	initial := []Category{
		{Name: "Clothing", Slug: "clothing", ParentID: 0, Description: "Apparel and clothing"},
		{Name: "Shoes", Slug: "shoes", ParentID: 1, Description: "Footwear"},
		{Name: "Accessories", Slug: "accessories", ParentID: 1, Description: "Accessories"},
	}
	inserted := 0
	for _, c := range initial {
		if _, ok := repo.GetBySlug(c.Slug); ok {
			continue
		}
		created := repo.Create(c)
		if created.ID != 0 {
			inserted++
		}
	}
	return inserted, nil
}

// helper for CLI call
func RunSeedCLI(repo CategoryRepo) error {
	n, err := SeedInitialCategories(repo)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted %d categories\n", n)
	return nil
}
