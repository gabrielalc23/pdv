package memberships

import "time"

type ListInput struct {
	Search   string
	Status   string
	Page     *int
	PageSize *int
}

type UpdateDefaultStoreInput struct {
	DefaultStoreID *string `json:"defaultStoreId"`
}

type UpdateStatusInput struct {
	Status string `json:"status"`
}

type UserResponse struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	DisplayName     string  `json:"displayName"`
	Status          string  `json:"status"`
	EmailVerifiedAt *string `json:"emailVerifiedAt"`
}

type StoreResponse struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

type RoleBindingResponse struct {
	ID              string         `json:"id"`
	RoleID          string         `json:"roleId"`
	RoleKey         string         `json:"roleKey"`
	RoleName        string         `json:"roleName"`
	AssignmentScope string         `json:"assignmentScope"`
	Store           *StoreResponse `json:"store"`
	ExpiresAt       *string        `json:"expiresAt"`
	IsActive        bool           `json:"isActive"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type MembershipResponse struct {
	ID                   string                `json:"id"`
	OrganizationID       string                `json:"organizationId"`
	User                 UserResponse          `json:"user"`
	Status               string                `json:"status"`
	DefaultStore         *StoreResponse        `json:"defaultStore"`
	AuthorizationVersion int64                 `json:"authorizationVersion"`
	RoleBindings         []RoleBindingResponse `json:"roleBindings,omitempty"`
	JoinedAt             time.Time             `json:"joinedAt"`
	SuspendedAt          *string               `json:"suspendedAt"`
	RemovedAt            *string               `json:"removedAt"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ListResponse struct {
	Data       []MembershipResponse `json:"data"`
	Pagination PaginationResponse   `json:"pagination"`
}
