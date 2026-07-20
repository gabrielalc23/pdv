package audit

import "time"

type ListInput struct {
	Page              *int
	PageSize          *int
	StoreID           string
	EventType         string
	Outcome           string
	ActorUserID       string
	ActorMembershipID string
	EntityType        string
	EntityID          string
	OccurredFrom      string
	OccurredTo        string
	Sort              string
	Order             string
}

type EventResponse struct {
	ID                string         `json:"id"`
	OrganizationID    string         `json:"organizationId"`
	StoreID           *string        `json:"storeId"`
	ActorUserID       *string        `json:"actorUserId"`
	ActorMembershipID *string        `json:"actorMembershipId"`
	SessionID         *string        `json:"sessionId"`
	EventType         string         `json:"eventType"`
	EntityType        *string        `json:"entityType"`
	EntityID          *string        `json:"entityId"`
	RequestID         *string        `json:"requestId"`
	IPAddress         *string        `json:"ipAddress"`
	UserAgent         *string        `json:"userAgent"`
	Outcome           string         `json:"outcome"`
	Metadata          map[string]any `json:"metadata"`
	OccurredAt        time.Time      `json:"occurredAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ListResponse struct {
	Data       []EventResponse    `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
