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
func RestartResource(node string, vmid int, resourceType string) (string, error) {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", node, vmid)
	} else {
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		return outputStr, fmt.Errorf("failed to restart resource: %w, output: %s", err, outputStr)
	}

	return outputStr, nil
}

// StopResource stops a VM or Container
func StopResource(node string, vmid int, resourceType string) (string, error) {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", node, vmid)
	} else {
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		return outputStr, fmt.Errorf("failed to stop resource: %w, output: %s", err, outputStr)
	}

	return outputStr, nil
}

// StartResource starts a VM or Container
func StartResource(node string, vmid int, resourceType string) (string, error) {
	var cmdPath string

	if resourceType == "lxc" {
		cmdPath = fmt.Sprintf("/nodes/%s/lxc/%d/status/start", node, vmid)
	} else if resourceType == "qemu" {
		cmdPath = fmt.Sprintf("/nodes/%s/qemu/%d/status/start", node, vmid)
	} else {
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}

	cmd := exec.Command("pvesh", "create", cmdPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		// Check if resource is already running
		if strings.Contains(outputStr, "already running") {
			return outputStr, nil
		}
		return outputStr, fmt.Errorf("failed to start resource: %w, output: %s", err, outputStr)
	}

	return outputStr, nil
}

// CloneContainer clones a container to a new VMID
// Usage: pct clone <source> <new> --target <node>
func CloneContainer(sourceVMID, newVMID int, targetNode, hostname string) error {
	args := []string{"clone", fmt.Sprintf("%d", sourceVMID), fmt.Sprintf("%d", newVMID), "--target", targetNode}

	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}

	cmd := exec.Command("pct", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone container: %w, output: %s", err, string(output))
	}

	return nil
}

// DeleteContainer deletes a container
// Usage: pct destroy <vmid> --purge
func DeleteContainer(vmid int, node string) error {
	// Stop container first if running
	resource, err := GetResource(node, vmid)
	if err == nil && resource.Status == "running" {
		_, _ = StopResource(node, vmid, resource.Type)
	}

	cmd := exec.Command("pct", "destroy", fmt.Sprintf("%d", vmid), "--purge")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete container: %w, output: %s", err, string(output))
	}

	return nil
}

// ExecuteInContainer executes a command inside a container
// Usage: pct exec <vmid> -- <command>
func ExecuteInContainer(vmid int, command string) error {
	cmd := exec.Command("pct", "exec", fmt.Sprintf("%d", vmid), "--", "bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command in container: %w, output: %s", err, string(output))
	}

	return nil
}

// DeployBlockchainNode orchestrates the full deployment: clone → start → exec commands
func DeployBlockchainNode(sourceVMID, newVMID int, targetNode, hostname string, commands []string) error {
	// Step 1: Clone the container
	if err := CloneContainer(sourceVMID, newVMID, targetNode, hostname); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Step 2: Start the container
	if _, err := StartResource(targetNode, newVMID, "lxc"); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	// Step 3: Execute base setup commands (always run these first)
	baseSetupCommands := []string{
		"apt-get update && apt-get install -y locales",
		"locale-gen en_US.UTF-8",
		"update-locale LANG=en_US.UTF-8",
		"apt-get update && apt-get install -y curl wget",
	}

	for i, cmd := range baseSetupCommands {
		if err := ExecuteInContainer(newVMID, cmd); err != nil {
			return fmt.Errorf("base setup command %d failed: %w", i+1, err)
		}
	}

	// Step 4: Execute user-provided commands
	for i, cmd := range commands {
		if err := ExecuteInContainer(newVMID, cmd); err != nil {
			return fmt.Errorf("command %d failed: %w", i+1, err)
		}
	}

	return nil
}
