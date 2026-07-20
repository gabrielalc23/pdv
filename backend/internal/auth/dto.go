package auth

type RegisterRequest struct {
	Email        string              `json:"email"`
	Password     string              `json:"password"`
	DisplayName  string              `json:"displayName"`
	Organization OrganizationRequest `json:"organization"`
	Store        StoreRequest        `json:"store"`
	ClientID     string              `json:"clientId"`
	DeviceName   string              `json:"deviceName"`
}

type OrganizationRequest struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
	Currency string `json:"currency"`
}

type StoreRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

type LoginRequest struct {
	Email          string  `json:"email"`
	Password       string  `json:"password"`
	ClientID       string  `json:"clientId"`
	DeviceName     string  `json:"deviceName"`
	OrganizationID *string `json:"organizationId"`
	StoreID        *string `json:"storeId"`
}

type ContextRequest struct {
	OrganizationID *string `json:"organizationId"`
	StoreID        *string `json:"storeId"`
}

type VerificationRequiredResponse struct {
	Status string `json:"status"`
}

type CSRFResponse struct {
	CSRFToken string `json:"csrfToken"`
}

type EmailActionRequest struct {
	Email string `json:"email"`
}

type PasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type EmailVerifyRequest struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type UpdateMeRequest struct {
	DisplayName string `json:"displayName"`
}

type AcceptedResponse struct {
	Status string `json:"status"`
}
