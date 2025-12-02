package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rakib/proxmox-auto-restart/internal/models"
)

// No resource queries - fetch real-time from Proxmox

// Whitelist functions

// GetAllWhitelist retrieves all whitelist entries
func GetAllWhitelist() ([]models.Whitelist, error) {
	query := `SELECT id, vmid, resource_name, node, enabled, created_at, created_by, notes 
	          FROM whitelist ORDER BY vmid ASC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var whitelist []models.Whitelist
	for rows.Next() {
		var wl models.Whitelist
		err := rows.Scan(&wl.ID, &wl.VMID, &wl.ResourceName, &wl.Node,
			&wl.Enabled, &wl.CreatedAt, &wl.CreatedBy, &wl.Notes)
		if err != nil {
			return nil, err
		}
		whitelist = append(whitelist, wl)
	}

	return whitelist, nil
}

// GetWhitelistByID retrieves a whitelist entry by ID
func GetWhitelistByID(id int64) (*models.Whitelist, error) {
	query := `SELECT id, vmid, resource_name, node, enabled, created_at, created_by, notes
	          FROM whitelist WHERE id = ?`

	var wl models.Whitelist
	err := DB.QueryRow(query, id).Scan(&wl.ID, &wl.VMID, &wl.ResourceName, &wl.Node,
		&wl.Enabled, &wl.CreatedAt, &wl.CreatedBy, &wl.Notes)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &wl, nil
}

// AddToWhitelist adds a VM/Container to the whitelist
func AddToWhitelist(vmid int, resourceName, node, createdBy, notes string) error {
	query := `INSERT INTO whitelist (vmid, resource_name, node, enabled, created_by, notes)
	          VALUES (?, ?, ?, 1, ?, ?)`

	_, err := DB.Exec(query, vmid, resourceName, node, createdBy, notes)
	return err
}

// DeleteFromWhitelist removes an entry from the whitelist
func DeleteFromWhitelist(id int64) error {
	query := `DELETE FROM whitelist WHERE id = ?`
	_, err := DB.Exec(query, id)
	return err
}

// GetEnabledWhitelist retrieves all enabled whitelist entries
func GetEnabledWhitelist() ([]models.Whitelist, error) {
	query := `SELECT id, vmid, resource_name, node, enabled, restart_interval_hours, created_at, created_by, notes 
	          FROM whitelist WHERE enabled = 1`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var whitelist []models.Whitelist
	for rows.Next() {
		var w models.Whitelist
		err := rows.Scan(&w.ID, &w.VMID, &w.ResourceName, &w.Node, &w.Enabled, &w.RestartIntervalHours,
			&w.CreatedAt, &w.CreatedBy, &w.Notes)
		if err != nil {
			return nil, err
		}
		whitelist = append(whitelist, w)
	}
	return whitelist, nil
}

// CreateWhitelist adds a new entry to the whitelist
func CreateWhitelist(req *models.CreateWhitelistRequest) error {
	query := `INSERT INTO whitelist (vmid, resource_name, node, created_by, notes, restart_interval_hours) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	// Default to 6 hours if not specified
	interval := req.RestartIntervalHours
	if interval < 1 {
		interval = 6
	}

	_, err := DB.Exec(query, req.VMID, req.ResourceName, req.Node, req.CreatedBy, req.Notes, interval)
	return err
}

// UpdateWhitelist updates an existing whitelist entry
func UpdateWhitelist(vmid int, req *models.UpdateWhitelistRequest) error {
	query := `UPDATE whitelist SET enabled = ?, notes = ?, restart_interval_hours = ? WHERE vmid = ?`

	// Default to 6 hours if not specified
	interval := req.RestartIntervalHours
	if interval < 1 {
		interval = 6
	}

	_, err := DB.Exec(query, req.Enabled, req.Notes, interval, vmid)
	return err
}

// DeleteWhitelistByVMID removes a whitelist entry by VMID
func DeleteWhitelistByVMID(vmid int) error {
	query := `DELETE FROM whitelist WHERE vmid = ?`
	_, err := DB.Exec(query, vmid)
	return err
}

// Restart logs functions

// CreateRestartLog creates a new restart log entry
func CreateRestartLog(log *models.RestartLog) (int64, error) {
	query := `INSERT INTO restart_logs (vmid, resource_name, node, action, trigger_type, triggered_by, status, started_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, log.VMID, log.ResourceName, log.Node, log.Action, log.TriggerType, log.TriggeredBy, log.Status, log.StartedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateRestartLog updates an existing restart log entry with completion details
func UpdateRestartLog(log *models.RestartLog) error {
	query := `UPDATE restart_logs 
	          SET status = ?, error_message = ?, output = ?, completed_at = ?, duration_seconds = ? 
	          WHERE id = ?`
	_, err := DB.Exec(query, log.Status, log.ErrorMessage, log.Output, log.CompletedAt, log.DurationSeconds, log.ID)
	return err
}

// GetLogs retrieves logs with filtering and pagination
func GetLogs(filter models.LogsFilter) ([]models.RestartLog, error) {
	query := `SELECT id, vmid, resource_name, node, action, trigger_type, triggered_by, status, error_message, output, started_at, completed_at, duration_seconds 
	          FROM restart_logs WHERE 1=1`
	args := []interface{}{}

	if filter.VMID != 0 {
		query += " AND vmid = ?"
		args = append(args, filter.VMID)
	}
	if filter.ResourceName != "" {
		query += " AND resource_name LIKE ?"
		args = append(args, "%"+filter.ResourceName+"%")
	}
	if filter.Node != "" {
		query += " AND node = ?"
		args = append(args, filter.Node)
	}
	if filter.Action != "" {
		query += " AND action = ?"
		args = append(args, filter.Action)
	}
	if filter.TriggerType != "" {
		query += " AND trigger_type = ?"
		args = append(args, filter.TriggerType)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.StartDate != nil {
		query += " AND started_at >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND started_at <= ?"
		args = append(args, filter.EndDate)
	}

	query += " ORDER BY started_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.RestartLog
	for rows.Next() {
		var log models.RestartLog
		var errorMsg sql.NullString
		var output sql.NullString
		var completedAt sql.NullTime
		var duration sql.NullInt64

		err := rows.Scan(&log.ID, &log.VMID, &log.ResourceName, &log.Node,
			&log.Action, &log.TriggerType, &log.TriggeredBy, &log.Status,
			&errorMsg, &output, &log.StartedAt, &completedAt, &duration)
		if err != nil {
			return nil, err
		}

		if errorMsg.Valid {
			log.ErrorMessage = errorMsg.String
		}
		if output.Valid {
			log.Output = output.String
		}
		if completedAt.Valid {
			t := completedAt.Time
			log.CompletedAt = &t
		}
		if duration.Valid {
			log.DurationSeconds = duration.Int64
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetSystemStatus retrieves aggregated system status
func GetSystemStatus() (*models.SystemStatus, error) {
	var status models.SystemStatus

	// Get whitelisted count
	err := DB.QueryRow(`SELECT COUNT(*) FROM whitelist WHERE enabled = 1`).
		Scan(&status.WhitelistedCount)
	if err != nil {
		return nil, err
	}

	// Get total and failed restarts
	err = DB.QueryRow(`SELECT 
	                     COUNT(*),
	                     COUNT(CASE WHEN status = 'failed' THEN 1 END)
	                     FROM restart_logs`).
		Scan(&status.TotalRestarts, &status.FailedRestarts)
	if err != nil {
		return nil, err
	}

	// Calculate next restart time (6 hours from last auto restart)
	var lastAutoRestartStr sql.NullString
	err = DB.QueryRow(`SELECT MAX(started_at) FROM restart_logs 
	                    WHERE trigger_type = 'auto'`).Scan(&lastAutoRestartStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if lastAutoRestartStr.Valid && lastAutoRestartStr.String != "" {
		// Parse the datetime string from SQLite
		lastRestart, err := time.Parse("2006-01-02 15:04:05", lastAutoRestartStr.String)
		if err != nil {
			// Try RFC3339 format as fallback
			lastRestart, err = time.Parse(time.RFC3339, lastAutoRestartStr.String)
			if err != nil {
				// If parsing fails, use current time
				status.NextRestartTime = time.Now().Add(6 * time.Hour)
			} else {
				status.NextRestartTime = lastRestart.Add(6 * time.Hour)
			}
		} else {
			status.NextRestartTime = lastRestart.Add(6 * time.Hour)
		}
	} else {
		status.NextRestartTime = time.Now().Add(6 * time.Hour)
	}

	return &status, nil
}

// Container Service functions

// CreateContainerService records a service installation on a container
func CreateContainerService(vmid int, node, serviceName, serviceType, installCommands string) error {
	query := `INSERT INTO container_services (vmid, node, service_name, service_type, install_commands)
	          VALUES (?, ?, ?, ?, ?)`

	_, err := DB.Exec(query, vmid, node, serviceName, serviceType, installCommands)
	return err
}

// GetServicesByVMID retrieves all services installed on a container
func GetServicesByVMID(vmid int, node string) ([]models.ContainerService, error) {
	query := `SELECT id, vmid, node, service_name, service_type, install_commands, installed_at
	          FROM container_services
	          WHERE vmid = ? AND node = ?
	          ORDER BY installed_at DESC`

	rows, err := DB.Query(query, vmid, node)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.ContainerService
	for rows.Next() {
		var svc models.ContainerService
		var installCommands sql.NullString

		err := rows.Scan(&svc.ID, &svc.VMID, &svc.Node, &svc.ServiceName,
			&svc.ServiceType, &installCommands, &svc.InstalledAt)
		if err != nil {
			return nil, err
		}

		if installCommands.Valid {
			svc.InstallCommands = installCommands.String
		}

		services = append(services, svc)
	}

	return services, nil
}

// DeleteServicesByVMID removes all service records for a container
func DeleteServicesByVMID(vmid int, node string) error {
	query := `DELETE FROM container_services WHERE vmid = ? AND node = ?`
	_, err := DB.Exec(query, vmid, node)
	return err
}

// GetLastRestartTime retrieves the timestamp of the last successful restart for a VMID
func GetLastRestartTime(vmid int) (time.Time, error) {
	query := `SELECT completed_at FROM restart_logs 
	          WHERE vmid = ? AND action = 'restart' AND status = 'success' 
	          ORDER BY completed_at DESC LIMIT 1`

	var lastRestart sql.NullTime
	err := DB.QueryRow(query, vmid).Scan(&lastRestart)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil // Never restarted
		}
		return time.Time{}, err
	}

	if lastRestart.Valid {
		return lastRestart.Time, nil
	}
	return time.Time{}, nil
}
