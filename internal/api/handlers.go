package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rakib/proxmox-auto-restart/internal/db"
	"github.com/rakib/proxmox-auto-restart/internal/models"
	"github.com/rakib/proxmox-auto-restart/internal/proxmox"
	"github.com/rakib/proxmox-auto-restart/internal/scheduler"
)

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

// Resource handlers (VMs and Containers) - Real-time data from Proxmox

func GetResources(w http.ResponseWriter, r *http.Request) {
	// Fetch real-time from Proxmox
	resources, err := proxmox.GetAllResources()
	if err != nil {
		log.Printf("ERROR: Failed to get resources from Proxmox: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get resources from Proxmox")
		return
	}
	respondJSON(w, http.StatusOK, resources)
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	// Fetch real-time from Proxmox
	resource, err := proxmox.GetResource(node, vmid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get resource")
		return
	}
	if resource == nil {
		respondError(w, http.StatusNotFound, "Resource not found")
		return
	}

	respondJSON(w, http.StatusOK, resource)
}

func RestartResource(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	var req models.ResourceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.TriggeredBy = "unknown"
	}

	if req.TriggeredBy == "" {
		req.TriggeredBy = "api"
	}

	// Check if Proxmox is installed
	if !proxmox.IsProxmoxInstalled() {
		respondError(w, http.StatusServiceUnavailable, "Proxmox not available on this server")
		return
	}

	// Trigger restart asynchronously
	err = scheduler.ManualRestartResource(vmid, node, req.TriggeredBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Restart triggered",
		"vmid":    vmid,
		"node":    node,
	})
}

func StopResource(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	var req models.ResourceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.TriggeredBy = "unknown"
	}

	if req.TriggeredBy == "" {
		req.TriggeredBy = "api"
	}

	// Check if Proxmox is installed
	if !proxmox.IsProxmoxInstalled() {
		respondError(w, http.StatusServiceUnavailable, "Proxmox not available on this server")
		return
	}

	err = scheduler.ManualStopResource(vmid, node, req.TriggeredBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Stop triggered",
		"vmid":    vmid,
		"node":    node,
	})
}

func StartResource(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	var req models.ResourceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.TriggeredBy = "unknown"
	}

	if req.TriggeredBy == "" {
		req.TriggeredBy = "api"
	}

	// Check if Proxmox is installed
	if !proxmox.IsProxmoxInstalled() {
		respondError(w, http.StatusServiceUnavailable, "Proxmox not available on this server")
		return
	}

	err = scheduler.ManualStartResource(vmid, node, req.TriggeredBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Start triggered",
		"vmid":    vmid,
		"node":    node,
	})
}

// Whitelist handlers

func GetWhitelist(w http.ResponseWriter, r *http.Request) {
	whitelist, err := db.GetAllWhitelist()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get whitelist")
		return
	}
	respondJSON(w, http.StatusOK, whitelist)
}

func AddToWhitelist(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.VMID == 0 {
		respondError(w, http.StatusBadRequest, "vmid is required")
		return
	}
	if req.ResourceName == "" {
		respondError(w, http.StatusBadRequest, "resource_name is required")
		return
	}
	if req.Node == "" {
		respondError(w, http.StatusBadRequest, "node is required")
		return
	}

	if req.CreatedBy == "" {
		req.CreatedBy = "api"
	}

	err := db.AddToWhitelist(req.VMID, req.ResourceName, req.Node, req.CreatedBy, req.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add to whitelist")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":       "Added to whitelist successfully",
		"vmid":          req.VMID,
		"resource_name": req.ResourceName,
		"node":          req.Node,
	})
}

func UpdateWhitelist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req models.UpdateWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = db.UpdateWhitelist(int(id), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update whitelist")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Updated successfully"})
}

func DeleteFromWhitelist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = db.DeleteFromWhitelist(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete from whitelist")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}

// Logs handlers

func GetLogs(w http.ResponseWriter, r *http.Request) {
	filter := models.LogsFilter{
		Action:      r.URL.Query().Get("action"),
		TriggerType: r.URL.Query().Get("trigger_type"),
		Status:      r.URL.Query().Get("status"),
		Limit:       100, // Default limit
		Offset:      0,
	}

	// Parse VMID
	if vmidStr := r.URL.Query().Get("vmid"); vmidStr != "" {
		if vmid, err := strconv.Atoi(vmidStr); err == nil {
			filter.VMID = vmid
		}
	}

	// Parse resource_name
	if resourceName := r.URL.Query().Get("resource_name"); resourceName != "" {
		filter.ResourceName = resourceName
	}

	// Parse node
	if node := r.URL.Query().Get("node"); node != "" {
		filter.Node = node
	}

	// Parse limit and offset
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// Parse date filters
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	logs, err := db.GetLogs(filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get logs")
		return
	}

	respondJSON(w, http.StatusOK, logs)
}

// System handlers

func GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := db.GetSystemStatus()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get system status")
		return
	}

	// Get real-time resource counts from Proxmox
	resources, err := proxmox.GetAllResources()
	if err == nil {
		status.TotalResources = len(resources)
		runningCount := 0
		for _, r := range resources {
			if r.Status == "running" {
				runningCount++
			}
		}
		status.RunningResources = runningCount
	}

	respondJSON(w, http.StatusOK, status)
}

// Container Management Handlers

func CloneContainerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceVMID int    `json:"source_vmid"`
		NewVMID    int    `json:"new_vmid"`
		TargetNode string `json:"target_node"`
		Hostname   string `json:"hostname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceVMID == 0 || req.NewVMID == 0 || req.TargetNode == "" {
		respondError(w, http.StatusBadRequest, "source_vmid, new_vmid, and target_node are required")
		return
	}

	// Clone the container
	if err := proxmox.CloneContainer(req.SourceVMID, req.NewVMID, req.TargetNode, req.Hostname); err != nil {
		log.Printf("ERROR: Failed to clone container: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Container cloned successfully",
		"source_vmid": req.SourceVMID,
		"new_vmid":    req.NewVMID,
		"target_node": req.TargetNode,
	})
}

func DeleteContainerHandler(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	// Delete the container
	if err := proxmox.DeleteContainer(vmid, node); err != nil {
		log.Printf("ERROR: Failed to delete container: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Remove from whitelist if exists
	_ = db.DeleteWhitelistByVMID(vmid)

	// Remove service records
	_ = db.DeleteServicesByVMID(vmid, node)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Container deleted successfully",
		"vmid":    vmid,
		"node":    node,
	})
}

func DeployBlockchainNodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceVMID int      `json:"source_vmid"`
		NewVMID    int      `json:"new_vmid"`
		TargetNode string   `json:"target_node"`
		Hostname   string   `json:"hostname"`
		Commands   []string `json:"commands"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceVMID == 0 || req.NewVMID == 0 || req.TargetNode == "" {
		respondError(w, http.StatusBadRequest, "source_vmid, new_vmid, and target_node are required")
		return
	}

	if len(req.Commands) == 0 {
		respondError(w, http.StatusBadRequest, "commands array is required")
		return
	}

	// Deploy blockchain node
	if err := proxmox.DeployBlockchainNode(req.SourceVMID, req.NewVMID, req.TargetNode, req.Hostname, req.Commands); err != nil {
		log.Printf("ERROR: Failed to deploy blockchain node: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Record service installation in database
	serviceName := req.Hostname
	if serviceName == "" {
		serviceName = fmt.Sprintf("blockchain-node-%d", req.NewVMID)
	}
	// Determine service type from commands
	serviceType := "custom"
	commandStr := ""
	if len(req.Commands) > 0 {
		commandStr = req.Commands[0]
		if len(commandStr) > 0 {
			if contains(commandStr, "growblockchain") {
				serviceType = "grow"
			} else if contains(commandStr, "connectblockchain") {
				serviceType = "connect"
			}
		}
	}

	// Store all commands as JSON-like string
	allCommands := ""
	for i, cmd := range req.Commands {
		if i > 0 {
			allCommands += "\n"
		}
		allCommands += cmd
	}

	_ = db.CreateContainerService(req.NewVMID, req.TargetNode, serviceName, serviceType, allCommands)

	// Log the deployment
	now := time.Now()
	_, _ = db.CreateRestartLog(&models.RestartLog{
		VMID:         req.NewVMID,
		ResourceName: req.Hostname,
		Node:         req.TargetNode,
		Action:       "deploy",
		TriggerType:  "manual",
		TriggeredBy:  "api",
		Status:       "success",
		StartedAt:    now,
		CompletedAt:  &now,
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Blockchain node deployed successfully",
		"new_vmid":    req.NewVMID,
		"target_node": req.TargetNode,
		"hostname":    req.Hostname,
	})
}

func GetNextAvailableVMID(w http.ResponseWriter, r *http.Request) {
	// Fetch all resources to find max VMID
	resources, err := proxmox.GetAllResources()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch resources")
		return
	}

	maxVMID := 100 // Start from 100 if no resources exist
	for _, resource := range resources {
		if resource.VMID > maxVMID {
			maxVMID = resource.VMID
		}
	}

	nextVMID := maxVMID + 1

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"suggested_vmid": nextVMID,
		"max_vmid":       maxVMID,
	})
}

func GetContainerServicesHandler(w http.ResponseWriter, r *http.Request) {
	vmidStr := chi.URLParam(r, "vmid")
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VMID")
		return
	}

	node := r.URL.Query().Get("node")
	if node == "" {
		respondError(w, http.StatusBadRequest, "node query parameter is required")
		return
	}

	services, err := db.GetServicesByVMID(vmid, node)
	if err != nil {
		log.Printf("ERROR: Failed to get services: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get services")
		return
	}

	respondJSON(w, http.StatusOK, services)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":            "ok",
		"proxmox_available": proxmox.IsProxmoxInstalled(),
		"timestamp":         time.Now(),
	}
	respondJSON(w, http.StatusOK, health)
}
