package main

type ProductService struct {
	repo ProductRepo
}

func NewProductService(repo ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List() ([]Product, error) {
	return s.repo.List()
}

func (s *ProductService) GetByID(id int64) (Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) Create(product Product) (Product, error) {
	return s.repo.Create(product)
}

func (s *ProductService) Update(product Product) (bool, error) {
	return s.repo.Update(product)
}

func (s *ProductService) Delete(id int64) (bool, error) {
	return s.repo.Delete(id)
}

func (s *ProductService) GetByCategory(categoryID int64) ([]Product, error) {
	return s.repo.GetByCategory(categoryID)
}
