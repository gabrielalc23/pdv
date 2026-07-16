package fiscal

import (
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func toFiscalDocumentResponse(row database.FiscalDocument) FiscalDocumentResponse {
	var series *int32
	if row.Series.Valid {
		value := row.Series.Int32
		series = &value
	}

	var number *int64
	if row.Number.Valid {
		value := row.Number.Int64
		number = &value
	}

	var accessKey *string
	if row.AccessKey.Valid {
		value := row.AccessKey.String
		accessKey = &value
	}

	var protocol *string
	if row.Protocol.Valid {
		value := row.Protocol.String
		protocol = &value
	}

	var provider *string
	if row.Provider.Valid {
		value := row.Provider.String
		provider = &value
	}

	var externalReference *string
	if row.ExternalReference.Valid {
		value := row.ExternalReference.String
		externalReference = &value
	}

	var xml *string
	if row.XML.Valid {
		value := row.XML.String
		xml = &value
	}

	var errorCode *string
	if row.ErrorCode.Valid {
		value := row.ErrorCode.String
		errorCode = &value
	}

	var errorMessage *string
	if row.ErrorMessage.Valid {
		value := row.ErrorMessage.String
		errorMessage = &value
	}

	var issuedAt *time.Time
	if row.IssuedAt.Valid {
		value := row.IssuedAt.Time.UTC()
		issuedAt = &value
	}

	var cancelledAt *time.Time
	if row.CancelledAt.Valid {
		value := row.CancelledAt.Time.UTC()
		cancelledAt = &value
	}

	return FiscalDocumentResponse{
		ID:                row.ID.String(),
		SaleID:            row.SaleID.String(),
		Status:            string(row.Status),
		Environment:       string(row.Environment),
		DocumentModel:     row.DocumentModel,
		Series:            series,
		Number:            number,
		AccessKey:         accessKey,
		Protocol:          protocol,
		Provider:          provider,
		ExternalReference: externalReference,
		XML:               xml,
		ErrorCode:         errorCode,
		ErrorMessage:      errorMessage,
		IssuedAt:          issuedAt,
		CancelledAt:       cancelledAt,
		CreatedAt:         row.CreatedAt.Time.UTC(),
		UpdatedAt:         row.UpdatedAt.Time.UTC(),
	}
}
