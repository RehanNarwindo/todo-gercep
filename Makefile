# ================================================================
# MAKEFILE - Todo gercep Service Monitoring
# ================================================================
# Purpose: Manage all services (server, client, monitoring, prometheus)
# ================================================================

# ================================================================
# PHONY: Declare targets that are not actual files
# Always executes even if a file with the same name exists
# ================================================================
.PHONY: help run-server run-client run-monitor run-prometheus run-all stop-all

# ================================================================
# COLOR VARIABLES (for terminal output)
# ================================================================
GREEN  := \033[32m   # Green color for success messages
YELLOW := \033[33m   # Yellow color for warnings
RESET  := \033[0m    # Reset color to default

# ================================================================
# TARGET: help
# ================================================================
# Purpose: Display all available commands with descriptions
# Usage: make help
# ================================================================
help:
	@echo ''
	@echo '$(GREEN)Available commands:$(RESET)'
	@echo '  make run-server     - Run gRPC server (port 50051)'
	@echo '  make run-client     - Run gRPC client (CLI)'
	@echo '  make run-monitor    - Run monitoring service (port 3002)'
	@echo '  make run-prometheus - Run Prometheus (port 9090)'
	@echo '  make run-all        - Run all services'

# ================================================================
# TARGET: run-server
# ================================================================
# Purpose: Start the gRPC Server (Todo Service)
# Port: 50051
# Usage: make run-server
# ================================================================
run-server:
	@echo "$(GREEN)🚀 Starting gRPC Server...$(RESET)"  # Print notification message
	cd server && go run main.go                         # Navigate to server folder and run

# ================================================================
# TARGET: run-client
# ================================================================
# Purpose: Start the gRPC Client (Interactive CLI)
# Usage: make run-client
# ================================================================
run-client:
	@echo "$(GREEN)🚀 Starting gRPC Client...$(RESET)"  # Print notification message
	cd client && go run main.go                         # Navigate to client folder and run

# ================================================================
# TARGET: run-monitor
# ================================================================
# Purpose: Start the Monitoring Service
# Port: 3002
# Metrics endpoint: http://localhost:3002/metrics
# Usage: make run-monitor
# ================================================================
run-monitor:
	@echo "$(GREEN)📊 Starting Monitoring Service...$(RESET)"  # Print notification message
	cd monitoring && go run main.go                           # Navigate to monitoring folder and run

# ================================================================
# TARGET: run-prometheus
# ================================================================
# Purpose: Start Prometheus monitoring system
# Port: 9090
# Config file: monitoring/prometheus.yml
# Web UI: http://localhost:9090
# Usage: make run-prometheus
# ================================================================
run-prometheus:
	@echo "$(GREEN)📈 Starting Prometheus...$(RESET)"                          # Print notification message
	prometheus --config.file=monitoring/prometheus.yml                        # Run Prometheus with config

# ================================================================
# TARGET: run-all
# ================================================================
# Purpose: Display instructions to run all services
# Note: Requires 3 separate terminals (services run continuously)
# Usage: make run-all
# ================================================================
run-all:
	@echo "$(GREEN)🚀 Starting all services...$(RESET)"
	@echo "  Terminal 1: make run-server"      # Instruction for terminal 1
	@echo "  Terminal 2: make run-monitor"     # Instruction for terminal 2
	@echo "  Terminal 3: make run-prometheus"  # Instruction for terminal 3

# ================================================================
# TARGET: stop-all
# ================================================================
# Purpose: Stop all running Prometheus processes
# Usage: make stop-all
# ================================================================
stop-all:
	@echo "$(YELLOW)🛑 Stopping Prometheus...$(RESET)"                              # Print notification message
	-pkill -f "prometheus" 2>/dev/null || true                                     # Kill prometheus processes
	# -pkill      = kill process by name
	# -f          = match full command line (not just process name)
	# 2>/dev/null = redirect error output to null (ignore errors)
	# || true     = continue even if command fails (no error exit)
	@echo "$(GREEN)✅ Done$(RESET)"                                                 # Print completion message