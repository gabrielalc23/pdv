package inventory

import "time"

type CreateInventoryEntryInput struct {
	ProductID     string  `json:"productId"`
	Quantity      string  `json:"quantity"`
	Reason        *string `json:"reason"`
	ReferenceType string  `json:"referenceType"`
	ReferenceID   string  `json:"referenceId"`
}

type CreateInventoryAdjustmentInput struct {
	ProductID     string `json:"productId"`
	Direction     string `json:"direction"`
	Quantity      string `json:"quantity"`
	Reason        string `json:"reason"`
	ReferenceType string `json:"referenceType"`
	ReferenceID   string `json:"referenceId"`
}

type ListInventoryInput struct {
	Search     string `json:"search"`
	Page       *int   `json:"page"`
	PageSize   *int   `json:"pageSize"`
	ActiveOnly bool   `json:"activeOnly"`
}

type ListInventoryMovementsInput struct {
	Page     *int   `json:"page"`
	PageSize *int   `json:"pageSize"`
	Type     string `json:"type"`
}

type InventoryResponse struct {
	ProductID string    `json:"productId"`
	SKU       string    `json:"sku"`
	Barcode   *string   `json:"barcode,omitempty"`
	Name      string    `json:"name"`
	Quantity  string    `json:"quantity"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type InventoryChangeSummary struct {
	ProductID        string    `json:"productId"`
	PreviousQuantity string    `json:"previousQuantity"`
	CurrentQuantity  string    `json:"currentQuantity"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type InventoryMovementResponse struct {
	ID               string    `json:"id"`
	ProductID        string    `json:"productId"`
	Type             string    `json:"type"`
	Quantity         string    `json:"quantity"`
	PreviousQuantity string    `json:"previousQuantity"`
	CurrentQuantity  string    `json:"currentQuantity"`
	Reason           *string   `json:"reason,omitempty"`
	ReferenceType    string    `json:"referenceType"`
	ReferenceID      string    `json:"referenceId"`
	CreatedAt        time.Time `json:"createdAt"`
}

type InventoryChangeResponse struct {
	Inventory InventoryChangeSummary    `json:"inventory"`
	Movement  InventoryMovementResponse `json:"movement"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type InventoryListResponse struct {
	Data       []InventoryResponse `json:"data"`
	Pagination PaginationResponse  `json:"pagination"`
}

type InventoryMovementListResponse struct {
	Data       []InventoryMovementResponse `json:"data"`
	Pagination PaginationResponse          `json:"pagination"`
}
