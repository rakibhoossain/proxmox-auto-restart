package db

import (
	"database/sql"
	"log"
)

// RunMigrations creates all necessary database tables and indexes
// Only whitelist and restart_logs - no resources table (fetched real-time from Proxmox)
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Create whitelist table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS whitelist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vmid INTEGER NOT NULL,
			resource_name TEXT NOT NULL,
			node TEXT NOT NULL,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT NOT NULL,
			notes TEXT,
			UNIQUE(vmid, node)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for whitelist
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_whitelist_vmid ON whitelist(vmid)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON whitelist(enabled)`)
	if err != nil {
		return err
	}

	// Create restart_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS restart_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vmid INTEGER NOT NULL,
			resource_name TEXT NOT NULL,
			node TEXT NOT NULL,
			action TEXT NOT NULL,
			trigger_type TEXT NOT NULL,
			triggered_by TEXT NOT NULL,
			status TEXT NOT NULL,
			error_message TEXT,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			duration_seconds INTEGER
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes for restart_logs
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_restart_logs_vmid ON restart_logs(vmid)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_restart_logs_status ON restart_logs(status)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_restart_logs_started_at ON restart_logs(started_at DESC)`)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
