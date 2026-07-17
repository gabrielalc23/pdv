package catalog

import "time"

type ListCatalogInput struct {
	Search        string
	Page          *int
	PageSize      *int
	ActiveOnly    bool
	InStockOnly   bool
	ActiveOnlySet bool
	CategoryID    string
}

type CatalogProductResponse struct {
	ID           string    `json:"id"`
	SKU          string    `json:"sku"`
	Barcode      *string   `json:"barcode"`
	Name         string    `json:"name"`
	CategoryID   *string   `json:"categoryId"`
	CategoryName *string   `json:"categoryName"`
	Price        string    `json:"price"`
	Quantity     string    `json:"quantity"`
	IsActive     bool      `json:"isActive"`
	InStock      bool      `json:"inStock"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type CatalogListResponse struct {
	Data       []CatalogProductResponse `json:"data"`
	Pagination PaginationResponse       `json:"pagination"`
}
