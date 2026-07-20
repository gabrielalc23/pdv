package organizations

import "time"

type OrganizationInput struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
	Currency string `json:"currency"`
}

type InitialStoreInput struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

type CreateOrganizationRequest struct {
	Organization OrganizationInput `json:"organization"`
	Store        InitialStoreInput `json:"store"`
}

type UpdateOrganizationRequest struct {
	Name     *string `json:"name"`
	Slug     *string `json:"slug"`
	Timezone *string `json:"timezone"`
	Locale   *string `json:"locale"`
	Currency *string `json:"currency"`
}

type ArchiveOrganizationRequest struct {
	Confirm bool `json:"confirm"`
}

type OrganizationResponse struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	Status               string     `json:"status"`
	Timezone             string     `json:"timezone"`
	Locale               string     `json:"locale"`
	Currency             string     `json:"currency"`
	AuthorizationVersion int64      `json:"authorizationVersion"`
	ArchivedAt           *time.Time `json:"archivedAt"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type OrganizationSummaryResponse struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Slug                 string `json:"slug"`
	Status               string `json:"status"`
	Timezone             string `json:"timezone"`
	Locale               string `json:"locale"`
	Currency             string `json:"currency"`
	AuthorizationVersion int64  `json:"authorizationVersion"`
}

type StoreResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Timezone string `json:"timezone,omitempty"`
}

type OrganizationMembershipResponse struct {
	MembershipID                   string                      `json:"membershipId"`
	MembershipStatus               string                      `json:"membershipStatus"`
	MembershipAuthorizationVersion int64                       `json:"membershipAuthorizationVersion"`
	JoinedAt                       time.Time                   `json:"joinedAt"`
	Organization                   OrganizationSummaryResponse `json:"organization"`
	DefaultStore                   *StoreResponse              `json:"defaultStore"`
}

type OrganizationListResponse struct {
	Data []OrganizationMembershipResponse `json:"data"`
}

type StoreListResponse struct {
	Data []StoreResponse `json:"data"`
}

type CreateOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
	MembershipID string               `json:"membershipId"`
	Store        StoreResponse        `json:"store"`
}
