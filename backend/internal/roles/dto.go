package roles

import "time"

type UpsertRoleInput struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	Description     *string  `json:"description"`
	AssignmentScope string   `json:"assignmentScope"`
	Scopes          []string `json:"scopes"`
}

type CreateBindingInput struct {
	RoleID    string     `json:"roleId"`
	StoreID   *string    `json:"storeId"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type ScopeResponse struct {
	Code         string    `json:"code"`
	Resource     string    `json:"resource"`
	Action       string    `json:"action"`
	ScopeLevel   string    `json:"scopeLevel"`
	Description  string    `json:"description"`
	IsAssignable bool      `json:"isAssignable"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ScopeListResponse struct {
	Data []ScopeResponse `json:"data"`
}

type RoleResponse struct {
	ID              string    `json:"id"`
	Key             string    `json:"key"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	AssignmentScope string    `json:"assignmentScope"`
	IsSystem        bool      `json:"isSystem"`
	IsMutable       bool      `json:"isMutable"`
	IsActive        bool      `json:"isActive"`
	Scopes          []string  `json:"scopes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type RoleListResponse struct {
	Data []RoleResponse `json:"data"`
}

type BindingRoleResponse struct {
	ID              string `json:"id"`
	Key             string `json:"key"`
	Name            string `json:"name"`
	AssignmentScope string `json:"assignmentScope"`
}

type BindingStoreResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type BindingResponse struct {
	ID           string                `json:"id"`
	MembershipID string                `json:"membershipId"`
	Role         BindingRoleResponse   `json:"role"`
	Store        *BindingStoreResponse `json:"store"`
	ExpiresAt    *time.Time            `json:"expiresAt"`
	CreatedAt    time.Time             `json:"createdAt"`
}
