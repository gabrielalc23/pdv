package stores

import "time"

type ListStoresInput struct {
	Search   string
	Status   string
	Page     *int
	PageSize *int
}

type CreateStoreInput struct {
	Code               string `json:"code"`
	Name               string `json:"name"`
	Timezone           string `json:"timezone"`
	CopyPaymentMethods *bool  `json:"copyPaymentMethods,omitempty"`
}

type UpdateStoreInput struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

type StoreResponse struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	Timezone   string     `json:"timezone"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type StoreListResponse struct {
	Data       []StoreResponse    `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type UpsertPaymentMethodInput struct {
	Code                      string  `json:"code"`
	Name                      string  `json:"name"`
	Kind                      string  `json:"kind"`
	Provider                  *string `json:"provider"`
	AllowsChange              bool    `json:"allowsChange"`
	RequiresExternalReference bool    `json:"requiresExternalReference"`
	AllowsInstallments        bool    `json:"allowsInstallments"`
	MaxInstallments           int16   `json:"maxInstallments"`
	FeePercentage             string  `json:"feePercentage"`
	SettlementDays            int32   `json:"settlementDays"`
	IsActive                  *bool   `json:"isActive,omitempty"`
	SortOrder                 int32   `json:"sortOrder"`
}

type PaymentMethodResponse struct {
	ID                        string    `json:"id"`
	Code                      string    `json:"code"`
	Name                      string    `json:"name"`
	Kind                      string    `json:"kind"`
	Provider                  *string   `json:"provider"`
	AllowsChange              bool      `json:"allowsChange"`
	RequiresExternalReference bool      `json:"requiresExternalReference"`
	AllowsInstallments        bool      `json:"allowsInstallments"`
	MaxInstallments           int16     `json:"maxInstallments"`
	FeePercentage             string    `json:"feePercentage"`
	SettlementDays            int32     `json:"settlementDays"`
	IsActive                  bool      `json:"isActive"`
	SortOrder                 int32     `json:"sortOrder"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

type PaymentMethodListResponse struct {
	Data []PaymentMethodResponse `json:"data"`
}

type StorePaymentMethodInput struct {
	PaymentMethodID string `json:"paymentMethodId"`
	IsActive        bool   `json:"isActive"`
	SortOrder       int32  `json:"sortOrder"`
}

type ReplaceStorePaymentMethodsInput struct {
	PaymentMethods []StorePaymentMethodInput `json:"paymentMethods"`
}

type UpdateStorePaymentMethodInput struct {
	IsActive  *bool  `json:"isActive,omitempty"`
	SortOrder *int32 `json:"sortOrder,omitempty"`
}

type StorePaymentMethodResponse struct {
	PaymentMethodID    string    `json:"paymentMethodId"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Kind               string    `json:"kind"`
	IsActive           bool      `json:"isActive"`
	OrganizationActive bool      `json:"organizationActive"`
	SortOrder          int32     `json:"sortOrder"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type StorePaymentMethodListResponse struct {
	Data []StorePaymentMethodResponse `json:"data"`
}
