.PHONY: all build clean test migrate migrate-up migrate-down migrate-status migrate-force create-migration install-migrate seed-wallets topup-wallet dev check-routes check-coverage login login-token login-env login-quick login-quick-token login-quick-env mock-jnt mock-services kiwi-setup kiwi-export-tests kiwi-run-tests kiwi-update-mapping kiwi-export-auth-tests kiwi-run-auth-tests kiwi-install-deps tunnel-start tunnel-stop tunnel-status droplet-on droplet-off tunnel-migrate-up tunnel-migrate-down tunnel-migrate-status tunnel-migrate-force db-health-check db-health-quick db-fix-blocking db-analyze-blocking db-configure-prevention sicepat-tunnel-start sicepat-tunnel-stop sicepat-tunnel-status sicepat-test-forwarded sicepat-test-forwarded-prod sicepat-test-remote sicepat-test-remote-prod sicepat-test-connectivity sicepat-help dev-receipt-build dev-receipt-stats dev-receipt-reset dev-receipt-check dev-receipt-test dev-receipt-clean docker-build docker-run docker-push docker-login docker-tag docker-clean docker-dev docker-build-preprod docker-push-preprod docker-run-preprod preprod-build-and-push ghcr-login ghcr-push ghcr-tag-latest ghcr-push-latest ghcr-push-all reset-password reset-password-interactive reset-password-help test-retry-deduct test-retry-deduct-functionality

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=smartseller-backend
MAIN_PATH=./cmd/main.go
MOCK_PORT=8081

# Docker parameters
DOCKER_REGISTRY=ghcr.io
DOCKER_NAMESPACE=kirimku
DOCKER_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(BINARY_NAME)
DOCKER_TAG ?= latest
DOCKER_PLATFORM ?= linux/amd64,linux/arm64
DOCKERFILE_PATH=./Dockerfile

# Preproduction Docker parameters
PREPROD_DOCKERFILE_PATH=./Dockerfile.preproduction
PREPROD_TAG=preproduction
PREPROD_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(BINARY_NAME)

# Git parameters for versioning
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell git describe --tags --always --dirty)

# Migration parameters
MIGRATION_DIR=internal/infrastructure/database/migrations
MIGRATION_DSN=$(shell grep DATABASE_URL .env | cut -d '=' -f2-)

# Kiwi TCMS parameters
KIWI_PORT=8089
KIWI_URL=http://localhost:$(KIWI_PORT)/xml-rpc
KIWI_USERNAME=admin
KIWI_PASSWORD=admin

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	./$(BINARY_NAME)

# Run the application in development mode (with hot reload)
dev:
	$(GOGET) -u github.com/air-verse/air
	air

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Download and install dependencies
deps:
	$(GOGET) -v ./...
	$(GOCMD) mod tidy
	$(GOCMD) mod verify

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Check test coverage
check-coverage:
	bash scripts/check_coverage.sh

# Platform login for testing
# Usage: make login EMAIL=user@example.com PASSWORD=password123
# or:    make login-token EMAIL=user@example.com PASSWORD=password123
# or:    make login-env EMAIL=user@example.com PASSWORD=password123
login:
	@if [ -z "$(EMAIL)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Error: EMAIL and PASSWORD are required"; \
		echo "Usage: make login EMAIL=user@example.com PASSWORD=password123"; \
		echo "Or set environment variables: LOGIN_EMAIL and LOGIN_PASSWORD"; \
		exit 1; \
	fi
	@echo "Logging in to platform..."
	@$(GORUN) scripts/login.go -email="$(EMAIL)" -password="$(PASSWORD)" -url=http://localhost:8090 -output=json

login-token:
	@if [ -z "$(EMAIL)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Error: EMAIL and PASSWORD are required"; \
		echo "Usage: make login-token EMAIL=user@example.com PASSWORD=password123"; \
		echo "Or set environment variables: LOGIN_EMAIL and LOGIN_PASSWORD"; \
		exit 1; \
	fi
	@$(GORUN) scripts/login.go -email="$(EMAIL)" -password="$(PASSWORD)" -url=http://localhost:8090 -output=token

login-env:
	@if [ -z "$(EMAIL)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Error: EMAIL and PASSWORD are required"; \
		echo "Usage: make login-env EMAIL=user@example.com PASSWORD=password123"; \
		echo "Example: eval $$(make login-env EMAIL=user@example.com PASSWORD=password123)"; \
		echo "Or set environment variables: LOGIN_EMAIL and LOGIN_PASSWORD"; \
		exit 1; \
	fi
	@$(GORUN) scripts/login.go -email="$(EMAIL)" -password="$(PASSWORD)" -url=http://localhost:8090 -output=env

# Login with environment variables (no parameters needed if LOGIN_EMAIL and LOGIN_PASSWORD are set)
login-quick:
	@echo "Logging in with environment variables..."
	@$(GORUN) scripts/login.go -url=http://localhost:8090 -output=json

login-quick-token:
	@$(GORUN) scripts/login.go -url=http://localhost:8090 -output=token

login-quick-env:
	@$(GORUN) scripts/login.go -url=http://localhost:8090 -output=env

# Database migrations
migrate-up:
	@echo "Running migrations..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
		DATABASE_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE"; \
		migrate -path $(MIGRATION_DIR) -database "$$DATABASE_URL" up; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

migrate-down:
	@echo "Rolling back migrations..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
		DATABASE_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE"; \
		migrate -path $(MIGRATION_DIR) -database "$$DATABASE_URL" down; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

migrate-status:
	@echo "Checking migration status..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
		DATABASE_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE"; \
		migrate -path $(MIGRATION_DIR) -database "$$DATABASE_URL" version; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

migrate-force:
	@echo "Forcing migration version..."
	@if [ -f .env ] && [ -n "$(VERSION)" ]; then \
		export $$(grep -v '^#' .env | xargs); \
		DATABASE_URL="postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE"; \
		migrate -path $(MIGRATION_DIR) -database "$$DATABASE_URL" force $(VERSION); \
	else \
		echo "Error: .env file not found or VERSION not specified"; \
		echo "Usage: make migrate-force VERSION=<version_number>"; \
		exit 1; \
	fi

# Create a new migration file
create-migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $${name}

# Install migration tool
install-migrate:
	@echo "Installing migrate tool..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Seed wallets for users without wallets
seed-wallets:
	@echo "Seeding wallets for users without wallets..."
	$(GORUN) ./cmd/seed_wallets/main.go

# Topup wallet for development/testing
# Usage: make topup-wallet EMAIL=user@example.com AMOUNT=50000 NOTES="Test topup"
# or:    make topup-wallet USER_ID=user-uuid AMOUNT=50000 NOTES="Test topup"
topup-wallet:
	@if [ -z "$(AMOUNT)" ]; then \
		echo "Error: AMOUNT is required. Usage: make topup-wallet EMAIL=user@example.com AMOUNT=50000"; \
		exit 1; \
	fi
	@if [ -z "$(EMAIL)" ] && [ -z "$(USER_ID)" ]; then \
		echo "Error: Either EMAIL or USER_ID is required."; \
		echo "Usage: make topup-wallet EMAIL=user@example.com AMOUNT=50000"; \
		echo "   or: make topup-wallet USER_ID=user-uuid AMOUNT=50000"; \
		exit 1; \
	fi
	@echo "Topping up wallet..."
	@if [ -n "$(EMAIL)" ]; then \
		$(GORUN) ./cmd/topup_wallet/main.go -email="$(EMAIL)" -amount="$(AMOUNT)" -notes="$(NOTES)"; \
	else \
		$(GORUN) ./cmd/topup_wallet/main.go -user-id="$(USER_ID)" -amount="$(AMOUNT)" -notes="$(NOTES)"; \
	fi

# Check route consistency
check-routes:
	$(GORUN) pkg/utils/check_routes_consistency.go

# Start the JNT mock server
mock-jnt:
	@echo "Starting JNT mock server on port $(MOCK_PORT)..."
	$(GORUN) cmd/mockserver/main.go -port $(MOCK_PORT)

# Start all mock services
mock-services: mock-jnt

# Run app with mock services enabled
dev-with-mocks:
	@echo "Starting app with mock services..."
	@USE_JNT_MOCK=true air

# Kiwi TCMS integration
kiwi-install-deps:
	@echo "Installing Kiwi TCMS dependencies..."
	@pip3 install tcms-api

kiwi-setup: kiwi-install-deps
	@echo "Setting up Kiwi TCMS integration..."
	@bash scripts/setup_kiwi_tcms.sh

kiwi-export-tests:
	@echo "Exporting test cases to Kiwi TCMS..."
	@KIWI_URL=$(KIWI_URL) KIWI_USERNAME=$(KIWI_USERNAME) KIWI_PASSWORD=$(KIWI_PASSWORD) \
	$(GORUN) internal/tests/tcms/examples/export_jnt.go

kiwi-update-mapping:
	@echo "Updating test case mappings from Kiwi TCMS..."
	@KIWI_URL=$(KIWI_URL) KIWI_USERNAME=$(KIWI_USERNAME) KIWI_PASSWORD=$(KIWI_PASSWORD) \
	bash scripts/update_tcms_mapping.sh

kiwi-run-tests: kiwi-update-mapping
	@echo "Running tests and reporting results to Kiwi TCMS..."
	@KIWI_URL=$(KIWI_URL) KIWI_USERNAME=$(KIWI_USERNAME) KIWI_PASSWORD=$(KIWI_PASSWORD) \
	$(GORUN) internal/tests/tcms/examples/run_tests.go

# Auth endpoint tests with Kiwi TCMS (Python implementation)
kiwi-export-auth-tests:
	@echo "Exporting auth test cases to Kiwi TCMS..."
	@chmod +x internal/tests/tcms/python/export_auth_tests.py
	@KIWI_URL=$(KIWI_URL) KIWI_USERNAME=$(KIWI_USERNAME) KIWI_PASSWORD=$(KIWI_PASSWORD) \
	internal/tests/tcms/python/export_auth_tests.py

kiwi-run-auth-tests: 
	@echo "Running auth tests and reporting results to Kiwi TCMS..."
	@chmod +x internal/tests/tcms/python/run_auth_tests.py
	@KIWI_URL=$(KIWI_URL) KIWI_USERNAME=$(KIWI_USERNAME) KIWI_PASSWORD=$(KIWI_PASSWORD) \
	internal/tests/tcms/python/run_auth_tests.py --api-url=http://localhost:8088

# SSH tunnel commands
tunnel-start:
	@echo "Starting SSH tunnel to database..."
	@chmod +x scripts/setup_ssh_tunnel.sh
	@scripts/setup_ssh_tunnel.sh start

tunnel-stop:
	@echo "Stopping SSH tunnel to database..."
	@chmod +x scripts/setup_ssh_tunnel.sh
	@scripts/setup_ssh_tunnel.sh stop

tunnel-status:
	@echo "Checking SSH tunnel status..."
	@chmod +x scripts/setup_ssh_tunnel.sh
	@scripts/setup_ssh_tunnel.sh status

droplet-on:
	@echo "Powering on the Droplet..."
	@chmod +x scripts/setup_ssh_tunnel.sh
	@scripts/setup_ssh_tunnel.sh start

droplet-off:
	@echo "Powering off the Droplet..."
	@chmod +x scripts/setup_ssh_tunnel.sh
	@scripts/setup_ssh_tunnel.sh poweroff

# Database migrations via SSH tunnel
tunnel-migrate-up:
	@echo "Running migrations via SSH tunnel..."
	@chmod +x scripts/droplet_migrate_db.sh
	@scripts/droplet_migrate_db.sh up

tunnel-migrate-down:
	@echo "Rolling back migrations via SSH tunnel..."
	@chmod +x scripts/droplet_migrate_db.sh
	@scripts/droplet_migrate_db.sh down

tunnel-migrate-status:
	@echo "Checking migration status via SSH tunnel..."
	@chmod +x scripts/droplet_migrate_db.sh
	@scripts/droplet_migrate_db.sh status

tunnel-migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION not specified"; \
		echo "Usage: make tunnel-migrate-force VERSION=<version_number>"; \
		exit 1; \
	fi
	@echo "Forcing migration version via SSH tunnel..."
	@chmod +x scripts/droplet_migrate_db.sh
	@scripts/droplet_migrate_db.sh force $(VERSION)

# Database health check commands
db-health-check:
	@echo "Running full database health check..."
	@chmod +x scripts/db_health_check.sh
	@scripts/db_health_check.sh

db-health-quick:
	@echo "Running quick database health check..."
	@chmod +x scripts/db_health_check.sh
	@scripts/db_health_check.sh --quick

# SiCepat API testing and tunnel management
# Prerequisites: Ensure SICEPAT_API_KEY is set in .env file

# Build development receipt manager
dev-receipt-build:
	@echo "Building development receipt manager..."
	go build -o dev-receipt-manager cmd/dev_receipt_manager/main.go

# Show development receipt statistics
dev-receipt-stats: dev-receipt-build
	@echo "Development SiCepat Receipt Statistics"
	@echo "======================================"
	./dev-receipt-manager -action=stats

# Reset development receipt tracking (clears all used receipts)
dev-receipt-reset: dev-receipt-build
	@echo "Resetting development receipt tracking..."
	./dev-receipt-manager -action=reset

# Check if a specific receipt number is used
# Usage: make dev-receipt-check RECEIPT=888889340577
dev-receipt-check: dev-receipt-build
	@if [ -z "$(RECEIPT)" ]; then \
		echo "❌ Please provide a receipt number: make dev-receipt-check RECEIPT=888889340577"; \
		exit 1; \
	fi
	@echo "Checking receipt number: $(RECEIPT)"
	./dev-receipt-manager -action=check -receipt=$(RECEIPT)

# Test development receipt generation
dev-receipt-test:
	@echo "Testing development receipt generation..."
	@echo 'package main' > test_dev_receipts_temp.go
	@echo 'import (' >> test_dev_receipts_temp.go
	@echo '	"context"' >> test_dev_receipts_temp.go
	@echo '	"fmt"' >> test_dev_receipts_temp.go
	@echo '	"os"' >> test_dev_receipts_temp.go
	@echo '	"github.com/kirimku/kirimku-backend/pkg/utils"' >> test_dev_receipts_temp.go
	@echo ')' >> test_dev_receipts_temp.go
	@echo 'func main() {' >> test_dev_receipts_temp.go
	@echo '	os.Setenv("APP_ENV", "development")' >> test_dev_receipts_temp.go
	@echo '	fmt.Println("🧪 Testing Development SiCepat Receipt Generation")' >> test_dev_receipts_temp.go
	@echo '	for i := 1; i <= 3; i++ {' >> test_dev_receipts_temp.go
	@echo '		receipt, err := utils.GenerateSiCepatReceiptNumber(context.Background(), nil, i)' >> test_dev_receipts_temp.go
	@echo '		if err != nil { fmt.Printf("❌ Error: %v\\n", err); continue }' >> test_dev_receipts_temp.go
	@echo '		fmt.Printf("✅ Transaction %d: Receipt %s\\n", i, receipt)' >> test_dev_receipts_temp.go
	@echo '	}' >> test_dev_receipts_temp.go
	@echo '	stats := utils.GetDevelopmentReceiptStats()' >> test_dev_receipts_temp.go
	@echo '	fmt.Printf("📊 Used: %v/%v (%.1f%%)\\n", stats["used_count"], stats["total_range"], stats["usage_percentage"])' >> test_dev_receipts_temp.go
	@echo '}' >> test_dev_receipts_temp.go
	@APP_ENV=development go run test_dev_receipts_temp.go
	@rm test_dev_receipts_temp.go

# Clean up development receipt files and binaries
dev-receipt-clean:
	@echo "Cleaning up development receipt files..."
	rm -f dev-receipt-manager
	rm -f dev_sicepat_receipts.json
	@echo "✅ Cleaned up development receipt manager and tracking file"
sicepat-tunnel-start:
	@echo "Starting SiCepat API tunnels..."
	@echo "This creates SSH tunnels through the whitelisted droplet IP"
	@echo "Sandbox:    localhost:8443 -> pickup.sandbox.sicepat.com:443"
	@echo "Production: localhost:8444 -> pickup.sicepat.com:443"
	@echo ""
	@ssh -f -N -L "8443:pickup.sandbox.sicepat.com:443" -i "$$HOME/.ssh/id_rsa" "root@159.89.192.219" || echo "Sandbox tunnel may already be running"
	@ssh -f -N -L "8444:pickup.sicepat.com:443" -i "$$HOME/.ssh/id_rsa" "root@159.89.192.219" || echo "Production tunnel may already be running"
	@sleep 2
	@echo "✅ SiCepat tunnels started successfully!"
	@echo "Use 'make sicepat-tunnel-status' to verify"

sicepat-tunnel-stop:
	@echo "Stopping SiCepat API tunnels..."
	@pkill -f "8443:pickup.sandbox.sicepat.com:443" || echo "Sandbox tunnel not running"
	@pkill -f "8444:pickup.sicepat.com:443" || echo "Production tunnel not running"
	# Ensure any process listening on the local forward ports is terminated
	@for port in 8443 8444; do \
		PIDS=$$(lsof -t -iTCP:$$port -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$PIDS" ]; then \
			echo "Killing process(es) listening on port $$port: $$PIDS"; \
			kill $$PIDS 2>/dev/null || true; \
			sleep 1; \
			PIDS_AFTER=$$(lsof -t -iTCP:$$port -sTCP:LISTEN 2>/dev/null || true); \
			if [ -n "$$PIDS_AFTER" ]; then \
				echo "Force killing remaining PIDs on port $$port: $$PIDS_AFTER"; \
				kill -9 $$PIDS_AFTER 2>/dev/null || true; \
			fi; \
			# Final check
			if lsof -nP -iTCP:$$port -sTCP:LISTEN >/dev/null 2>&1; then \
				echo "Warning: port $$port still in use after kill attempts"; \
			else \
				echo "Port $$port freed"; \
			fi; \
		else \
			echo "No process listening on port $$port"; \
		fi; \
	done
	@echo ""
	@echo "Post-stop port status:"
	@for port in 8443 8444; do \
		if lsof -nP -iTCP:$$port -sTCP:LISTEN >/dev/null 2>&1; then \
			echo "Port $$port: IN USE"; \
			lsof -nP -iTCP:$$port -sTCP:LISTEN || true; \
		else \
			echo "Port $$port: FREE"; \
		fi; \
	done
	@echo ""
	@echo "Tunnel process status (by pattern):"
	@if pgrep -f "8443:pickup.sandbox.sicepat.com:443" > /dev/null; then \
		echo "Sandbox tunnel process: PRESENT"; \
	else \
		echo "Sandbox tunnel process: NOT PRESENT"; \
	fi
	@if pgrep -f "8444:pickup.sicepat.com:443" > /dev/null; then \
		echo "Production tunnel process: PRESENT"; \
	else \
		echo "Production tunnel process: NOT PRESENT"; \
	fi
	@echo "✅ SiCepat tunnels stopped"

sicepat-tunnel-status:
	@echo "Checking SiCepat tunnel status..."
	@echo ""
	@if lsof -nP -iTCP:8443 -sTCP:LISTEN >/dev/null 2>&1; then \
		echo "✅ Sandbox tunnel (localhost:8443) - Running"; \
	else \
		echo "❌ Sandbox tunnel (localhost:8443) - Not running"; \
	fi
	@if lsof -nP -iTCP:8444 -sTCP:LISTEN >/dev/null 2>&1; then \
		echo "✅ Production tunnel (localhost:8444) - Running"; \
	else \
		echo "❌ Production tunnel (localhost:8444) - Not running"; \
	fi
	@echo ""
	@echo "Test connectivity:"
	@curl -k -s --max-time 3 --connect-timeout 3 "https://localhost:8443" >/dev/null 2>&1 && \
		echo "✅ Sandbox tunnel connectivity - OK" || \
		echo "❌ Sandbox tunnel connectivity - Failed"
	@curl -k -s --max-time 3 --connect-timeout 3 "https://localhost:8444" >/dev/null 2>&1 && \
		echo "✅ Production tunnel connectivity - OK" || \
		echo "❌ Production tunnel connectivity - Failed"

sicepat-test-forwarded:
	@echo "Testing SiCepat API via local tunnel (sandbox)..."
	@echo "This uses localhost:8443 tunnel to test the pickup API"
	@chmod +x scripts/sicepat/booking/test_pickup_forwarded.sh
	@scripts/sicepat/booking/test_pickup_forwarded.sh

sicepat-test-forwarded-prod:
	@echo "Testing SiCepat API via local tunnel (production)..."
	@echo "This uses localhost:8444 tunnel to test the pickup API"
	@chmod +x scripts/sicepat/booking/test_pickup_forwarded.sh
	@SICEPAT_PROD=true scripts/sicepat/booking/test_pickup_forwarded.sh

sicepat-test-remote:
	@echo "Testing SiCepat API directly on droplet..."
	@echo "This runs the test on the droplet to ensure whitelisted IP is used"
	@chmod +x scripts/sicepat/booking/test_pickup_remote.sh
	@scripts/sicepat/booking/test_pickup_remote.sh

sicepat-test-remote-prod:
	@echo "Testing SiCepat API directly on droplet (production)..."
	@chmod +x scripts/sicepat/booking/test_pickup_remote.sh
	@SICEPAT_PROD=true scripts/sicepat/booking/test_pickup_remote.sh

sicepat-test-connectivity:
	@echo "Testing basic SiCepat connectivity via droplet..."
	@chmod +x scripts/test_sicepat_via_droplet.sh
	@scripts/test_sicepat_via_droplet.sh

sicepat-help:
	@echo "SiCepat API Testing Commands:"
	@echo "============================"
	@echo ""
	@echo "🔧 Tunnel Management:"
	@echo "  make sicepat-tunnel-start    - Start SSH tunnels for SiCepat API"
	@echo "  make sicepat-tunnel-stop     - Stop SSH tunnels"
	@echo "  make sicepat-tunnel-status   - Check tunnel status and connectivity"
	@echo ""
	@echo "🧪 API Testing:"
	@echo "  make sicepat-test-forwarded      - Test pickup API via tunnel (sandbox)"
	@echo "  make sicepat-test-forwarded-prod - Test pickup API via tunnel (production)"
	@echo "  make sicepat-test-remote         - Test pickup API on droplet (sandbox)"
	@echo "  make sicepat-test-remote-prod    - Test pickup API on droplet (production)"
	@echo "  make sicepat-test-connectivity   - Basic connectivity test"
	@echo ""
	@echo "📋 Development Receipt Management:"
	@echo "  make dev-receipt-stats           - Show development receipt usage statistics"
	@echo "  make dev-receipt-test            - Test development receipt generation"
	@echo "  make dev-receipt-check RECEIPT=... - Check if receipt number is used"
	@echo "  make dev-receipt-reset           - Reset all used receipt numbers"
	@echo "  make dev-receipt-clean           - Clean up receipt files and binaries"
	@echo ""
	@echo "📋 Prerequisites:"
	@echo "  - SICEPAT_API_KEY must be set in .env file"
	@echo "  - SSH key (id_rsa) must have access to droplet 159.89.192.219"
	@echo "  - For tunnel tests: Run 'make sicepat-tunnel-start' first"
	@echo ""
	@echo "💡 Environment Variables:"
	@echo "  RECEIPT_NUMBER=123456789012 - Custom receipt number"
	@echo "  REFERENCE_NUMBER=MY-REF-123 - Custom reference number"
	@echo "  APP_ENV=development          - Use development receipt range"
	@echo ""
	@echo "🔍 Quick Start:"
	@echo "  1. make sicepat-tunnel-start"
	@echo "  2. make sicepat-test-forwarded"
	@echo "  3. make sicepat-tunnel-stop (when done)"
	@echo ""
	@echo "📊 Development Receipt Range: 888889340571-888889341570 (1000 numbers)"

db-analyze-blocking:
	@echo "Analyzing database blocking queries..."
	@chmod +x scripts/fix_blocking_queries.sh
	@scripts/fix_blocking_queries.sh --analyze

db-fix-blocking:
	@echo "Fixing database blocking queries (DRY RUN - use 'make db-fix-blocking-live' for actual execution)..."
	@chmod +x scripts/fix_blocking_queries.sh
	@scripts/fix_blocking_queries.sh --auto-fix --dry-run

db-fix-blocking-live:
	@echo "⚠️  WARNING: This will actually terminate blocking processes!"
	@echo "⚠️  Make sure you want to proceed. Press Ctrl+C to cancel."
	@echo "⚠️  Proceeding in 5 seconds..."
	@sleep 5
	@chmod +x scripts/fix_blocking_queries.sh
	@scripts/fix_blocking_queries.sh --auto-fix --no-dry-run

db-configure-prevention:
	@echo "Configuring database for blocking prevention (DRY RUN - use 'make db-configure-prevention-live' for actual execution)..."
	@chmod +x scripts/configure_db_prevention.sh
	@scripts/configure_db_prevention.sh --dry-run

db-configure-prevention-live:
	@echo "⚠️  WARNING: This will modify database configuration!"
	@echo "⚠️  Make sure you want to proceed. Press Ctrl+C to cancel."
	@echo "⚠️  Proceeding in 5 seconds..."
	@sleep 5
	@chmod +x scripts/configure_db_prevention.sh
	@scripts/configure_db_prevention.sh --no-dry-run

# Docker commands
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build date: $(BUILD_DATE)"
	docker build \
		--build-arg VERSION="$(VERSION)" \
		--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg GIT_BRANCH="$(GIT_BRANCH)" \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT) \
		-t $(DOCKER_IMAGE):latest \
		-f $(DOCKERFILE_PATH) .

docker-build-multiarch:
	@echo "Building multi-architecture Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker buildx build \
		--platform $(DOCKER_PLATFORM) \
		--build-arg VERSION="$(VERSION)" \
		--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg GIT_BRANCH="$(GIT_BRANCH)" \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT) \
		-t $(DOCKER_IMAGE):latest \
		-f $(DOCKERFILE_PATH) \
		--push .

docker-run:
	@echo "Running Docker container from $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker run --rm -it \
		-p 8080:8080 \
		-e APP_ENV=development \
		-e LOG_LEVEL=debug \
		--name $(BINARY_NAME)-container \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run-detached:
	@echo "Running Docker container in detached mode from $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker run -d \
		-p 8080:8080 \
		-e APP_ENV=development \
		-e LOG_LEVEL=debug \
		--name $(BINARY_NAME)-container \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME)-container || true
	docker rm $(BINARY_NAME)-container || true

docker-logs:
	@echo "Showing Docker container logs..."
	docker logs -f $(BINARY_NAME)-container

docker-shell:
	@echo "Opening shell in running container..."
	docker exec -it $(BINARY_NAME)-container /bin/sh

docker-tag:
	@echo "Tagging image $(DOCKER_IMAGE):$(DOCKER_TAG) with additional tags..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

docker-push:
	@echo "Pushing Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT)
	docker push $(DOCKER_IMAGE):latest

docker-clean:
	@echo "Cleaning up Docker images and containers..."
	docker stop $(BINARY_NAME)-container 2>/dev/null || true
	docker rm $(BINARY_NAME)-container 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):latest 2>/dev/null || true
	docker system prune -f

docker-dev: docker-build docker-run

# Preproduction Docker commands (specifically for Ubuntu VM deployment)
docker-build-preprod:
	@echo "Building Preproduction Docker image for Ubuntu VM: $(PREPROD_IMAGE):$(PREPROD_TAG)"
	@echo "Target platform: linux/amd64 (Ubuntu VM)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build date: $(BUILD_DATE)"
	docker build \
		--platform linux/amd64 \
		--build-arg VERSION="$(VERSION)" \
		--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg GIT_BRANCH="$(GIT_BRANCH)" \
		-t $(PREPROD_IMAGE):$(PREPROD_TAG) \
		-t $(PREPROD_IMAGE):$(PREPROD_TAG)-$(GIT_SHORT_COMMIT) \
		-f $(PREPROD_DOCKERFILE_PATH) .

docker-push-preprod:
	@echo "Pushing Preproduction Docker image to registry..."
	docker push $(PREPROD_IMAGE):$(PREPROD_TAG)
	docker push $(PREPROD_IMAGE):$(PREPROD_TAG)-$(GIT_SHORT_COMMIT)

docker-run-preprod:
	@echo "Running Preproduction Docker container from $(PREPROD_IMAGE):$(PREPROD_TAG)"
	docker run --rm -it \
		-p 8080:8080 \
		-e APP_ENV=preproduction \
		-e LOG_LEVEL=debug \
		-e LOG_FORMAT=pretty \
		--name $(BINARY_NAME)-preprod-container \
		$(PREPROD_IMAGE):$(PREPROD_TAG)

# Complete preproduction pipeline: build and push
preprod-build-and-push: ghcr-login docker-build-preprod docker-push-preprod
	@echo "Preproduction build and push completed successfully!"
	@echo "Image: $(PREPROD_IMAGE):$(PREPROD_TAG)"
	@echo "Commit: $(PREPROD_IMAGE):$(PREPROD_TAG)-$(GIT_SHORT_COMMIT)"

# Clean preproduction images
docker-clean-preprod:
	@echo "Cleaning up Preproduction Docker images and containers..."
	docker stop $(BINARY_NAME)-preprod-container 2>/dev/null || true
	docker rm $(BINARY_NAME)-preprod-container 2>/dev/null || true
	docker rmi $(PREPROD_IMAGE):$(PREPROD_TAG) 2>/dev/null || true
	docker rmi $(PREPROD_IMAGE):$(PREPROD_TAG)-$(GIT_SHORT_COMMIT) 2>/dev/null || true

# GitHub Container Registry (GHCR) commands
ghcr-login:
	@echo "Logging into GitHub Container Registry..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "Error: GITHUB_TOKEN environment variable is required"; \
		echo "Create a Personal Access Token with 'write:packages' permission"; \
		echo "Export it as: export GITHUB_TOKEN=your_token_here"; \
		exit 1; \
	fi
	@GITHUB_USERNAME=$${GITHUB_USER:-$${GITHUB_ACTOR:-$$(git remote get-url origin | sed -n 's#.*github.com[:/]\([^/]*\)/.*#\1#p')}}; \
	if [ -z "$$GITHUB_USERNAME" ]; then \
		echo "Error: Could not determine GitHub username"; \
		echo "Please set GITHUB_USER environment variable"; \
		echo "Example: export GITHUB_USER=your-github-username"; \
		exit 1; \
	fi; \
	echo "Logging in as GitHub user: $$GITHUB_USERNAME"; \
	echo "$$GITHUB_TOKEN" | docker login $(DOCKER_REGISTRY) -u "$$GITHUB_USERNAME" --password-stdin

ghcr-build-and-push: ghcr-login docker-build docker-push
	@echo "Built and pushed $(DOCKER_IMAGE):$(DOCKER_TAG) to GHCR"

ghcr-build-and-push-multiarch: ghcr-login
	@echo "Building and pushing multi-architecture image to GHCR..."
	docker buildx build \
		--platform $(DOCKER_PLATFORM) \
		--build-arg VERSION="$(VERSION)" \
		--build-arg GIT_COMMIT="$(GIT_COMMIT)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg GIT_BRANCH="$(GIT_BRANCH)" \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):$(GIT_SHORT_COMMIT) \
		-t $(DOCKER_IMAGE):latest \
		-f $(DOCKERFILE_PATH) \
		--push .

ghcr-tag-latest:
	@echo "Tagging and pushing latest image to GHCR..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):latest

ghcr-push: docker-push
	@echo "Pushed to GitHub Container Registry"

ghcr-push-all: ghcr-login docker-build docker-tag docker-push
	@echo "Built, tagged, and pushed all versions to GHCR"

# Reset tracking for testing (requires TRACKING_NUMBER and optionally DRY_RUN=true)
reset-tracking:
ifndef TRACKING_NUMBER
	@echo "Error: TRACKING_NUMBER is required"
	@echo "Usage: make reset-tracking TRACKING_NUMBER=JP123456789ID [DRY_RUN=true]"
	@exit 1
endif
	@echo "🔄 Resetting tracking for testing..."
	@if [ "$(DRY_RUN)" = "true" ]; then \
		echo "🔍 Running in DRY RUN mode"; \
		$(GORUN) scripts/reset_tracking_for_testing.go -tracking_number=$(TRACKING_NUMBER) -dry_run; \
	else \
		echo "🚀 Executing reset"; \
		$(GORUN) scripts/reset_tracking_for_testing.go -tracking_number=$(TRACKING_NUMBER); \
	fi

# Cashback Management Commands
# Usage: make retry-cashback TRANSACTION_ID=<id>
retry-cashback:
ifndef TRANSACTION_ID
	@echo "Error: TRANSACTION_ID is required"
	@echo "Usage: make retry-cashback TRANSACTION_ID=7"
	@exit 1
endif
	@echo "🔄 Retrying cashback for transaction $(TRANSACTION_ID)..."
	@echo "⚠️  This will reset cashback status to pending and attempt to process it again"
	$(GORUN) scripts/retry_cashback.go $(TRANSACTION_ID)

# Usage: make transition-transaction TRANSACTION_ID=<id>
transition-transaction:
ifndef TRANSACTION_ID
	@echo "Error: TRANSACTION_ID is required"
	@echo "Usage: make transition-transaction TRANSACTION_ID=7"
	@exit 1
endif
	@echo "🔄 Transitioning transaction $(TRANSACTION_ID) from 'received' to 'remitted' state..."
	@echo "⚠️  This will only work if cashback is already completed"
	$(GORUN) scripts/transition_transaction_state.go $(TRANSACTION_ID)

# User Role Management Commands
# Usage: make update-user-role EMAIL=<email> ROLE=<role>
update-user-role:
ifndef EMAIL
	@echo "Error: EMAIL is required"
	@echo "Usage: make update-user-role EMAIL=user@example.com ROLE=admin"
	@exit 1
endif
ifndef ROLE
	@echo "Error: ROLE is required"  
	@echo "Usage: make update-user-role EMAIL=user@example.com ROLE=admin"
	@echo "Available roles: owner, admin, manager, support, user"
	@exit 1
endif
	@echo "🔧 Updating user role for $(EMAIL) to $(ROLE)..."
	$(GORUN) scripts/simple_update_user_role.go update $(EMAIL) $(ROLE)

# Usage: make show-user EMAIL=<email>
show-user:
ifndef EMAIL
	@echo "Error: EMAIL is required"
	@echo "Usage: make show-user EMAIL=user@example.com"
	@exit 1
endif
	@echo "👤 Showing user details for $(EMAIL)..."
	$(GORUN) scripts/simple_update_user_role.go show $(EMAIL)

# Usage: make list-users-by-role ROLE=<role>
list-users-by-role:
ifndef ROLE
	@echo "Error: ROLE is required"
	@echo "Usage: make list-users-by-role ROLE=admin"
	@echo "Available roles: owner, admin, manager, support, user"
	@exit 1
endif
	@echo "📋 Listing users with role: $(ROLE)..."
	$(GORUN) scripts/simple_update_user_role.go list $(ROLE)

# Show available user roles
show-user-roles:
	@echo "📊 Available user roles and descriptions:"
	$(GORUN) scripts/simple_update_user_role.go roles

# Password Reset Commands
reset-password:
	@if [ -z "$(EMAIL_OR_PHONE)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "❌ Usage: make reset-password EMAIL_OR_PHONE=<email_or_phone> PASSWORD=<new_password>"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make reset-password EMAIL_OR_PHONE=user@example.com PASSWORD=newpassword123"; \
		echo "  make reset-password EMAIL_OR_PHONE=+628123456789 PASSWORD=newpassword123"; \
		exit 1; \
	fi
	@echo "🔐 Resetting password for: $(EMAIL_OR_PHONE)"
	$(GORUN) scripts/reset_user_password.go "$(EMAIL_OR_PHONE)" "$(PASSWORD)"

# Interactive password reset using the standalone script
reset-password-interactive:
	@echo "🔐 Starting interactive password reset tool..."
	@chmod +x scripts/reset_password.sh
	@./scripts/reset_password.sh

# Password reset help
reset-password-help:
	@echo ""
	@echo "Password Reset Commands"
	@echo "======================"
	@echo ""
	@echo "Two methods are available for resetting user passwords:"
	@echo ""
	@echo "1. Command Line Reset:"
	@echo "   make reset-password EMAIL_OR_PHONE=<email_or_phone> PASSWORD=<new_password>"
	@echo ""
	@echo "   Examples:"
	@echo "     make reset-password EMAIL_OR_PHONE=user@example.com PASSWORD=newpassword123"
	@echo "     make reset-password EMAIL_OR_PHONE=+628123456789 PASSWORD=newpassword123"
	@echo ""
	@echo "2. Interactive Reset:"
	@echo "   make reset-password-interactive"
	@echo ""
	@echo "   This opens an interactive tool that will:"
	@echo "   - Prompt for database connection details"
	@echo "   - Search for the user by email or phone"
	@echo "   - Confirm user details before password reset"
	@echo "   - Securely prompt for new password (hidden input)"
	@echo "   - Optionally clear existing sessions/tokens"
	@echo ""
	@echo "Security Features:"
	@echo "   ✓ Uses Argon2id password hashing (same as your app)"
	@echo "   ✓ Generates new random salt for each password"
	@echo "   ✓ Validates minimum password length (8 characters)"
	@echo "   ✓ Hidden password input (not visible in terminal)"
	@echo "   ✓ User confirmation before executing reset"
	@echo "   ✓ Option to clear existing sessions for security"
	@echo ""

# Test retry deduct wallet endpoint
# Usage: make test-retry-deduct REFUND_ID=<id> JWT_TOKEN=<token> [NOTES="Custom notes"]
test-retry-deduct:
ifndef REFUND_ID
	@echo "Error: REFUND_ID is required"
	@echo "Usage: make test-retry-deduct REFUND_ID=01234567-89ab-cdef-0123-456789abcdef JWT_TOKEN=your_jwt_token"
	@exit 1
endif
ifndef JWT_TOKEN
	@echo "Error: JWT_TOKEN is required"
	@echo "Usage: make test-retry-deduct REFUND_ID=01234567-89ab-cdef-0123-456789abcdef JWT_TOKEN=your_jwt_token"
	@exit 1
endif
	@echo "🧪 Testing Retry Deduct Wallet Endpoint..."
	@if [ -n "$(NOTES)" ]; then \
		$(GORUN) test_retry_deduct_endpoint.go $(REFUND_ID) $(JWT_TOKEN) "$(NOTES)"; \
	else \
		$(GORUN) test_retry_deduct_endpoint.go $(REFUND_ID) $(JWT_TOKEN); \
	fi

# Test retry deduct wallet functionality (creates test refund request)
test-retry-deduct-functionality:
	@echo "🧪 Testing Retry Deduct Wallet Functionality..."
	@echo "This will create a test refund request and test the retry functionality"
	$(GORUN) test_retry_deduct_functionality.go

# Show help for reset-tracking command
reset-tracking-help:
	@echo ""
	@echo "Reset Tracking for Testing"
	@echo "=========================="
	@echo ""
	@echo "This command resets tracking status, transaction state, and cashback state for testing."
	@echo ""
	@echo "Usage:"
	@echo "  make reset-tracking TRACKING_NUMBER=<AWB> [DRY_RUN=true]"
	@echo ""
	@echo "Examples:"
	@echo "  # Dry run (see what would be changed)"
	@echo "  make reset-tracking TRACKING_NUMBER=JP123456789ID DRY_RUN=true"
	@echo ""
	@echo "  # Execute the reset"
	@echo "  make reset-tracking TRACKING_NUMBER=JP123456789ID"
	@echo ""
	@echo "What it does:"
	@echo "  1. 🔍 Find tracking record by tracking number"
	@echo "  2. 📊 Delete all tracking events"
	@echo "  3. 🔄 Reset tracking status to 'pickup_pending'"
	@echo "  4. 💰 Reset transaction state to 'paid' (if linked)"
	@echo "  5. 🎁 Reset cashback state to 'pending' and cashback record to 'pending' (if exists)"
	@echo "  6. 📝 Create initial tracking event"
	@echo "  7. 🧹 Clear all error states and circuit breakers"
	@echo ""
	@echo "⚠️  WARNING: This modifies production data. Always use DRY_RUN=true first!"
	@echo ""

# Show help for cashback commands
cashback-help:
	@echo ""
	@echo "Cashback Management Commands"
	@echo "============================"
	@echo ""
	@echo "These commands help manage cashback processing for transactions."
	@echo ""
	@echo "Commands:"
	@echo "  make retry-cashback TRANSACTION_ID=<id>        - Retry failed cashback processing (CLI script)"
	@echo "  make transition-transaction TRANSACTION_ID=<id> - Transition transaction state after cashback (CLI script)"
	@echo ""
	@echo "API Endpoints:"
	@echo "  POST /api/v1/admin/transactions/{id}/retry-cashback - Admin API endpoint to retry cashback"
	@echo ""
	@echo "Examples:"
	@echo "  # CLI: Retry cashback for transaction 7"
	@echo "  make retry-cashback TRANSACTION_ID=7"
	@echo ""
	@echo "  # CLI: Transition transaction 7 from 'received' to 'remitted'"
	@echo "  make transition-transaction TRANSACTION_ID=7"
	@echo ""
	@echo "  # API: Retry cashback via admin endpoint"
	@echo "  curl -X POST 'http://localhost:8080/api/v1/admin/transactions/7/retry-cashback' \\"
	@echo "       -H 'Authorization: Bearer <admin_jwt_token>' \\"
	@echo "       -H 'Content-Type: application/json'"
	@echo ""
	@echo "Cashback Flow:"
	@echo "  1. 🎁 Customer receives package → Transaction state: 'received'"
	@echo "  2. 💰 Cashback is processed → Cashback state: 'completed'"
	@echo "  3. 🏁 Transaction transitions → Transaction state: 'remitted' (final state)"
	@echo ""
	@echo "When to use:"
	@echo "  📈 retry-cashback (CLI & API):"
	@echo "    - When cashback processing failed (cashback_state: 'failed')"
	@echo "    - When cashback is stuck in 'pending' state"
	@echo "    - After fixing database constraints or other issues"
	@echo ""
	@echo "  🔄 transition-transaction (CLI only):"
	@echo "    - When cashback is completed but transaction is still in 'received' state"
	@echo "    - To fix stuck transactions that should be in 'remitted' state"
	@echo "    - Only works if cashback_state is already 'completed'"
	@echo ""
	@echo "⚠️  Prerequisites:"
	@echo "  - SSH tunnel to production database must be active (make tunnel-start)"
	@echo "  - Database migrations must be up to date"
	@echo "  - For retry-cashback: Ensure wallet_transactions constraint allows 'cashback' type"
	@echo "  - For API: Valid admin JWT token required"
	@echo ""
	@echo "💡 Quick troubleshooting:"
	@echo "  # Check transaction status"
	@echo "  psql -h localhost -p 15432 -U kirimku -d kirimku -c \"SELECT id, state, cashback, cashback_state FROM transactions WHERE id = <TRANSACTION_ID>;\""
	@echo ""
	@echo "  # Check cashback record"
	@echo "  psql -h localhost -p 15432 -U kirimku -d kirimku -c \"SELECT id, status, amount FROM cashbacks WHERE transaction_id = '<TRANSACTION_ID>';\""
	@echo ""
	@echo "🔗 API Documentation:"
	@echo "  - Full API docs: /api/openapi/admin-transaction-endpoints.yaml"
	@echo "  - Swagger UI: http://localhost:8080/swagger/index.html (if enabled)"
	@echo ""

# Help command to show available commands
help:
	@echo "Available commands:"
	@echo ""
	@echo "Basic Commands:"
	@echo "  make build          - Build the application"
	@echo "  make run           - Build and run the application"
	@echo "  make dev           - Run the application in development mode with hot reload"
	@echo "  make clean         - Clean build files"
	@echo "  make deps          - Download and install dependencies"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make check-coverage - Check test coverage"
	@echo ""
	@echo "Authentication Commands:"
	@echo "  make login EMAIL=<email> PASSWORD=<password>       - Login and get full JSON response"
	@echo "  make login-token EMAIL=<email> PASSWORD=<password> - Login and get only access token"
	@echo "  make login-env EMAIL=<email> PASSWORD=<password>   - Login and get environment variables"
	@echo "  make login-quick                                   - Login using LOGIN_EMAIL and LOGIN_PASSWORD env vars"
	@echo "  make login-quick-token                             - Get token using env vars"
	@echo "  make login-quick-env                               - Get env vars using env vars"
	@echo ""
	@echo "Database Commands:"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make migrate-status - Check migration status"
	@echo "  make migrate-force - Force migration version"
	@echo "  make create-migration - Create a new migration file"
	@echo "  make install-migrate - Install the migration tool"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make docker-run-detached - Run Docker container in background"
	@echo "  make docker-stop   - Stop and remove Docker container"
	@echo "  make docker-logs   - Show Docker container logs"
	@echo "  make docker-shell  - Open shell in running container"
	@echo "  make docker-push   - Push Docker image to registry"
	@echo "  make docker-clean  - Clean up Docker images and containers"
	@echo "  make docker-dev    - Build and run Docker container for development"
	@echo ""
	@echo "Preproduction Docker Commands:"
	@echo "  make docker-build-preprod  - Build Preproduction Docker image"
	@echo "  make docker-push-preprod   - Push Preproduction Docker image to registry"
	@echo "  make docker-run-preprod    - Run Preproduction Docker container"
	@echo "  make preprod-build-and-push - Complete preproduction pipeline: build and push"
	@echo "  make docker-clean-preprod  - Clean up Preproduction Docker images and containers"
	@echo ""
	@echo "GitHub Container Registry (GHCR) Commands:"
	@echo "  make ghcr-login    - Login to GitHub Container Registry"
	@echo "  make ghcr-build-and-push - Build and push image to GHCR"
	@echo "  make ghcr-build-and-push-multiarch - Build and push multi-arch image to GHCR"
	@echo "  make ghcr-push-all - Build, tag and push all versions to GHCR"
	@echo "  make release       - Build and push multi-architecture release"
	@echo ""
	@echo "Quick Workflows:"
	@echo "  make build-and-push - Build and push Docker image"
	@echo "  make build-and-run  - Build and run Docker container"
	@echo "  make dev-docker    - Start development container with volume mounts"
	@echo ""
	@echo "Mock Services:"
	@echo "  make mock-jnt      - Start the JNT mock server"
	@echo "  make mock-services - Start all mock services"
	@echo "  make dev-with-mocks - Run development mode with mock services"
	@echo ""
	@echo "Kiwi TCMS Commands:"
	@echo "  make kiwi-install-deps - Install Kiwi TCMS dependencies"
	@echo "  make kiwi-setup    - Set up Kiwi TCMS integration"
	@echo "  make kiwi-export-tests - Export test cases to Kiwi TCMS"
	@echo "  make kiwi-update-mapping - Update test case ID mappings"
	@echo "  make kiwi-run-tests - Run tests and report results to Kiwi TCMS"
	@echo "  make kiwi-export-auth-tests - Export auth test cases to Kiwi TCMS"
	@echo "  make kiwi-run-auth-tests - Run auth tests and report results to Kiwi TCMS"
	@echo ""
	@echo "SSH Tunnel Commands:"
	@echo "  make tunnel-start  - Start SSH tunnel to database"
	@echo "  make tunnel-stop   - Stop SSH tunnel to database"
	@echo "  make tunnel-status - Check SSH tunnel status"
	@echo "  make droplet-on    - Power on the DigitalOcean Droplet"
	@echo "  make droplet-off   - Power off the DigitalOcean Droplet"
	@echo "  make tunnel-migrate-up    - Run database migrations via SSH tunnel"
	@echo "  make tunnel-migrate-down  - Rollback database migrations via SSH tunnel"
	@echo "  make tunnel-migrate-status - Check migration status via SSH tunnel"
	@echo "  make tunnel-migrate-force VERSION=<version> - Force migration version via SSH tunnel"
	@echo ""
	@echo "Database Health Check Commands:"
	@echo "  make db-health-check          - Run full database health check and analysis"
	@echo "  make db-health-quick          - Run quick database health check (essential checks only)"
	@echo "  make db-analyze-blocking      - Analyze database blocking queries"
	@echo "  make db-fix-blocking          - Fix blocking queries (dry run preview)"
	@echo "  make db-fix-blocking-live     - Fix blocking queries (ACTUALLY EXECUTES - DANGEROUS!)"
	@echo "  make db-configure-prevention  - Configure DB for blocking prevention (dry run)"
	@echo "  make db-configure-prevention-live - Configure DB for blocking prevention (LIVE!)"
	@echo ""
	@echo "Cashback Management Commands:"
	@echo "  make retry-cashback TRANSACTION_ID=<id>        - Retry failed cashback processing"
	@echo "  make transition-transaction TRANSACTION_ID=<id> - Transition transaction state after cashback"
	@echo "  make cashback-help                             - Show detailed cashback command help"
	@echo ""
	@echo "User Role Management Commands:"
	@echo "  make update-user-role EMAIL=<email> ROLE=<role> - Update user role (owner/admin/manager/support/user)"
	@echo "  make show-user EMAIL=<email>                   - Show user details"
	@echo "  make list-users-by-role ROLE=<role>           - List users by role"
	@echo "  make show-user-roles                          - Show available roles and descriptions"
	@echo ""
	@echo "Password Reset Commands:"
	@echo "  make reset-password EMAIL_OR_PHONE=<email_or_phone> PASSWORD=<new_password> - Reset user password"
	@echo "  make reset-password-interactive                - Interactive password reset tool"
	@echo "  make reset-password-help                      - Show detailed password reset help"
	@echo ""
	@echo "Testing Commands:"
	@echo "  make reset-tracking TRACKING_NUMBER=<AWB>     - Reset tracking for testing"
	@echo "  make reset-tracking-help                      - Show detailed reset-tracking help"
	@echo "  make test-retry-deduct REFUND_ID=<id> JWT_TOKEN=<token> - Test retry deduct wallet endpoint"
	@echo ""
	@echo "Environment Variables for Docker:"
	@echo "  DOCKER_TAG=tag     - Set Docker image tag (default: latest)"
	@echo "  GITHUB_TOKEN=token - GitHub Personal Access Token for GHCR"
	@echo "  GITHUB_USER=user   - GitHub username for GHCR login"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-push    - Push Docker image to registry"
	@echo "  make docker-login   - Log in to Docker registry"
	@echo "  make docker-tag     - Tag Docker image"
	@echo "  make docker-clean   - Clean up Docker images"
	@echo "  make docker-dev     - Build and run Docker image for development"
	@echo ""
	@echo "GitHub Container Registry (GHCR) Commands:"
	@echo "  make ghcr-login     - Log in to GitHub Container Registry"
	@echo "  make ghcr-push      - Push image to GHCR"
	@echo "  make ghcr-tag-latest - Tag image as latest in GHCR"
	@echo "  make ghcr-push-latest - Push latest tag to GHCR"
	@echo "  make ghcr-push-all  - Tag and push all images to GHCR"