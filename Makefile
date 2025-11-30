.PHONY: build run test clean deploy

# Build binary for Linux (Proxmox server)
build:
	GOOS=linux GOARCH=amd64 go build -o proxmox-auto-restart ./cmd/server

# Build for current OS (development)
build-local:
	go build -o proxmox-auto-restart ./cmd/server

# Run locally
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f proxmox-auto-restart
	rm -f *.db *.db-*

# Deploy to Proxmox server (set PROXMOX_HOST environment variable)
deploy: build
	@if [ -z "$(PROXMOX_HOST)" ]; then \
		echo "Error: PROXMOX_HOST not set. Usage: make deploy PROXMOX_HOST=root@your-proxmox-server"; \
		exit 1; \
	fi
	ssh $(PROXMOX_HOST) 'mkdir -p /opt/proxmox-auto-restart'
	scp proxmox-auto-restart $(PROXMOX_HOST):/opt/proxmox-auto-restart/
	scp deployments/proxmox-auto-restart.service $(PROXMOX_HOST):/etc/systemd/system/
	ssh $(PROXMOX_HOST) 'systemctl daemon-reload && systemctl restart proxmox-auto-restart && systemctl status proxmox-auto-restart'

# Install dependencies
deps:
	go mod download
	go mod tidy
