package port

import (
	"github.com/your-org/go-minimall/internal/product/domain"
)

type Repository interface {
	ListCategories() ([]domain.CategoryRow, error)
	CreateCategory(c domain.CategoryRow) error
	CreateProduct(p domain.ProductRow) error
	ListProducts(onlyOnSale bool, limit, offset int) ([]domain.ProductRow, error)
	FindProductByID(id string) (domain.ProductRow, error)
	ListSkusByProductID(productID string) ([]domain.SkuRow, error)
	CreateSku(s domain.SkuRow) error
}
