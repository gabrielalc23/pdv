package sales

import "time"

type CreateSaleInput struct {
	IdempotencyKey string `json:"idempotencyKey"`
}

type AddSaleItemInput struct {
	ProductID string  `json:"productId"`
	Quantity  string  `json:"quantity"`
	Discount  *string `json:"discount"`
}

type UpdateSaleItemInput struct {
	Quantity string  `json:"quantity"`
	Discount *string `json:"discount"`
}

type ListSalesInput struct {
	Status   string
	Page     *int
	PageSize *int
}

type SaleItemResponse struct {
	ID          string    `json:"id"`
	SaleID      string    `json:"saleId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	ProductSKU  string    `json:"productSku"`
	UnitPrice   string    `json:"unitPrice"`
	Quantity    string    `json:"quantity"`
	Discount    string    `json:"discount"`
	Total       string    `json:"total"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SaleResponse struct {
	ID             string             `json:"id"`
	Number         int64              `json:"number"`
	Status         string             `json:"status"`
	Subtotal       string             `json:"subtotal"`
	Discount       string             `json:"discount"`
	Addition       string             `json:"addition"`
	Total          string             `json:"total"`
	OpenedAt       time.Time          `json:"openedAt"`
	CompletedAt    time.Time          `json:"completedAt"`
	CancelledAt    time.Time          `json:"cancelledAt"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	IdempotencyKey string             `json:"idempotencyKey"`
	Items          []SaleItemResponse `json:"items"`
}

type SaleListItemResponse struct {
	ID             string    `json:"id"`
	Number         int64     `json:"number"`
	Status         string    `json:"status"`
	Subtotal       string    `json:"subtotal"`
	Discount       string    `json:"discount"`
	Addition       string    `json:"addition"`
	Total          string    `json:"total"`
	OpenedAt       time.Time `json:"openedAt"`
	CompletedAt    time.Time `json:"completedAt"`
	CancelledAt    time.Time `json:"cancelledAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	IdempotencyKey string    `json:"idempotencyKey"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type SaleListResponse struct {
	Data       []SaleListItemResponse `json:"data"`
	Pagination PaginationResponse     `json:"pagination"`
}
