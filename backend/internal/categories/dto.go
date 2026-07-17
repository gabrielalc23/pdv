package categories

import "time"

type UpsertCategoryInput struct {
	Name string `json:"name"`
}

type ListCategoriesInput struct {
	Search     string `json:"search"`
	ActiveOnly bool   `json:"activeOnly"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CategoryListResponse struct {
	Data []CategoryResponse `json:"data"`
}
