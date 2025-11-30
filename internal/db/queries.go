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

// UpdateWhitelist updates a whitelist entry
func UpdateWhitelist(id int64, enabled bool, notes string) error {
	query := `UPDATE whitelist SET enabled = ?, notes = ? WHERE id = ?`
	_, err := DB.Exec(query, enabled, notes, id)
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
	query := `SELECT id, vmid, resource_name, node, enabled, created_at, created_by, notes
	          FROM whitelist WHERE enabled = 1`

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

// Restart logs functions

// CreateRestartLog creates a new restart log entry
func CreateRestartLog(log *models.RestartLog) (int64, error) {
	query := `INSERT INTO restart_logs 
	          (vmid, resource_name, node, action, trigger_type, triggered_by, status, started_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := DB.Exec(query, log.VMID, log.ResourceName, log.Node,
		log.Action, log.TriggerType, log.TriggeredBy, log.Status, log.StartedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateRestartLog updates a restart log entry
func UpdateRestartLog(log *models.RestartLog) error {
	query := `UPDATE restart_logs 
	          SET status = ?, error_message = ?, completed_at = ?, duration_seconds = ?
	          WHERE id = ?`

	_, err := DB.Exec(query, log.Status, log.ErrorMessage, log.CompletedAt,
		log.DurationSeconds, log.ID)
	return err
}

// GetRestartLogs retrieves restart logs with optional filtering
func GetRestartLogs(filter models.LogsFilter) ([]models.RestartLog, error) {
	query := `SELECT id, vmid, resource_name, node, action, trigger_type, triggered_by,
	          status, error_message, started_at, completed_at, duration_seconds
	          FROM restart_logs WHERE 1=1`

	args := []interface{}{}

	if filter.VMID > 0 {
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
		var completedAt sql.NullTime
		var duration sql.NullInt64

		err := rows.Scan(&log.ID, &log.VMID, &log.ResourceName, &log.Node,
			&log.Action, &log.TriggerType, &log.TriggeredBy, &log.Status,
			&errorMsg, &log.StartedAt, &completedAt, &duration)
		if err != nil {
			return nil, err
		}

		if errorMsg.Valid {
			log.ErrorMessage = errorMsg.String
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
	var lastAutoRestart sql.NullTime
	err = DB.QueryRow(`SELECT MAX(started_at) FROM restart_logs 
	                    WHERE trigger_type = 'auto'`).Scan(&lastAutoRestart)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if lastAutoRestart.Valid {
		status.NextRestartTime = lastAutoRestart.Time.Add(6 * time.Hour)
	} else {
		status.NextRestartTime = time.Now().Add(6 * time.Hour)
	}

	return &status, nil
}
