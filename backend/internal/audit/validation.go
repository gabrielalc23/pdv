package audit

import (
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	sortOccurredAt  = "occurredAt"
	orderDescending = "desc"
)

type normalizedListInput struct {
	page              int
	pageSize          int
	storeID           pgtype.UUID
	eventType         pgtype.Text
	outcome           database.NullAuditOutcome
	actorUserID       pgtype.UUID
	actorMembershipID pgtype.UUID
	entityType        pgtype.Text
	entityID          pgtype.UUID
	occurredFrom      pgtype.Timestamptz
	occurredTo        pgtype.Timestamptz
}

func normalizeListInput(input ListInput) (normalizedListInput, error) {
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return normalizedListInput{}, err
	}

	eventType, err := optionalText(input.EventType, "eventType", 100)
	if err != nil {
		return normalizedListInput{}, err
	}
	storeID, err := optionalUUID(input.StoreID, "storeId")
	if err != nil {
		return normalizedListInput{}, err
	}
	var outcome database.NullAuditOutcome
	if value := strings.ToUpper(strings.TrimSpace(input.Outcome)); value != "" {
		parsed := database.AuditOutcome(value)
		if !parsed.Valid() {
			return normalizedListInput{}, validationError("outcome", "must be SUCCESS or FAILURE")
		}
		outcome = database.NullAuditOutcome{AuditOutcome: parsed, Valid: true}
	}
	actorUserID, err := optionalUUID(input.ActorUserID, "actorUserId")
	if err != nil {
		return normalizedListInput{}, err
	}
	actorMembershipID, err := optionalUUID(input.ActorMembershipID, "actorMembershipId")
	if err != nil {
		return normalizedListInput{}, err
	}
	entityType, err := optionalText(input.EntityType, "entityType", 80)
	if err != nil {
		return normalizedListInput{}, err
	}
	entityID, err := optionalUUID(input.EntityID, "entityId")
	if err != nil {
		return normalizedListInput{}, err
	}
	occurredFrom, err := optionalTimestamp(input.OccurredFrom, "occurredFrom")
	if err != nil {
		return normalizedListInput{}, err
	}
	occurredTo, err := optionalTimestamp(input.OccurredTo, "occurredTo")
	if err != nil {
		return normalizedListInput{}, err
	}
	if occurredFrom.Valid && occurredTo.Valid && !occurredFrom.Time.Before(occurredTo.Time) {
		return normalizedListInput{}, validationError("occurredTo", "must be after occurredFrom")
	}

	sort := strings.TrimSpace(input.Sort)
	if sort == "" {
		sort = sortOccurredAt
	}
	if sort != sortOccurredAt && sort != "occurred_at" {
		return normalizedListInput{}, validationError("sort", "must be occurredAt")
	}
	order := strings.ToLower(strings.TrimSpace(input.Order))
	if order == "" {
		order = orderDescending
	}
	if order != orderDescending {
		return normalizedListInput{}, validationError("order", "must be desc")
	}

	return normalizedListInput{
		page: page, pageSize: pageSize, storeID: storeID, eventType: eventType, outcome: outcome,
		actorUserID: actorUserID, actorMembershipID: actorMembershipID,
		entityType: entityType, entityID: entityID,
		occurredFrom: occurredFrom, occurredTo: occurredTo,
	}, nil
}

func (input normalizedListInput) listParams(organizationID pgtype.UUID) database.ListAuditEventsParams {
	return database.ListAuditEventsParams{
		OrganizationID: organizationID, StoreID: input.storeID, ActorUserID: input.actorUserID,
		ActorMembershipID: input.actorMembershipID, EventType: input.eventType,
		Outcome:    input.outcome,
		EntityType: input.entityType, EntityID: input.entityID,
		OccurredFrom: input.occurredFrom, OccurredTo: input.occurredTo,
		PageOffset: int32((input.page - 1) * input.pageSize), PageSize: int32(input.pageSize),
	}
}

func (input normalizedListInput) countParams(organizationID pgtype.UUID) database.CountAuditEventsParams {
	return database.CountAuditEventsParams{
		OrganizationID: organizationID, StoreID: input.storeID, ActorUserID: input.actorUserID,
		ActorMembershipID: input.actorMembershipID, EventType: input.eventType,
		Outcome:    input.outcome,
		EntityType: input.entityType, EntityID: input.entityID,
		OccurredFrom: input.occurredFrom, OccurredTo: input.occurredTo,
	}
}

func normalizePagination(page, pageSize *int) (int, int, error) {
	resolvedPage := 1
	if page != nil {
		if *page < 1 {
			return 0, 0, validationError("page", "must be greater than zero")
		}
		resolvedPage = *page
	}
	resolvedPageSize := defaultPageSize
	if pageSize != nil {
		if *pageSize < 1 {
			return 0, 0, validationError("pageSize", "must be greater than zero")
		}
		if *pageSize > maxPageSize {
			return 0, 0, validationError("pageSize", "must be at most 100")
		}
		resolvedPageSize = *pageSize
	}
	if resolvedPage-1 > math.MaxInt32/resolvedPageSize {
		return 0, 0, validationError("page", "is too large")
	}
	return resolvedPage, resolvedPageSize, nil
}

func optionalText(raw, field string, maxLength int) (pgtype.Text, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return pgtype.Text{}, nil
	}
	if utf8.RuneCountInString(value) > maxLength {
		return pgtype.Text{}, validationError(field, "is too long")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return pgtype.Text{}, validationError(field, "must not contain control characters")
		}
	}
	return pgtype.Text{String: value, Valid: true}, nil
}

func optionalUUID(raw, field string) (pgtype.UUID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return pgtype.UUID{}, nil
	}
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid || zeroUUID(id) {
		return pgtype.UUID{}, validationError(field, "must be a valid UUID")
	}
	return id, nil
}

func optionalTimestamp(raw, field string) (pgtype.Timestamptz, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return pgtype.Timestamptz{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return pgtype.Timestamptz{}, validationError(field, "must be a valid RFC3339 timestamp")
	}
	return pgtype.Timestamptz{Time: parsed.UTC(), Valid: true}, nil
}

func zeroUUID(id pgtype.UUID) bool {
	for _, value := range id.Bytes {
		if value != 0 {
			return false
		}
	}
	return true
}

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
