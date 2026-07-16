package checkout

import "time"

type CheckoutInput struct {
	Payments []CheckoutPaymentInput `json:"payments"`
}

type CheckoutPaymentInput struct {
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
