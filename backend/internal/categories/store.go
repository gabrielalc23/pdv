package categories

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateCategory(ctx context.Context, scope tenancy.OrganizationScope, params database.CreateCategoryForOrganizationParams) (database.CreateCategoryForOrganizationRow, error)
	GetCategoryByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.GetCategoryByIDForOrganizationRow, error)
	GetCategoryByName(ctx context.Context, scope tenancy.OrganizationScope, name string) (database.GetCategoryByNameForOrganizationRow, error)
	GetCategoryBySlug(ctx context.Context, scope tenancy.OrganizationScope, slug string) (database.GetCategoryBySlugForOrganizationRow, error)
	ListCategories(ctx context.Context, scope tenancy.OrganizationScope, params database.ListCategoriesForOrganizationParams) ([]database.ListCategoriesForOrganizationRow, error)
	UpdateCategory(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateCategoryForOrganizationParams) (database.UpdateCategoryForOrganizationRow, error)
	ActivateCategory(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateCategoryForOrganizationRow, error)
	DeactivateCategory(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateCategoryForOrganizationRow, error)
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) CreateCategory(ctx context.Context, scope tenancy.OrganizationScope, params database.CreateCategoryForOrganizationParams) (database.CreateCategoryForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.CreateCategoryForOrganization(ctx, params)
}

func (s *storeImpl) GetCategoryByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.GetCategoryByIDForOrganizationRow, error) {
	return s.q.GetCategoryByIDForOrganization(ctx, database.GetCategoryByIDForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}

func (s *storeImpl) GetCategoryByName(ctx context.Context, scope tenancy.OrganizationScope, name string) (database.GetCategoryByNameForOrganizationRow, error) {
	return s.q.GetCategoryByNameForOrganization(ctx, database.GetCategoryByNameForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		Name:           name,
	})
}

func (s *storeImpl) GetCategoryBySlug(ctx context.Context, scope tenancy.OrganizationScope, slug string) (database.GetCategoryBySlugForOrganizationRow, error) {
	return s.q.GetCategoryBySlugForOrganization(ctx, database.GetCategoryBySlugForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		Slug:           slug,
	})
}

func (s *storeImpl) ListCategories(ctx context.Context, scope tenancy.OrganizationScope, params database.ListCategoriesForOrganizationParams) ([]database.ListCategoriesForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.ListCategoriesForOrganization(ctx, params)
}

func (s *storeImpl) UpdateCategory(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateCategoryForOrganizationParams) (database.UpdateCategoryForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.UpdateCategoryForOrganization(ctx, params)
}

func (s *storeImpl) ActivateCategory(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateCategoryForOrganizationRow, error) {
	return s.q.ActivateCategoryForOrganization(ctx, database.ActivateCategoryForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}

func (s *storeImpl) DeactivateCategory(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateCategoryForOrganizationRow, error) {
	return s.q.DeactivateCategoryForOrganization(ctx, database.DeactivateCategoryForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}
