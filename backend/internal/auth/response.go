package auth

type AuthResponse struct {
	AccessToken string          `json:"accessToken"`
	TokenType   string          `json:"tokenType"`
	ExpiresIn   int64           `json:"expiresIn"`
	User        UserResponse    `json:"user"`
	Session     SessionResponse `json:"session"`
	Context     ContextResponse `json:"context"`
}

type MeResponse struct {
	User    UserResponse    `json:"user"`
	Session SessionResponse `json:"session"`
	Context ContextResponse `json:"context"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	EmailVerified bool   `json:"emailVerified"`
}

type SessionResponse struct {
	ID                string `json:"id"`
	ClientID          string `json:"clientId"`
	CreatedAt         string `json:"createdAt"`
	IdleExpiresAt     string `json:"idleExpiresAt"`
	AbsoluteExpiresAt string `json:"absoluteExpiresAt"`
}

type ContextResponse struct {
	Kind         string                `json:"kind"`
	MembershipID *string               `json:"membershipId"`
	Organization *OrganizationResponse `json:"organization"`
	Store        *StoreResponse        `json:"store"`
	Roles        []string              `json:"roles"`
	Scopes       []string              `json:"scopes"`
}

type OrganizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type StoreResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type SessionsResponse struct {
	Data []SessionListItem `json:"data"`
}

type SessionListItem struct {
	ID                string `json:"id"`
	ClientID          string `json:"clientId"`
	DeviceName        string `json:"deviceName,omitempty"`
	IPAddress         string `json:"ipAddress,omitempty"`
	UserAgent         string `json:"userAgent,omitempty"`
	IsCurrent         bool   `json:"isCurrent"`
	Status            string `json:"status"`
	LastSeenAt        string `json:"lastSeenAt"`
	CreatedAt         string `json:"createdAt"`
	IdleExpiresAt     string `json:"idleExpiresAt"`
	AbsoluteExpiresAt string `json:"absoluteExpiresAt"`
}
