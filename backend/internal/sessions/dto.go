package sessions

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
