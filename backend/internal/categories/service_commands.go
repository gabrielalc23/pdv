package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ensureNameAvailable(ctx context.Context, name, currentID string) error {
	category, err := s.store.GetCategoryByName(ctx, name)
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

func (s *Service) ensureSlugAvailable(ctx context.Context, slug, currentID string) error {
	category, err := s.store.GetCategoryBySlug(ctx, slug)
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

func (s *Service) Create(ctx context.Context, input UpsertCategoryInput) (CategoryResponse, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return CategoryResponse{}, err
	}
	slug := slugify(name)
	if err := s.ensureNameAvailable(ctx, name, ""); err != nil {
		return CategoryResponse{}, err
	}
	if err := s.ensureSlugAvailable(ctx, slug, ""); err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.store.CreateCategory(ctx, database.CreateCategoryParams{Name: name, Slug: slug})
	if err != nil {
		return CategoryResponse{}, translatePersistenceError(err)
	}
	return toCategoryResponse(category), nil
}

func (s *Service) Update(ctx context.Context, rawID string, input UpsertCategoryInput) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}
	name, err := normalizeName(input.Name)
	if err != nil {
		return CategoryResponse{}, err
	}
	slug := slugify(name)
	if err := s.ensureNameAvailable(ctx, name, rawID); err != nil {
		return CategoryResponse{}, err
	}
	if err := s.ensureSlugAvailable(ctx, slug, rawID); err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.store.UpdateCategory(ctx, database.UpdateCategoryParams{
		ID: id, Name: name, Slug: slug,
	})
	if err != nil {
		return CategoryResponse{}, translatePersistenceError(err)
	}
	return toCategoryResponse(category), nil
}

func (s *Service) Activate(ctx context.Context, rawID string) (CategoryResponse, error) {
	return s.setActive(ctx, rawID, true)
}

func (s *Service) Deactivate(ctx context.Context, rawID string) (CategoryResponse, error) {
	return s.setActive(ctx, rawID, false)
}

func (s *Service) setActive(ctx context.Context, rawID string, active bool) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}

	var category database.Category
	if active {
		category, err = s.store.ActivateCategory(ctx, id)
	} else {
		category, err = s.store.DeactivateCategory(ctx, id)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CategoryResponse{}, ErrCategoryNotFound
		}
		return CategoryResponse{}, translatePersistenceError(err)
	}
	return toCategoryResponse(category), nil
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
