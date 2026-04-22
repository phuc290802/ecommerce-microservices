package main

type ProductRepo interface {
	List() ([]Product, error)
	GetByID(id int64) (Product, error)
	Create(product Product) (Product, error)
	Update(product Product) (bool, error)
	Delete(id int64) (bool, error)
	GetByCategory(categoryID int64) ([]Product, error)
}
