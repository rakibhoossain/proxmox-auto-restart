package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rakib/proxmox-auto-restart/internal/api"
	"github.com/rakib/proxmox-auto-restart/internal/db"
	"github.com/rakib/proxmox-auto-restart/internal/scheduler"
)

func main() {
	log.Println("Starting Proxmox Auto-Restart Service...")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	dbPath := db.GetDBPath()
	log.Printf("Using database: %s", dbPath)

	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Get database handle
	database := db.GetDB()

	// Run migrations (only whitelist and restart_logs tables)
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start auto-restart scheduler (NO sync scheduler - data fetched real-time)
	log.Println("Starting auto-restart scheduler...")
	if err := scheduler.StartRestartScheduler(""); err != nil {
		log.Fatalf("Failed to start restart scheduler: %v", err)
	}

	// Setup HTTP server
	router := api.SetupRoutes()

	// Start HTTP server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("HTTP server starting on %s", addr)

	// Setup graceful shutdown
	go func() {
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Println("✓ Proxmox VM/Container Auto-Restart Service is running")
	log.Println("✓ VM/Container data: fetched real-time from Proxmox")
	log.Println("✓ Auto-restart: every 6 hours")
	log.Printf("✓ API available at http://localhost:%s", port)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	log.Println("Shutting down server...")
	scheduler.StopRestartScheduler()
	log.Println("Service stopped")
}
