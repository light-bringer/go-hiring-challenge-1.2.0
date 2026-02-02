package catalog

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/app/dto"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CatalogService struct {
	productRepo models.ProductRepository
}

func NewCatalogService(repo models.ProductRepository) *CatalogService {
	return &CatalogService{productRepo: repo}
}

func (s *CatalogService) ListProducts(ctx context.Context, opts models.ProductListOptions) ([]ProductDTO, int64, error) {
	products, total, err := s.productRepo.GetAllProducts(ctx, opts)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]ProductDTO, len(products))
	for i, p := range products {
		dtos[i] = toProductDTO(p)
	}

	return dtos, total, nil
}

func (s *CatalogService) GetProductWithVariants(ctx context.Context, code string) (*ProductDetailDTO, error) {
	product, err := s.productRepo.GetProductByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return toProductDetailDTO(product), nil
}

// Helper: Map Product to DTO
func toProductDTO(p models.Product) ProductDTO {
	result := ProductDTO{
		Code:  p.Code,
		Price: DecimalPrice{p.Price},
	}

	if p.Category != nil {
		result.Category = &dto.CategoryDTO{
			Code: p.Category.Code,
			Name: p.Category.Name,
		}
	}

	return result
}

// Helper: Map Product to Detail DTO with variants
func toProductDetailDTO(p *models.Product) *ProductDetailDTO {
	result := &ProductDetailDTO{
		Code:  p.Code,
		Price: DecimalPrice{p.Price},
	}

	if p.Category != nil {
		result.Category = &dto.CategoryDTO{
			Code: p.Category.Code,
			Name: p.Category.Name,
		}
	}

	// Map variants with price inheritance
	result.Variants = make([]VariantDTO, len(p.Variants))
	for i, v := range p.Variants {
		price := p.Price // Default to product price

		// Use variant price if set (not NULL)
		if v.Price != nil && !v.Price.IsZero() {
			price = *v.Price
		}

		result.Variants[i] = VariantDTO{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: DecimalPrice{price},
		}
	}

	return result
}
