package fiscal

import "time"

type AuthorizationInput struct {
	SaleID        string
	SaleNumber    int64
	SaleTotal     string
	ItemCount     int
	PaymentCount  int
	CompletedAt   time.Time
	Environment   string
	DocumentModel int16
}

type AuthorizationResult struct {
	Provider          string
	AccessKey         string
	Protocol          string
	XML               string
	ExternalReference string
	AuthorizedAt      time.Time
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
	XML               *string    `json:"xml,omitempty"`
	ErrorCode         *string    `json:"errorCode,omitempty"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
	IssuedAt          *time.Time `json:"issuedAt,omitempty"`
	CancelledAt       *time.Time `json:"cancelledAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
