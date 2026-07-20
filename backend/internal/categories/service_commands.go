package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ensureNameAvailable(ctx context.Context, actor authn.OrganizationActor, name, currentID string) error {
	scope := actor.ToOrganizationScope()
	category, err := s.store.GetCategoryByName(ctx, scope, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check category name availability: %w", err)
	}
	if currentID != "" && category.ID.String() == currentID {
		return nil
	}
	return ErrCategoryNameExists
}

func (s *Service) ensureSlugAvailable(ctx context.Context, actor authn.OrganizationActor, slug, currentID string) error {
	scope := actor.ToOrganizationScope()
	category, err := s.store.GetCategoryBySlug(ctx, scope, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check category slug availability: %w", err)
	}
	if currentID != "" && category.ID.String() == currentID {
		return nil
	}
	return ErrCategorySlugExists
}

func (s *Service) Create(ctx context.Context, actor authn.OrganizationActor, input UpsertCategoryInput) (CategoryResponse, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return CategoryResponse{}, err
	}
	slug := slugify(name)
	scope := actor.ToOrganizationScope()
	if err := s.ensureNameAvailable(ctx, actor, name, ""); err != nil {
		return CategoryResponse{}, err
	}
	if err := s.ensureSlugAvailable(ctx, actor, slug, ""); err != nil {
		return CategoryResponse{}, err
	}

	row, err := s.store.CreateCategory(ctx, scope, database.CreateCategoryForOrganizationParams{Name: name, Slug: slug})
	if err != nil {
		return CategoryResponse{}, translatePersistenceError(err)
	}
	return toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) Update(ctx context.Context, actor authn.OrganizationActor, rawID string, input UpsertCategoryInput) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}
	name, err := normalizeName(input.Name)
	if err != nil {
		return CategoryResponse{}, err
	}
	slug := slugify(name)
	scope := actor.ToOrganizationScope()
	if err := s.ensureNameAvailable(ctx, actor, name, rawID); err != nil {
		return CategoryResponse{}, err
	}
	if err := s.ensureSlugAvailable(ctx, actor, slug, rawID); err != nil {
		return CategoryResponse{}, err
	}

	row, err := s.store.UpdateCategory(ctx, scope, database.UpdateCategoryForOrganizationParams{
		ID: id, Name: name, Slug: slug,
	})
	if err != nil {
		return CategoryResponse{}, translatePersistenceError(err)
	}
	return toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) Activate(ctx context.Context, actor authn.OrganizationActor, rawID string) (CategoryResponse, error) {
	return s.setActive(ctx, actor, rawID, true)
}

func (s *Service) Deactivate(ctx context.Context, actor authn.OrganizationActor, rawID string) (CategoryResponse, error) {
	return s.setActive(ctx, actor, rawID, false)
}

func (s *Service) setActive(ctx context.Context, actor authn.OrganizationActor, rawID string, active bool) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}
	scope := actor.ToOrganizationScope()

	var row struct {
		ID        pgtype.UUID
		Name      string
		Slug      string
		IsActive  bool
		CreatedAt pgtype.Timestamptz
		UpdatedAt pgtype.Timestamptz
	}
	if active {
		r, err := s.store.ActivateCategory(ctx, scope, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CategoryResponse{}, ErrCategoryNotFound
			}
			return CategoryResponse{}, translatePersistenceError(err)
		}
		row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt = r.ID, r.Name, r.Slug, r.IsActive, r.CreatedAt, r.UpdatedAt
	} else {
		r, err := s.store.DeactivateCategory(ctx, scope, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CategoryResponse{}, ErrCategoryNotFound
			}
			return CategoryResponse{}, translatePersistenceError(err)
		}
		row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt = r.ID, r.Name, r.Slug, r.IsActive, r.CreatedAt, r.UpdatedAt
	}

	return toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func parseUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil || !id.Valid {
		return pgtype.UUID{}, newValidationError("id", "must be a valid UUID")
	}
	return id, nil
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func translatePersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCategoryNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "categories_organization_id_name_unique":
			return ErrCategoryNameExists
		case "categories_organization_id_slug_unique":
			return ErrCategorySlugExists
		}
	}

	return fmt.Errorf("database operation failed: %w", err)
}
