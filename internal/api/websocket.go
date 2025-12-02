package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// TODO: Restrict this in production
		return true
	},
}

// TerminalHandler handles WebSocket connections for terminal access
func TerminalHandler(w http.ResponseWriter, r *http.Request) {
	// WebSocket connections can't send Basic Auth headers during upgrade
	// Check credentials from query parameters instead
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	// Get expected credentials
	expectedUsername := os.Getenv("AUTH_USERNAME")
	expectedPassword := os.Getenv("AUTH_PASSWORD")
	if expectedUsername == "" {
		expectedUsername = "admin"
	}
	if expectedPassword == "" {
		expectedPassword = "proxmox2024"
	}

	log.Printf("INFO: expectedUsername: %s expectedPassword: %s username: %s password: %s", expectedUsername, expectedPassword, username, password)

	// Validate credentials
	if username != expectedUsername || password != expectedPassword {
		log.Printf("ERROR: WebSocket authentication failed for user: %s", username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get parameters
	vmidStr := r.URL.Query().Get("vmid")
	node := r.URL.Query().Get("node")
	resourceType := r.URL.Query().Get("type") // "lxc" or "qemu"

	if vmidStr == "" || node == "" {
		log.Printf("ERROR: Missing required parameters: vmid=%s, node=%s", vmidStr, node)
		http.Error(w, "vmid and node are required", http.StatusBadRequest)
		return
	}

	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		log.Printf("ERROR: Invalid vmid '%s': %v", vmidStr, err)
		http.Error(w, "Invalid vmid", http.StatusBadRequest)
		return
	}

	log.Printf("Terminal connection request: vmid=%d, node=%s, type=%s", vmid, node, resourceType)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket upgraded successfully for VMID %d", vmid)

	// Determine command based on resource type
	var cmd *exec.Cmd
	if resourceType == "lxc" {
		// For LXC containers, use pct enter
		log.Printf("Starting LXC terminal for VMID %d using 'pct enter'", vmid)
		cmd = exec.Command("pct", "enter", strconv.Itoa(vmid))
	} else {
		// For QEMU VMs, use qm terminal (requires serial console)
		log.Printf("Starting QEMU terminal for VMID %d using 'qm terminal'", vmid)
		cmd = exec.Command("qm", "terminal", strconv.Itoa(vmid))
	}

	// Start PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start terminal: %v\r\n\r\nNote: The backend must run ON the Proxmox server with root privileges.\r\n", err)
		log.Printf("ERROR: Failed to start PTY for VMID %d: %v", vmid, err)
		conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return
	}
	defer func() {
		ptmx.Close()
		cmd.Process.Kill()
	}()

	log.Printf("Terminal session started for VMID %d (type: %s)", vmid, resourceType)

	// Handle PTY -> WebSocket (output)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("ERROR: PTY read error: %v", err)
				}
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("ERROR: WebSocket write error: %v", err)
				return
			}
		}
	}()

	// Handle WebSocket -> PTY (input)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ERROR: WebSocket read error: %v", err)
			}
			break
		}
		if _, err := ptmx.Write(message); err != nil {
			log.Printf("ERROR: PTY write error: %v", err)
			break
		}
	}

	log.Printf("Terminal session ended for VMID %d", vmid)
}
