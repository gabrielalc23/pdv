package audit

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

type Service struct {
	store ReadStore
}

func NewService(store ReadStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, principal authcontext.Principal, input ListInput) (ListResponse, error) {
	if !principal.HasOrganizationScope() || !principal.OrganizationID.Valid || !principal.MembershipID.Valid {
		return ListResponse{}, ErrOrganizationContext
	}
	if !principal.Scopes.Has(authz.ScopeAuditRead) {
		return ListResponse{}, ErrInsufficientScope
	}
	normalized, err := normalizeListInput(input)
	if err != nil {
		return ListResponse{}, err
	}

	rows, err := s.store.ListAuditEvents(ctx, normalized.listParams(principal.OrganizationID))
	if err != nil {
		return ListResponse{}, fmt.Errorf("%w: list events: %w", ErrReadFailed, err)
	}
	total, err := s.store.CountAuditEvents(ctx, normalized.countParams(principal.OrganizationID))
	if err != nil {
		return ListResponse{}, fmt.Errorf("%w: count events: %w", ErrReadFailed, err)
	}

	data := make([]EventResponse, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapEvent(row)
		if err != nil {
			return ListResponse{}, err
		}
		data = append(data, mapped)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(normalized.pageSize) - 1) / int64(normalized.pageSize))
	}
	return ListResponse{
		Data: data,
		Pagination: PaginationResponse{
			Page: normalized.page, PageSize: normalized.pageSize,
			Total: total, TotalPages: totalPages,
		},
	}, nil
}
