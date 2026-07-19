package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

const redactedValue = "[REDACTED]"

func mapEvent(row database.SecurityAuditEvent) (EventResponse, error) {
	metadata, err := mapMetadata(row.Metadata)
	if err != nil {
		return EventResponse{}, err
	}
	response := EventResponse{
		ID: row.ID.String(), OrganizationID: row.OrganizationID.String(),
		StoreID: uuidPointer(row.StoreID), ActorUserID: uuidPointer(row.ActorUserID),
		ActorMembershipID: uuidPointer(row.ActorMembershipID), SessionID: uuidPointer(row.SessionID),
		EventType: row.EventType, EntityType: textPointer(row.EntityType), EntityID: uuidPointer(row.EntityID),
		RequestID: textPointer(row.RequestID), UserAgent: textPointer(row.UserAgent),
		Outcome: string(row.Outcome), Metadata: metadata,
	}
	if row.IpAddress != nil {
		value := row.IpAddress.String()
		response.IPAddress = &value
	}
	if row.OccurredAt.Valid {
		response.OccurredAt = row.OccurredAt.Time.UTC()
	}
	return response, nil
}

func mapMetadata(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil || metadata == nil {
		if err == nil {
			err = fmt.Errorf("metadata is not an object")
		}
		return nil, fmt.Errorf("%w: decode stored metadata: %w", ErrReadFailed, err)
	}
	return redactMap(metadata), nil
}

func redactMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		if sensitiveMetadataKey(key) {
			output[key] = redactedValue
			continue
		}
		output[key] = redactValue(value)
	}
	return output
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMap(typed)
	case []any:
		redacted := make([]any, len(typed))
		for i, item := range typed {
			redacted[i] = redactValue(item)
		}
		return redacted
	default:
		return value
	}
}

func sensitiveMetadataKey(key string) bool {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, key)
	return strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "cookie") ||
		strings.Contains(normalized, "credential") ||
		strings.Contains(normalized, "token") ||
		normalized == "apikey" || normalized == "csrf"
}

func uuidPointer(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	result := value.String()
	return &result
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
