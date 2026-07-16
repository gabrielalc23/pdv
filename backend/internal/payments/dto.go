package payments

import "time"

type PaymentMethodsResponse struct {
	Data []PaymentMethodResponse `json:"data"`
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
