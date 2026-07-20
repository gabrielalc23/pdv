package invitations

type AssignmentInput struct {
	RoleID  string  `json:"roleId"`
	StoreID *string `json:"storeId"`
}

type CreateInput struct {
	Email       string            `json:"email"`
	Assignments []AssignmentInput `json:"assignments"`
}

type ListInput struct {
	Status      string
	Email       string
	CreatedFrom string
	CreatedTo   string
	Page        int
	PageSize    int
}

type InspectInput struct {
	Token string `json:"token"`
}

type AcceptInput struct {
	Token       string `json:"token"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	ClientID    string `json:"clientId"`
	DeviceName  string `json:"deviceName"`
}

type OrganizationResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type RoleResponse struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type StoreResponse struct {
	ID   string `json:"id"`
	Code string `json:"code,omitempty"`
	Name string `json:"name"`
}

type AssignmentResponse struct {
	Role  RoleResponse   `json:"role"`
	Store *StoreResponse `json:"store"`
}

type InvitationResponse struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	Status      string               `json:"status"`
	ExpiresAt   string               `json:"expiresAt"`
	Assignments []AssignmentResponse `json:"assignments"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ListResponse struct {
	Data       []InvitationResponse `json:"data"`
	Pagination PaginationResponse   `json:"pagination"`
}

type InspectResponse struct {
	Organization OrganizationResponse `json:"organization"`
	EmailMasked  string               `json:"emailMasked"`
	ExpiresAt    string               `json:"expiresAt"`
	Assignments  []AssignmentResponse `json:"assignments"`
	ExistingUser bool                 `json:"existingUser"`
}

type AcceptedResponse struct {
	Status       string `json:"status"`
	MembershipID string `json:"membershipId"`
}
