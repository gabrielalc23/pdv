package receipt

import "time"

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
