package api

import (
	"encoding/json"
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

	err = db.UpdateWhitelist(id, req.Enabled, req.Notes)
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

	logs, err := db.GetRestartLogs(filter)
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

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":            "ok",
		"proxmox_available": proxmox.IsProxmoxInstalled(),
		"timestamp":         time.Now(),
	}
	respondJSON(w, http.StatusOK, health)
}
