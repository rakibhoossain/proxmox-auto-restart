BINARY_NAME := proxmox-auto-restart

.PHONY: all build build-local build-linux build-linux-amd64 build-linux-arm64 clean test run deploy install

all: build

# Build for all platforms
build: build-local build-linux-amd64

# Build for local platform
build-local:
	go build -o $(BINARY_NAME) ./cmd/server

# Build for Linux amd64 (most common for Proxmox servers)
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 ./cmd/server

# Build for Linux arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 ./cmd/server

# Build for Linux (default to amd64)
build-linux: build-linux-amd64

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
