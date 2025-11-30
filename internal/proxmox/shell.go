package proxmox

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rakib/proxmox-auto-restart/internal/models"
)

// IsProxmoxInstalled checks if pvesh command is available
func IsProxmoxInstalled() bool {
	cmd := exec.Command("which", "pvesh")
	err := cmd.Run()
	return err == nil
}

// ProxmoxResource represents a VM or Container from cluster resources API
type ProxmoxResource struct {
	VMID    json.Number `json:"vmid"`
	Name    string      `json:"name"`
	Type    string      `json:"type"` // "qemu" or "lxc"
	Node    string      `json:"node"`
	Status  string      `json:"status"`
	Uptime  int64       `json:"uptime"`
	CPU     float64     `json:"cpu"`
	MaxCPU  int         `json:"maxcpu"`
	Mem     int64       `json:"mem"`
	MaxMem  int64       `json:"maxmem"`
	Disk    int64       `json:"disk"`
	MaxDisk int64       `json:"maxdisk"`
}

// GetAllResources fetches all VMs and Containers from Proxmox
func GetAllResources() ([]models.Resource, error) {
	cmd := exec.Command("pvesh", "get", "/cluster/resources", "--type", "vm", "--output-format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute pvesh command: %w", err)
	}

	var proxmoxResources []ProxmoxResource
	if err := json.Unmarshal(output, &proxmoxResources); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	resources := make([]models.Resource, 0, len(proxmoxResources))
	for _, pr := range proxmoxResources {
		vmidInt, err := pr.VMID.Int64()
		if err != nil {
			continue // Skip invalid VMID
		}

		resource := models.Resource{
			VMID:        int(vmidInt),
			Name:        pr.Name,
			Type:        pr.Type,
			Node:        pr.Node,
			Status:      pr.Status,
			Uptime:      pr.Uptime,
			CPUUsage:    pr.CPU,
			MemoryUsed:  pr.Mem,
			MemoryTotal: pr.MaxMem,
			DiskUsed:    pr.Disk,
			DiskTotal:   pr.MaxDisk,
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetResource fetches a specific VM or Container by VMID and node
func GetResource(node string, vmid int) (*models.Resource, error) {
	// Get all resources and filter
	resources, err := GetAllResources()
	if err != nil {
		return nil, err
	}

	for _, r := range resources {
		if r.VMID == vmid && r.Node == node {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("resource %d not found on node %s", vmid, node)
}

// RestartResource restarts a VM or Container
func RestartResource(node string, vmid int, resourceType string) error {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", node, vmid)
	} else {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart resource: %w, output: %s", err, string(output))
	}

	return nil
}

// StopResource stops a VM or Container
func StopResource(node string, vmid int, resourceType string) error {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", node, vmid)
	} else {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop resource: %w, output: %s", err, string(output))
	}

	return nil
}

// StartResource starts a VM or Container
func StartResource(node string, vmid int, resourceType string) error {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/start", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/start", node, vmid)
	} else {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if resource is already running
		if strings.Contains(string(output), "already running") {
			return nil
		}
		return fmt.Errorf("failed to start resource: %w, output: %s", err, string(output))
	}

	return nil
}
