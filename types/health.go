package types

type MemoryStats struct {
	Used  string `json:"used"`
	Total string `json:"total"`
	Free  string `json:"free"`
	Usage string `json:"usage"`
}

type DatabaseStatus struct {
	Connected   bool   `json:"connected"`
	LastChecked string `json:"last_checked"`
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Memory    MemoryStats            `json:"memory"`
	Database  DatabaseStatus         `json:"database"`
	Tasks     map[string]*TaskStatus `json:"tasks"`
}
