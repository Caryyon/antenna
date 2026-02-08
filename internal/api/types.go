package api

// Session represents a monitored OpenClaw session.
type Session struct {
	SessionID    string  `json:"sessionId"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Model        string  `json:"model"`
	MessageCount int     `json:"messageCount"`
	TotalCost    float64 `json:"totalCost"`
	TodayCost    float64 `json:"todayCost"`
	UpdatedAt    int64   `json:"updatedAt"`
	IsActive     bool    `json:"isActive"`
}

// DashboardData is the full dashboard response.
type DashboardData struct {
	Sessions   []Session `json:"sessions"`
	TotalCount int       `json:"totalCount"`
	TotalCost  float64   `json:"totalCost"`
	TodayCost  float64   `json:"todayCost"`
}

// HourlyBucket represents activity in one hour.
type HourlyBucket struct {
	Hour     string  `json:"hour"`
	Messages int     `json:"messages"`
	Cost     float64 `json:"cost"`
}
