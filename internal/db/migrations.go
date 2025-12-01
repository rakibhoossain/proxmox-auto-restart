package db

import (
	"database/sql"
	"fmt"
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
			restart_interval_hours INTEGER DEFAULT 6,
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
			output TEXT,
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

	// Create container_services table for tracking installed services
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS container_services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vmid INTEGER NOT NULL,
			node TEXT NOT NULL,
			service_name TEXT NOT NULL,
			service_type TEXT NOT NULL,
			install_commands TEXT,
			installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(vmid, node, service_name)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for container_services
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_container_services_vmid ON container_services(vmid, node)`)
	if err != nil {
		return err
	}

	// Schema Updates: Add missing columns if they don't exist (for existing databases)

	// 1. Add restart_interval_hours to whitelist
	_, err = db.Exec(`ALTER TABLE whitelist ADD COLUMN restart_interval_hours INTEGER DEFAULT 6`)
	if err != nil {
		// Ignore error if column already exists (SQLite doesn't support IF NOT EXISTS for ADD COLUMN directly in all versions/drivers easily without checking,
		// but typically returns an error we can ignore if it says "duplicate column name")
		// For simplicity in this driver, we'll assume it might fail if exists.
		// A better way is to check pragma_table_info, but for this quick fix, we can just log/ignore or wrap in a separate check.
		// Let's try a safer approach by checking if column exists first?
		// Or just let it fail silently? SQLite error for duplicate column is usually safe to ignore in this context if we just want to ensure it's there.
		// However, Go's sql package returns an error.
		// Let's use a helper or just try-catch style.
		// Since we can't easily check error string cross-platform/driver without overhead, let's just run it and ignore specific error?
		// No, let's do it properly.
	}

	// Better approach: Check if column exists
	if !columnExists(db, "whitelist", "restart_interval_hours") {
		_, err = db.Exec(`ALTER TABLE whitelist ADD COLUMN restart_interval_hours INTEGER DEFAULT 6`)
		if err != nil {
			log.Printf("WARNING: Failed to add restart_interval_hours column: %v", err)
		} else {
			log.Println("Added restart_interval_hours column to whitelist table")
		}
	}

	// 2. Add output to restart_logs
	if !columnExists(db, "restart_logs", "output") {
		_, err = db.Exec(`ALTER TABLE restart_logs ADD COLUMN output TEXT`)
		if err != nil {
			log.Printf("WARNING: Failed to add output column: %v", err)
		} else {
			log.Println("Added output column to restart_logs table")
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func columnExists(db *sql.DB, tableName, columnName string) bool {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false
		}
		if name == columnName {
			return true
		}
	}
	return false
}
