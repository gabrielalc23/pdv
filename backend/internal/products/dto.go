package products

import "time"

type UpsertProductInput struct {
	SKU     string  `json:"sku"`
	Barcode *string `json:"barcode"`
	Name    string  `json:"name"`
	Price   string  `json:"price"`
	Cost    *string `json:"cost"`
}

type ListProductsInput struct {
	Search     string `json:"search"`
	Page       *int   `json:"page"`
	PageSize   *int   `json:"pageSize"`
	ActiveOnly bool   `json:"activeOnly"`
}

type ProductResponse struct {
	ID        string    `json:"id"`
	SKU       string    `json:"sku"`
	Barcode   *string   `json:"barcode"`
	Name      string    `json:"name"`
	Price     string    `json:"price"`
	Cost      *string   `json:"cost"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ProductListResponse struct {
	Data       []ProductResponse  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
