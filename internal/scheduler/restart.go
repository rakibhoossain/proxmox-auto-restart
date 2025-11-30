package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/rakib/proxmox-auto-restart/internal/db"
	"github.com/rakib/proxmox-auto-restart/internal/models"
	"github.com/rakib/proxmox-auto-restart/internal/proxmox"
	"github.com/robfig/cron/v3"
)

var restartCron *cron.Cron

// StartRestartScheduler starts the auto-restart scheduler for whitelisted VMs/Containers
func StartRestartScheduler(interval string) error {
	restartCron = cron.New()

	// Default: every 6 hours
	if interval == "" {
		interval = "@every 6h"
	}

	_, err := restartCron.AddFunc(interval, restartWhitelistedResources)
	if err != nil {
		return err
	}

	restartCron.Start()
	log.Printf("Auto-restart scheduler started (interval: %s, next restart: %s)",
		interval, time.Now().Add(6*time.Hour).Format(time.RFC3339))
	return nil
}

// StopRestartScheduler stops the auto-restart scheduler
func StopRestartScheduler() {
	if restartCron != nil {
		restartCron.Stop()
		log.Println("Auto-restart scheduler stopped")
	}
}

// restartWhitelistedResources restarts all enabled whitelisted VMs/Containers
func restartWhitelistedResources() {
	log.Println("Starting auto-restart of whitelisted resources...")

	// Get enabled whitelist entries
	whitelisted, err := db.GetEnabledWhitelist()
	if err != nil {
		log.Printf("ERROR: Failed to get whitelist: %v", err)
		return
	}

	if len(whitelisted) == 0 {
		log.Println("No whitelisted resources to restart")
		return
	}

	log.Printf("Found %d whitelisted resource(s) for auto-restart", len(whitelisted))

	// Fetch current resources to get type
	resources, err := proxmox.GetAllResources()
	if err != nil {
		log.Printf("ERROR: Failed to fetch resources from Proxmox: %v", err)
		return
	}

	// Create VMID to type map
	typeMap := make(map[int]string)
	for _, r := range resources {
		typeMap[r.VMID] = r.Type
	}

	for _, wl := range whitelisted {
		resourceType, exists := typeMap[wl.VMID]
		if !exists {
			log.Printf("WARNING: Resource %d (%s) not found in Proxmox, skipping", wl.VMID, wl.ResourceName)
			continue
		}

		log.Printf("Auto-restarting resource: %s (VMID: %d, Type: %s, Node: %s)",
			wl.ResourceName, wl.VMID, resourceType, wl.Node)

		restartResource(wl.VMID, wl.ResourceName, wl.Node, resourceType, "auto", "system")
	}

	log.Println("Auto-restart cycle completed")
}

// restartResource restarts a specific VM/Container and logs the operation
func restartResource(vmid int, resourceName, node, resourceType, triggerType, triggeredBy string) {
	// Create log entry
	logEntry := &models.RestartLog{
		VMID:         vmid,
		ResourceName: resourceName,
		Node:         node,
		Action:       "restart",
		TriggerType:  triggerType,
		TriggeredBy:  triggeredBy,
		Status:       "pending",
		StartedAt:    time.Now(),
	}

	logID, err := db.CreateRestartLog(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to create restart log for %d: %v", vmid, err)
		return
	}

	logEntry.ID = logID

	// Execute restart
	startTime := time.Now()
	err = proxmox.RestartResource(node, vmid, resourceType)
	duration := time.Since(startTime).Seconds()

	// Update log entry
	completedAt := time.Now()
	logEntry.CompletedAt = &completedAt
	logEntry.DurationSeconds = int64(duration)

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
		log.Printf("ERROR: Failed to restart resource %d (%s): %v", vmid, resourceName, err)
	} else {
		logEntry.Status = "success"
		log.Printf("Successfully restarted resource %d (%s) in %.2fs", vmid, resourceName, duration)
	}

	if err := db.UpdateRestartLog(logEntry); err != nil {
		log.Printf("ERROR: Failed to update restart log: %v", err)
	}
}

// ManualRestartResource handles manual restart requests
func ManualRestartResource(vmid int, node, triggeredBy string) error {
	// Get resource type from Proxmox
	resource, err := proxmox.GetResource(node, vmid)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	if resource == nil {
		return fmt.Errorf("resource %d not found on node %s", vmid, node)
	}

	log.Printf("Manual restart requested for %s (VMID: %d, Type: %s) by %s",
		resource.Name, vmid, resource.Type, triggeredBy)

	go restartResource(vmid, resource.Name, node, resource.Type, "manual", triggeredBy)
	return nil
}

// stopResource stops a specific VM/Container and logs the operation
func stopResource(vmid int, resourceName, node, resourceType, triggerType, triggeredBy string) {
	// Create log entry
	logEntry := &models.RestartLog{
		VMID:         vmid,
		ResourceName: resourceName,
		Node:         node,
		Action:       "stop",
		TriggerType:  triggerType,
		TriggeredBy:  triggeredBy,
		Status:       "pending",
		StartedAt:    time.Now(),
	}

	logID, err := db.CreateRestartLog(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to create stop log for %d: %v", vmid, err)
		return
	}

	logEntry.ID = logID

	// Execute stop
	startTime := time.Now()
	err = proxmox.StopResource(node, vmid, resourceType)
	duration := time.Since(startTime).Seconds()

	// Update log entry
	completedAt := time.Now()
	logEntry.CompletedAt = &completedAt
	logEntry.DurationSeconds = int64(duration)

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
		log.Printf("ERROR: Failed to stop resource %d (%s): %v", vmid, resourceName, err)
	} else {
		logEntry.Status = "success"
		log.Printf("Successfully stopped resource %d (%s) in %.2fs", vmid, resourceName, duration)
	}

	if err := db.UpdateRestartLog(logEntry); err != nil {
		log.Printf("ERROR: Failed to update stop log: %v", err)
	}
}

// ManualStopResource handles manual stop requests
func ManualStopResource(vmid int, node, triggeredBy string) error {
	// Get resource type from Proxmox
	resource, err := proxmox.GetResource(node, vmid)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	if resource == nil {
		return fmt.Errorf("resource %d not found on node %s", vmid, node)
	}

	log.Printf("Manual stop requested for %s (VMID: %d, Type: %s) by %s",
		resource.Name, vmid, resource.Type, triggeredBy)

	go stopResource(vmid, resource.Name, node, resource.Type, "manual", triggeredBy)
	return nil
}

// startResource starts a specific VM/Container and logs the operation
func startResource(vmid int, resourceName, node, resourceType, triggerType, triggeredBy string) {
	// Create log entry
	logEntry := &models.RestartLog{
		VMID:         vmid,
		ResourceName: resourceName,
		Node:         node,
		Action:       "start",
		TriggerType:  triggerType,
		TriggeredBy:  triggeredBy,
		Status:       "pending",
		StartedAt:    time.Now(),
	}

	logID, err := db.CreateRestartLog(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to create start log for %d: %v", vmid, err)
		return
	}

	logEntry.ID = logID

	// Execute start
	startTime := time.Now()
	err = proxmox.StartResource(node, vmid, resourceType)
	duration := time.Since(startTime).Seconds()

	// Update log entry
	completedAt := time.Now()
	logEntry.CompletedAt = &completedAt
	logEntry.DurationSeconds = int64(duration)

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
		log.Printf("ERROR: Failed to start resource %d (%s): %v", vmid, resourceName, err)
	} else {
		logEntry.Status = "success"
		log.Printf("Successfully started resource %d (%s) in %.2fs", vmid, resourceName, duration)
	}

	if err := db.UpdateRestartLog(logEntry); err != nil {
		log.Printf("ERROR: Failed to update start log: %v", err)
	}
}

// ManualStartResource handles manual start requests
func ManualStartResource(vmid int, node, triggeredBy string) error {
	// Get resource type from Proxmox
	resource, err := proxmox.GetResource(node, vmid)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	if resource == nil {
		return fmt.Errorf("resource %d not found on node %s", vmid, node)
	}

	log.Printf("Manual start requested for %s (VMID: %d, Type: %s) by %s",
		resource.Name, vmid, resource.Type, triggeredBy)

	go startResource(vmid, resource.Name, node, resource.Type, "manual", triggeredBy)
	return nil
}
