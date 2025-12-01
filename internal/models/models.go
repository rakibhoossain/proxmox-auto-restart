package models

import "time"

// Resource represents a Proxmox VM or Container (real-time data from Proxmox API)
type Resource struct {
	VMID        int     `json:"vmid"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "qemu" or "lxc"
	Node        string  `json:"node"`
	Status      string  `json:"status"`
	Uptime      int64   `json:"uptime"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsed  int64   `json:"memory_used"`
	MemoryTotal int64   `json:"memory_total"`
	DiskUsed    int64   `json:"disk_used"`
	DiskTotal   int64   `json:"disk_total"`
}

// Whitelist represents a VM/Container configured for auto-restart
type Whitelist struct {
	ID                   int64     `json:"id"`
	VMID                 int       `json:"vmid"`
	ResourceName         string    `json:"resource_name"`
	Node                 string    `json:"node"`
	Enabled              bool      `json:"enabled"`
	RestartIntervalHours int       `json:"restart_interval_hours"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            string    `json:"created_by"`
	Notes                string    `json:"notes"`
}

// RestartLog represents a restart operation audit log
type RestartLog struct {
	ID              int64      `json:"id"`
	VMID            int        `json:"vmid"`
	ResourceName    string     `json:"resource_name"`
	Node            string     `json:"node"`
	Action          string     `json:"action"`       // restart, stop, start
	TriggerType     string     `json:"trigger_type"` // auto, manual
	TriggeredBy     string     `json:"triggered_by"`
	Status          string     `json:"status"` // success, failed, pending
	ErrorMessage    string     `json:"error_message,omitempty"`
	Output          string     `json:"output,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	DurationSeconds int64      `json:"duration_seconds,omitempty"`
}

// CreateWhitelistRequest is the request body for adding a VM/Container to whitelist
type CreateWhitelistRequest struct {
	VMID                 int    `json:"vmid"`
	ResourceName         string `json:"resource_name"`
	Node                 string `json:"node"`
	CreatedBy            string `json:"created_by"`
	Notes                string `json:"notes"`
	RestartIntervalHours int    `json:"restart_interval_hours"`
}

// UpdateWhitelistRequest is the request body for updating a whitelist entry
type UpdateWhitelistRequest struct {
	Enabled              bool   `json:"enabled"`
	Notes                string `json:"notes"`
	RestartIntervalHours int    `json:"restart_interval_hours"`
}

// ResourceActionRequest is the request body for resource actions
type ResourceActionRequest struct {
	TriggeredBy string `json:"triggered_by"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	TotalResources   int       `json:"total_resources"`
	RunningResources int       `json:"running_resources"`
	WhitelistedCount int       `json:"whitelisted_count"`
	NextRestartTime  time.Time `json:"next_restart_time"`
	TotalRestarts    int64     `json:"total_restarts"`
	FailedRestarts   int64     `json:"failed_restarts"`
}

// LogsFilter represents filtering options for logs
type LogsFilter struct {
	VMID         int
	ResourceName string
	Node         string
	Action       string
	TriggerType  string
	Status       string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// ContainerService represents a service installed on a container
type ContainerService struct {
	ID              int64     `json:"id"`
	VMID            int       `json:"vmid"`
	Node            string    `json:"node"`
	ServiceName     string    `json:"service_name"`
	ServiceType     string    `json:"service_type"` // grow, connect, custom
	InstallCommands string    `json:"install_commands,omitempty"`
	InstalledAt     time.Time `json:"installed_at"`
}
