package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func doRequest(method, path string, body any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	return resp, nil
}

func decode[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode: %w", err)
	}
	return v, nil
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Field   string `json:"field,omitempty"`
	} `json:"error"`
}

func decodeError(resp *http.Response) string {
	defer resp.Body.Close()
	var eb errorBody
	json.NewDecoder(resp.Body).Decode(&eb)
	return eb.Error.Message
}

// ---- DTOs ----

type UpsertProductRequest struct {
	SKU     string  `json:"sku"`
	Barcode *string `json:"barcode"`
	Name    string  `json:"name"`
	Price   string  `json:"price"`
	Cost    *string `json:"cost"`
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

type CreateSaleRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
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

type AddSaleItemRequest struct {
	ProductID string  `json:"productId"`
	Quantity  string  `json:"quantity"`
	Discount  *string `json:"discount"`
}

type UpdateSaleItemRequest struct {
	Quantity string  `json:"quantity"`
	Discount *string `json:"discount"`
}

type CheckoutRequest struct {
	Payments []CheckoutPaymentRequest `json:"payments"`
}

type CheckoutPaymentRequest struct {
	PaymentMethodID   string  `json:"paymentMethodId"`
	Amount            string  `json:"amount"`
	ReceivedAmount    *string `json:"receivedAmount,omitempty"`
	Installments      *int    `json:"installments,omitempty"`
	ExternalReference *string `json:"externalReference,omitempty"`
}

type CheckoutResponse struct {
	Sale           CheckoutSaleResponse      `json:"sale"`
	Payments       []CheckoutPaymentResponse `json:"payments"`
	FiscalDocument FiscalDocumentResponse    `json:"fiscalDocument"`
}

type CheckoutSaleResponse struct {
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

type CheckoutPaymentResponse struct {
	ID                string    `json:"id"`
	SaleID            string    `json:"saleId"`
	PaymentMethodID   string    `json:"paymentMethodId"`
	PaymentMethodCode string    `json:"paymentMethodCode"`
	PaymentMethodName string    `json:"paymentMethodName"`
	PaymentMethodKind string    `json:"paymentMethodKind"`
	Amount            string    `json:"amount"`
	ReceivedAmount    *string   `json:"receivedAmount,omitempty"`
	ChangeAmount      *string   `json:"changeAmount,omitempty"`
	Status            string    `json:"status"`
	Installments      int16     `json:"installments"`
	ExternalReference *string   `json:"externalReference,omitempty"`
	PaidAt            time.Time `json:"paidAt"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type FiscalDocumentResponse struct {
	ID                string     `json:"id"`
	SaleID            string     `json:"saleId"`
	Status            string     `json:"status"`
	Environment       string     `json:"environment"`
	DocumentModel     int16      `json:"documentModel"`
	Series            *int32     `json:"series,omitempty"`
	Number            *int64     `json:"number,omitempty"`
	AccessKey         *string    `json:"accessKey,omitempty"`
	Protocol          *string    `json:"protocol,omitempty"`
	Provider          *string    `json:"provider,omitempty"`
	ExternalReference *string    `json:"externalReference,omitempty"`
	ErrorCode         *string    `json:"errorCode,omitempty"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
	IssuedAt          *time.Time `json:"issuedAt,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type ReceiptResponse struct {
	Sale           ReceiptSaleResponse      `json:"sale"`
	Items          []ReceiptItemResponse    `json:"items"`
	Payments       []ReceiptPaymentResponse `json:"payments"`
	FiscalDocument *ReceiptFiscalResponse   `json:"fiscalDocument"`
}

type ReceiptSaleResponse struct {
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

type ReceiptItemResponse struct {
	ProductID string    `json:"productId"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	UnitPrice string    `json:"unitPrice"`
	Quantity  string    `json:"quantity"`
	Subtotal  string    `json:"subtotal"`
	Discount  string    `json:"discount"`
	Total     string    `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
}

type ReceiptPaymentResponse struct {
	Method            string  `json:"method"`
	Amount            string  `json:"amount"`
	Status            string  `json:"status"`
	Installments      int16   `json:"installments"`
	ReceivedAmount    *string `json:"receivedAmount,omitempty"`
	ChangeAmount      *string `json:"changeAmount,omitempty"`
	ExternalReference *string `json:"externalReference,omitempty"`
}

type ReceiptFiscalResponse struct {
	Status            string     `json:"status"`
	AccessKey         *string    `json:"accessKey,omitempty"`
	Protocol          *string    `json:"protocol,omitempty"`
	Provider          *string    `json:"provider,omitempty"`
	ExternalReference *string    `json:"externalReference,omitempty"`
	ErrorCode         *string    `json:"errorCode,omitempty"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
	IssuedAt          *time.Time `json:"issuedAt,omitempty"`
}

type InventoryEntryRequest struct {
	ProductID     string  `json:"productId"`
	Quantity      string  `json:"quantity"`
	Reason        *string `json:"reason"`
	ReferenceType string  `json:"referenceType"`
	ReferenceID   string  `json:"referenceId"`
}

type InventoryChangeResponse struct {
	Inventory struct {
		ProductID        string    `json:"productId"`
		PreviousQuantity string    `json:"previousQuantity"`
		CurrentQuantity  string    `json:"currentQuantity"`
		UpdatedAt        time.Time `json:"updatedAt"`
	} `json:"inventory"`
	Movement struct {
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
	} `json:"movement"`
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

type InventoryListResponse struct {
	Data       []InventoryResponse `json:"data"`
	Pagination PaginationResponse  `json:"pagination"`
}

type PaymentMethodResponse struct {
	ID                 string    `json:"id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Kind               string    `json:"kind"`
	IsActive           bool      `json:"isActive"`
	AllowsChange       bool      `json:"allowsChange"`
	AllowsInstallments bool      `json:"allowsInstallments"`
	MaxInstallments    int16     `json:"maxInstallments"`
	FeePercentage      string    `json:"feePercentage"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type PaymentMethodsResponse struct {
	Data []PaymentMethodResponse `json:"data"`
}

type SalePaymentsResponse struct {
	Data []SalePaymentResponse `json:"data"`
}

type SalePaymentResponse struct {
	ID                string     `json:"id"`
	SaleID            string     `json:"saleId"`
	PaymentMethodID   string     `json:"paymentMethodId"`
	PaymentMethodCode string     `json:"paymentMethodCode"`
	PaymentMethodName string     `json:"paymentMethodName"`
	Amount            string     `json:"amount"`
	ReceivedAmount    *string    `json:"receivedAmount,omitempty"`
	ChangeAmount      *string    `json:"changeAmount,omitempty"`
	Status            string     `json:"status"`
	Installments      int16      `json:"installments"`
	ExternalReference *string    `json:"externalReference,omitempty"`
	PaidAt            *time.Time `json:"paidAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
