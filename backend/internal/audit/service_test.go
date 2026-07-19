package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type fakeReadStore struct {
	listFn  func(context.Context, database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error)
	countFn func(context.Context, database.CountAuditEventsParams) (int64, error)
}

func (f *fakeReadStore) ListAuditEvents(ctx context.Context, params database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error) {
	if f.listFn == nil {
		return nil, errors.New("unexpected ListAuditEvents call")
	}
	return f.listFn(ctx, params)
}

func (f *fakeReadStore) CountAuditEvents(ctx context.Context, params database.CountAuditEventsParams) (int64, error) {
	if f.countFn == nil {
		return 0, errors.New("unexpected CountAuditEvents call")
	}
	return f.countFn(ctx, params)
}

func TestServiceListDerivesTenantAndMapsFilters(t *testing.T) {
	organizationID := mustUUID("10000000-0000-4000-8000-000000000001")
	actorUserID := mustUUID("10000000-0000-4000-8000-000000000002")
	actorMembershipID := mustUUID("10000000-0000-4000-8000-000000000003")
	entityID := mustUUID("10000000-0000-4000-8000-000000000004")
	eventID := mustUUID("10000000-0000-4000-8000-000000000005")
	occurredAt := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)

	store := &fakeReadStore{}
	store.listFn = func(_ context.Context, params database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error) {
		if params.OrganizationID != organizationID {
			t.Fatalf("organization ID = %s, want principal organization %s", params.OrganizationID, organizationID)
		}
		if params.PageOffset != 50 || params.PageSize != 25 {
			t.Fatalf("pagination params = offset %d size %d", params.PageOffset, params.PageSize)
		}
		if !params.EventType.Valid || params.EventType.String != EventRoleUpdated {
			t.Fatalf("event type filter = %+v", params.EventType)
		}
		if params.ActorUserID != actorUserID || params.ActorMembershipID != actorMembershipID {
			t.Fatalf("actor filters were not forwarded")
		}
		if !params.EntityType.Valid || params.EntityType.String != "role" || params.EntityID != entityID {
			t.Fatalf("entity filters were not forwarded")
		}
		if !params.OccurredFrom.Valid || !params.OccurredTo.Valid {
			t.Fatalf("occurred range was not forwarded")
		}
		return []database.SecurityAuditEvent{{
			ID: eventID, OrganizationID: organizationID, ActorUserID: actorUserID,
			ActorMembershipID: actorMembershipID, EventType: EventRoleUpdated,
			EntityType: pgtype.Text{String: "role", Valid: true}, EntityID: entityID,
			Outcome:    database.AuditOutcomeSUCCESS,
			Metadata:   []byte(`{"reason":"renamed","access_token":"must-not-leak"}`),
			OccurredAt: pgtype.Timestamptz{Time: occurredAt, Valid: true},
		}}, nil
	}
	store.countFn = func(_ context.Context, params database.CountAuditEventsParams) (int64, error) {
		if params.OrganizationID != organizationID || params.ActorUserID != actorUserID || params.EntityID != entityID {
			t.Fatalf("count predicates do not match list predicates: %+v", params)
		}
		return 51, nil
	}

	page, pageSize := 3, 25
	response, err := NewService(store).List(context.Background(), auditPrincipal(organizationID, authz.ScopeAuditRead), ListInput{
		Page: &page, PageSize: &pageSize, EventType: EventRoleUpdated,
		ActorUserID: actorUserID.String(), ActorMembershipID: actorMembershipID.String(),
		EntityType: "role", EntityID: entityID.String(),
		OccurredFrom: "2026-07-01T00:00:00-03:00", OccurredTo: "2026-08-01T00:00:00Z",
		Sort: "occurredAt", Order: "desc",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if response.Pagination.Page != 3 || response.Pagination.PageSize != 25 || response.Pagination.Total != 51 || response.Pagination.TotalPages != 3 {
		t.Fatalf("pagination = %+v", response.Pagination)
	}
	if len(response.Data) != 1 || response.Data[0].OrganizationID != organizationID.String() {
		t.Fatalf("data = %+v", response.Data)
	}
	if response.Data[0].Metadata["access_token"] != redactedValue || response.Data[0].Metadata["reason"] != "renamed" {
		t.Fatalf("metadata was not safely mapped: %+v", response.Data[0].Metadata)
	}
}

func TestServiceListRequiresOrganizationContextAndScope(t *testing.T) {
	organizationID := mustUUID("20000000-0000-4000-8000-000000000001")
	service := NewService(&fakeReadStore{})

	identity := auditPrincipal(organizationID, authz.ScopeAuditRead)
	identity.ContextKind = authcontext.ContextIdentity
	identity.OrganizationID = pgtype.UUID{}
	identity.MembershipID = pgtype.UUID{}
	if _, err := service.List(context.Background(), identity, ListInput{}); !errors.Is(err, ErrOrganizationContext) {
		t.Fatalf("identity context error = %v, want ErrOrganizationContext", err)
	}

	withoutScope := auditPrincipal(organizationID)
	if _, err := service.List(context.Background(), withoutScope, ListInput{}); !errors.Is(err, ErrInsufficientScope) {
		t.Fatalf("missing scope error = %v, want ErrInsufficientScope", err)
	}
}

func TestNormalizeListInputRejectsInvalidValues(t *testing.T) {
	pageZero, pageSizeTooLarge := 0, 101
	tests := []struct {
		name  string
		input ListInput
		field string
	}{
		{name: "page", input: ListInput{Page: &pageZero}, field: "page"},
		{name: "page size", input: ListInput{PageSize: &pageSizeTooLarge}, field: "pageSize"},
		{name: "actor user", input: ListInput{ActorUserID: "invalid"}, field: "actorUserId"},
		{name: "actor membership", input: ListInput{ActorMembershipID: "invalid"}, field: "actorMembershipId"},
		{name: "entity", input: ListInput{EntityID: "00000000-0000-0000-0000-000000000000"}, field: "entityId"},
		{name: "from", input: ListInput{OccurredFrom: "yesterday"}, field: "occurredFrom"},
		{name: "range", input: ListInput{OccurredFrom: "2026-08-01T00:00:00Z", OccurredTo: "2026-07-01T00:00:00Z"}, field: "occurredTo"},
		{name: "sort", input: ListInput{Sort: "eventType"}, field: "sort"},
		{name: "order", input: ListInput{Order: "asc"}, field: "order"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizeListInput(test.input)
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != test.field {
				t.Fatalf("error = %v, want validation field %q", err, test.field)
			}
		})
	}
}

func auditPrincipal(organizationID pgtype.UUID, scopes ...authcontext.Scope) authcontext.Principal {
	return authcontext.Principal{
		UserID:         mustUUID("90000000-0000-4000-8000-000000000001"),
		SessionID:      mustUUID("90000000-0000-4000-8000-000000000002"),
		ContextKind:    authcontext.ContextOrganization,
		OrganizationID: organizationID,
		MembershipID:   mustUUID("90000000-0000-4000-8000-000000000003"),
		Scopes:         authcontext.NewScopeSet(scopes...),
	}
}
