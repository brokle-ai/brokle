#!/bin/bash

# =============================================================================
# Brokle Operations CLI
# =============================================================================
# Unified script for managing Brokle deployments
#
# Usage: ./brokle.sh <command> [options]
#
# Commands:
#   deploy              Fresh deployment (first-time setup)
#   update              Update existing deployment
#   migrate             Database migrations
#   status              Show service health and status
#   backup              Create database backups
#   logs                View service logs
#   help                Show this help message
# =============================================================================

set -e
set -o pipefail  # Pipeline fails if ANY command fails

# =============================================================================
# Configuration
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

COMPOSE_FILE="docker-compose.prod.yml"
BACKUP_BASE_DIR=~/backups

# =============================================================================
# Colors
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# =============================================================================
# Dry Run Mode
# =============================================================================

DRY_RUN=false

# Execute docker compose command or print in dry-run mode
docker_compose() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} docker compose -f $COMPOSE_FILE $*"
        return 0
    else
        docker compose -f "$COMPOSE_FILE" "$@"
    fi
}

# Execute docker command or print in dry-run mode
docker_cmd() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} docker $*"
        return 0
    else
        docker "$@"
    fi
}

# Execute git command or print in dry-run mode
git_cmd() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} git $*"
        return 0
    else
        git "$@"
    fi
}

# Sleep or print in dry-run mode
dry_sleep() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would wait $1 seconds"
    else
        sleep "$1"
    fi
}

# =============================================================================
# Output Functions
# =============================================================================

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

step() {
    echo -e "\n${BLUE}==>${NC} $1"
}

# =============================================================================
# Shared Validation Functions
# =============================================================================

check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker not found. Please install Docker first."
    fi

    if ! docker compose version &> /dev/null; then
        error "Docker Compose not found. Please install Docker Compose v2+"
    fi
}

check_env_file() {
    if [ ! -f ".env" ]; then
        error ".env file not found. Please create it:

    1. Copy the example file:
       cp .env.example .env

    2. Edit .env and configure ALL required values:
       - DOMAIN=your-domain.com
       - POSTGRES_PASSWORD (generate secure password)
       - CLICKHOUSE_PASSWORD (generate secure password)
       - JWT_SECRET (generate with: openssl rand -base64 64)
       - SESSION_SECRET (generate with: openssl rand -base64 64)
       - AI_KEY_ENCRYPTION_KEY (generate with: openssl rand -base64 32)
       - MINIO credentials

    3. Run this command again after configuring .env"
    fi

    # Check for CHANGE_ME values
    if grep -q "CHANGE_ME" .env; then
        error "Found CHANGE_ME values in .env file. Please update all credentials first."
    fi
}

validate_env_vars() {
    source .env

    if [ -z "$DOMAIN" ] || [ "$DOMAIN" == "your-domain.com" ]; then
        error "DOMAIN not configured in .env file"
    fi
    if [ -z "$JWT_SECRET" ]; then
        error "JWT_SECRET not configured in .env file"
    fi
    if [ -z "$DATABASE_URL" ]; then
        error "DATABASE_URL not configured in .env file"
    fi
}

secure_env_permissions() {
    ENV_PERMS=$(stat -c %a .env 2>/dev/null || stat -f %Lp .env 2>/dev/null)
    if [ "$ENV_PERMS" != "600" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo -e "${CYAN}[DRY-RUN]${NC} Would fix .env permissions: chmod 600 .env (current: $ENV_PERMS)"
        else
            warn ".env file permissions are $ENV_PERMS (should be 600). Fixing..."
            chmod 600 .env
            info "Fixed .env permissions"
        fi
    fi
}

check_disk_space() {
    local required=$1
    local available=$(df -BG / | tail -1 | awk '{print $4}' | sed 's/G//')
    if [ "$available" -lt "$required" ]; then
        error "Insufficient disk space. Available: ${available}GB, Required: ${required}GB minimum"
    fi
    info "Disk space: ${available}GB available"
}

check_caddyfile() {
    if [ ! -f "Caddyfile" ]; then
        error "Caddyfile not found. This file is required for the reverse proxy."
    fi

    if grep -q "your-domain.com" Caddyfile; then
        warn "Found 'your-domain.com' in Caddyfile. Please update with your actual domain."
    fi
}

check_firewall() {
    if command -v ufw &> /dev/null; then
        if sudo ufw status | grep -q "Status: active"; then
            local missing=false

            if ! sudo ufw status | grep -q "80/tcp"; then
                missing=true
            fi
            if ! sudo ufw status | grep -q "443/tcp"; then
                missing=true
            fi
            if ! sudo ufw status | grep -q "4317/tcp"; then
                missing=true
            fi

            if [ "$missing" = true ]; then
                warn "Required ports (80, 443, 4317) may not be open in UFW"
                echo ""
                echo "  Open required ports manually:"
                echo "    sudo ufw allow 80/tcp    # HTTP (Caddy)"
                echo "    sudo ufw allow 443/tcp   # HTTPS (Caddy)"
                echo "    sudo ufw allow 4317/tcp  # OTLP gRPC (Backend)"
                echo ""
            else
                info "Firewall ports open"
            fi
        fi
    fi
}

check_services_running() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Skipping: service running check"
        return 0
    fi
    if ! docker compose -f $COMPOSE_FILE ps 2>/dev/null | grep -q "Up"; then
        error "No services are running. Use 'deploy' for initial deployment."
    fi
}

check_no_services_running() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Skipping: no services running check"
        return 0
    fi
    if docker compose -f $COMPOSE_FILE ps 2>/dev/null | grep -q "Up"; then
        error "Services are already running. Use 'update' to update existing deployment."
    fi
}

# Wait for container to be healthy
# Args: $1 = container name, $2 = timeout seconds (default: 60)
# Returns: 0 if healthy, 1 if unhealthy or timeout
wait_for_healthy() {
    local container="$1"
    local timeout="${2:-60}"

    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would wait for $container to be healthy (timeout: ${timeout}s)"
        return 0
    fi

    local elapsed=0
    local interval=2

    while [ $elapsed -lt $timeout ]; do
        local status=$(docker inspect --format '{{.State.Health.Status}}' "$container" 2>/dev/null)

        if [ "$status" = "healthy" ]; then
            return 0
        elif [ "$status" = "unhealthy" ]; then
            return 1  # Fail immediately on unhealthy
        fi

        # Status is "starting" or container doesn't exist yet
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    return 1  # Timeout
}

# =============================================================================
# Database Functions
# =============================================================================

run_migrations() {
    local db="${1:-all}"
    local direction="${2:-up}"

    if [ "$db" == "all" ] || [ "$db" == "postgres" ]; then
        info "Running PostgreSQL migrations ($direction)..."
        if ! docker_compose run --rm \
            backend /app/brokle-server migrate -db postgres $direction; then
            error "PostgreSQL migration failed. Check database connectivity and migration files."
        fi
    fi

    if [ "$db" == "all" ] || [ "$db" == "clickhouse" ]; then
        info "Running ClickHouse migrations ($direction)..."
        if ! docker_compose run --rm \
            backend /app/brokle-server migrate -db clickhouse $direction; then
            error "ClickHouse migration failed. Check database connectivity and migration files."
        fi
    fi
}

migration_status() {
    local db="${1:-all}"

    if [ "$db" == "all" ] || [ "$db" == "postgres" ]; then
        info "PostgreSQL migration status:"
        docker_compose run --rm \
            backend /app/brokle-server migrate -db postgres status || warn "Could not get PostgreSQL status"
    fi

    if [ "$db" == "all" ] || [ "$db" == "clickhouse" ]; then
        info "ClickHouse migration status:"
        docker_compose run --rm \
            backend /app/brokle-server migrate -db clickhouse status || warn "Could not get ClickHouse status"
    fi
}

create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="${BACKUP_BASE_DIR}/backup-${timestamp}"

    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would create backup directory: $backup_dir"
        echo -e "${CYAN}[DRY-RUN]${NC} Would backup PostgreSQL to $backup_dir/postgres_backup.sql.gz"
        echo -e "${CYAN}[DRY-RUN]${NC} Would backup ClickHouse tables to $backup_dir/clickhouse/"
        echo -e "${CYAN}[DRY-RUN]${NC} Would backup .env to $backup_dir/.env.backup"
        echo "$backup_dir"
        return 0
    fi

    mkdir -p "$backup_dir"

    source .env

    # PostgreSQL backup with proper error handling
    info "Backing up PostgreSQL..."
    local pg_backup="$backup_dir/postgres_backup.sql.gz"

    # Run pipeline and capture PIPESTATUS for explicit error detection
    # stderr goes to terminal (visible in logs), stdout goes to gzip
    # Use || true to prevent set -e from exiting before we can check PIPESTATUS
    docker exec brokle-postgres pg_dump -U ${POSTGRES_USER:-brokle} ${POSTGRES_DB:-brokle_prod} | gzip > "$pg_backup" || true
    local pg_dump_exit=${PIPESTATUS[0]}
    local gzip_exit=${PIPESTATUS[1]}
    if [ $pg_dump_exit -ne 0 ]; then
        rm -f "$pg_backup"
        error "PostgreSQL backup failed (pg_dump exit code: $pg_dump_exit). Check if container is running: docker ps | grep postgres"
    fi
    if [ $gzip_exit -ne 0 ]; then
        rm -f "$pg_backup"
        error "PostgreSQL backup failed (gzip exit code: $gzip_exit). Check disk space."
    fi

    # Verify backup is not empty
    local pg_size=$(stat -c%s "$pg_backup" 2>/dev/null || stat -f%z "$pg_backup" 2>/dev/null || echo "0")
    if [ "$pg_size" -lt 100 ]; then
        rm -f "$pg_backup"
        error "PostgreSQL backup appears empty or corrupt (${pg_size} bytes)"
    fi
    info "PostgreSQL backup: $(du -h "$pg_backup" | cut -f1)"

    # ClickHouse backup with proper error handling
    info "Backing up ClickHouse..."
    mkdir -p "$backup_dir/clickhouse"

    local ch_db="${CLICKHOUSE_DB:-default}"
    local tables
    # stderr goes to terminal for visibility
    if ! tables=$(docker exec brokle-clickhouse clickhouse-client --query "SHOW TABLES FROM $ch_db"); then
        error "Failed to list ClickHouse tables. Check if container is running: docker ps | grep clickhouse"
    fi

    if [ -z "$tables" ]; then
        warn "No ClickHouse tables found to backup in database '$ch_db'"
    else
        local failed_tables=()
        while IFS= read -r table; do
            if [ -n "$table" ]; then
                local ch_backup="$backup_dir/clickhouse/${table}.tsv.gz"
                # Run pipeline and capture PIPESTATUS for explicit error detection
                # stderr goes to terminal (visible in logs), stdout goes to gzip
                docker exec brokle-clickhouse clickhouse-client --query "SELECT * FROM $ch_db.$table FORMAT TabSeparated" | gzip > "$ch_backup" || true
                local ch_dump_exit=${PIPESTATUS[0]}
                local ch_gzip_exit=${PIPESTATUS[1]}
                if [ $ch_dump_exit -ne 0 ] || [ $ch_gzip_exit -ne 0 ]; then
                    rm -f "$ch_backup"
                    failed_tables+=("$table")
                fi
            fi
        done <<< "$tables"

        if [ ${#failed_tables[@]} -gt 0 ]; then
            error "ClickHouse backup failed for tables: ${failed_tables[*]}"
        fi
        info "ClickHouse backup: $(du -sh "$backup_dir/clickhouse" | cut -f1)"
    fi

    # Backup .env
    cp .env "$backup_dir/.env.backup"

    # Final verification
    info "Verifying backup integrity..."
    local backup_ok=true

    if [ ! -f "$backup_dir/postgres_backup.sql.gz" ]; then
        warn "Missing: PostgreSQL backup"
        backup_ok=false
    fi

    if [ ! -d "$backup_dir/clickhouse" ] || [ -z "$(ls -A "$backup_dir/clickhouse" 2>/dev/null)" ]; then
        warn "Missing: ClickHouse backup (may be empty database)"
    fi

    if [ ! -f "$backup_dir/.env.backup" ]; then
        warn "Missing: .env backup"
        backup_ok=false
    fi

    if [ "$backup_ok" = false ]; then
        error "Backup verification failed. Check errors above."
    fi

    info "Backup completed and verified: $backup_dir"
    echo "$backup_dir"
}

cleanup_old_backups() {
    local days=${1:-7}

    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would delete backups older than $days days from $BACKUP_BASE_DIR"
        if [ -d "$BACKUP_BASE_DIR" ]; then
            local count=$(find "$BACKUP_BASE_DIR" -maxdepth 1 -type d \( -name "backup-*" -o -name "pre-update-*" \) -mtime +$days 2>/dev/null | wc -l)
            echo -e "${CYAN}[DRY-RUN]${NC} Found $count directories that would be deleted"
        fi
        return 0
    fi

    if [ -d "$BACKUP_BASE_DIR" ]; then
        find "$BACKUP_BASE_DIR" -maxdepth 1 -type d -name "backup-*" -mtime +$days -exec rm -rf {} \; 2>/dev/null || true
        find "$BACKUP_BASE_DIR" -maxdepth 1 -type d -name "pre-update-*" -mtime +$days -exec rm -rf {} \; 2>/dev/null || true
    fi
}

# =============================================================================
# Command: help
# =============================================================================

cmd_help() {
    echo ""
    echo -e "${CYAN}Brokle Operations CLI${NC}"
    echo ""
    echo "Usage: ./brokle.sh [--dry-run] <command> [options]"
    echo ""
    echo "Global Options:"
    echo "  --dry-run           Print commands instead of executing them"
    echo ""
    echo "Commands:"
    echo "  deploy              Fresh deployment (first-time setup)"
    echo "  update              Update existing deployment"
    echo "    --no-downtime     Rolling update with zero downtime"
    echo "    --skip-migrations Skip database migrations"
    echo "  migrate             Database migrations (default: up)"
    echo "    --status          Show migration status"
    echo "    --down            Rollback one migration"
    echo "    --db <name>       Target database (postgres|clickhouse|all)"
    echo "  status              Show service health and status"
    echo "  backup              Create database backups"
    echo "    --cleanup [days]  Also cleanup backups older than N days (default: 7)"
    echo "  logs [service]      View logs (all services or specific one)"
    echo "    -f, --follow      Follow log output (default)"
    echo "    --tail <n>        Number of lines to show"
    echo "  help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./brokle.sh deploy"
    echo "  ./brokle.sh update --no-downtime"
    echo "  ./brokle.sh --dry-run deploy          # Preview deploy commands"
    echo "  ./brokle.sh --dry-run update          # Preview update commands"
    echo "  ./brokle.sh migrate --status"
    echo "  ./brokle.sh migrate --down --db postgres"
    echo "  ./brokle.sh logs backend"
    echo "  ./brokle.sh backup --cleanup 14"
    echo ""
}

# =============================================================================
# Command: status
# =============================================================================

cmd_status() {
    step "Brokle Service Status"

    echo ""
    echo -e "${BLUE}Services:${NC}"
    docker_compose ps 2>/dev/null || warn "Could not get service status"

    echo ""
    echo -e "${BLUE}Backend Health:${NC}"
    # Use docker exec to check health inside container (works in both dev and prod)
    if docker exec brokle-backend wget -qO- http://localhost:8080/health 2>/dev/null >/dev/null; then
        local health=$(docker exec brokle-backend wget -qO- http://localhost:8080/health 2>/dev/null)
        echo -e "${GREEN}Healthy${NC}"
        echo "$health" | jq '.' 2>/dev/null || echo "$health"
    else
        echo -e "${RED}Not responding${NC}"
    fi

    echo ""
    echo -e "${BLUE}Frontend:${NC}"
    # Use docker exec to check frontend inside container (works in both dev and prod)
    if docker exec brokle-frontend wget -qO- http://localhost:3000 2>/dev/null >/dev/null; then
        echo -e "${GREEN}Responding${NC}"
    else
        echo -e "${YELLOW}Not responding (may be normal if container is starting)${NC}"
    fi

    echo ""
    echo -e "${BLUE}System Resources:${NC}"
    echo "Disk Usage:"
    df -h / | grep -v tmpfs
    echo ""
    echo "Memory Usage:"
    free -h 2>/dev/null || vm_stat 2>/dev/null || echo "Could not get memory info"

    echo ""
    echo -e "${BLUE}Docker Stats (snapshot):${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | head -10 || warn "Could not get Docker stats"
}

# =============================================================================
# Command: logs
# =============================================================================

cmd_logs() {
    local service=""
    local follow=true
    local tail=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--follow)
                follow=true
                shift
                ;;
            --no-follow)
                follow=false
                shift
                ;;
            --tail)
                tail="$2"
                shift 2
                ;;
            -*)
                error "Unknown option: $1"
                ;;
            *)
                service="$1"
                shift
                ;;
        esac
    done

    # Build args array
    local args=("logs")

    if [ "$follow" = true ]; then
        args+=("-f")
    fi

    if [ -n "$tail" ]; then
        args+=("--tail" "$tail")
    fi

    if [ -n "$service" ]; then
        args+=("$service")
    fi

    # Use wrapper (respects DRY_RUN)
    docker_compose "${args[@]}"
}

# =============================================================================
# Command: backup
# =============================================================================

cmd_backup() {
    local cleanup=false
    local cleanup_days=7

    while [[ $# -gt 0 ]]; do
        case $1 in
            --cleanup)
                cleanup=true
                if [[ "$2" =~ ^[0-9]+$ ]]; then
                    cleanup_days="$2"
                    shift
                fi
                shift
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done

    step "Creating database backups"

    check_env_file

    local backup_dir=$(create_backup)

    if [ "$cleanup" = true ]; then
        step "Cleaning up old backups (older than $cleanup_days days)"
        cleanup_old_backups $cleanup_days
        info "Cleanup completed"
    fi

    echo ""
    echo -e "${GREEN}Backup completed!${NC}"
    echo ""
    echo "Backup location: $backup_dir"
    echo ""
    echo "To restore PostgreSQL:"
    echo "  zcat $backup_dir/postgres_backup.sql.gz | docker exec -i brokle-postgres psql -U brokle brokle_prod"
    echo ""
}

# =============================================================================
# Command: migrate
# =============================================================================

cmd_migrate() {
    local db="all"
    local action="up"

    while [[ $# -gt 0 ]]; do
        case $1 in
            --status)
                action="status"
                shift
                ;;
            --down)
                action="down"
                shift
                ;;
            --up)
                action="up"
                shift
                ;;
            --db)
                db="$2"
                if [[ ! "$db" =~ ^(postgres|clickhouse|all)$ ]]; then
                    error "Invalid database: $db. Use: postgres, clickhouse, or all"
                fi
                shift 2
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done

    step "Database Migrations"

    check_docker

    if [ "$action" == "status" ]; then
        migration_status "$db"
    else
        run_migrations "$db" "$action"
        info "Migrations completed"
    fi
}

# =============================================================================
# Command: deploy
# =============================================================================

cmd_deploy() {
    step "Starting Brokle Deployment"

    # Check Ubuntu
    if [ -f /etc/os-release ]; then
        source /etc/os-release
        if [[ "$ID" != "ubuntu" ]]; then
            warn "This script is designed for Ubuntu. Detected: $ID"
        fi
        info "Deploying on $PRETTY_NAME"
    fi

    # Prerequisites
    step "Checking prerequisites..."
    check_docker
    info "Docker found: $(docker --version)"
    info "Docker Compose found: $(docker compose version)"

    # Check git
    if ! command -v git &> /dev/null; then
        if [ "$DRY_RUN" = true ]; then
            echo -e "${CYAN}[DRY-RUN]${NC} Would install git: sudo apt update && sudo apt install -y git"
            echo -e "${CYAN}[DRY-RUN]${NC} git not currently installed"
        else
            warn "Git not found. Installing git..."
            sudo apt update && sudo apt install -y git
            info "Git found: $(git --version)"
        fi
    else
        info "Git found: $(git --version)"
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        if [ "$DRY_RUN" = true ]; then
            echo -e "${CYAN}[DRY-RUN]${NC} Would install jq: sudo apt update && sudo apt install -y jq"
            echo -e "${CYAN}[DRY-RUN]${NC} jq not currently installed"
        else
            warn "jq not found. Installing jq..."
            sudo apt update && sudo apt install -y jq
            info "jq found: $(jq --version)"
        fi
    else
        info "jq found: $(jq --version)"
    fi

    # Environment
    step "Checking environment configuration..."
    check_env_file
    info ".env file found"

    secure_env_permissions
    info ".env file permissions secured (600)"

    validate_env_vars
    info "Critical environment variables configured"

    # Caddyfile
    step "Checking Caddyfile..."
    check_caddyfile
    info "Caddyfile found"

    # Disk space
    step "Checking disk space..."
    check_disk_space 20

    # Firewall
    step "Checking firewall..."
    check_firewall

    # Check no existing deployment
    check_no_services_running

    # Create directories
    step "Creating required directories..."
    mkdir -p logs
    mkdir -p "$BACKUP_BASE_DIR/postgres"
    mkdir -p "$BACKUP_BASE_DIR/clickhouse"
    info "Directories created"

    # Build images
    step "Building Docker images (this may take 5-10 minutes)..."
    if ! docker_compose build; then
        error "Failed to build Docker images"
    fi
    info "Docker images built successfully"

    # Start databases
    step "Starting database services..."
    docker_compose down 2>/dev/null || true

    if ! docker_compose up -d postgres clickhouse redis; then
        error "Failed to start database services"
    fi

    info "Waiting for databases to be ready..."
    dry_sleep 15
    info "Database services started"

    # Run migrations
    step "Running database migrations..."
    run_migrations "all" "up"
    info "Migrations completed"

    # Start all services
    step "Starting all services..."
    if ! docker_compose up -d; then
        error "Failed to start services"
    fi
    info "Services started"

    # Wait for healthy
    step "Waiting for services to be healthy..."
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would wait for backend health check (max 120s)"
        info "Backend is healthy"
    else
        local max_wait=120
        local elapsed=0
        local interval=5

        while [ $elapsed -lt $max_wait ]; do
            sleep $interval
            elapsed=$((elapsed + interval))

            # Use docker exec to check health inside container (works in both dev and prod)
            if docker exec brokle-backend wget -qO- http://localhost:8080/health 2>/dev/null >/dev/null; then
                info "Backend is healthy"
                break
            fi

            echo -n "."
        done

        if [ $elapsed -ge $max_wait ]; then
            error "Services did not become healthy within ${max_wait} seconds. Check logs with: ./brokle.sh logs"
        fi
    fi

    # Verify
    step "Verifying deployment..."
    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would verify all services are running"
        info "All services are running"
    else
        local services=("postgres" "clickhouse" "redis" "backend" "worker" "frontend" "caddy" "minio")
        local failed=()

        for service in "${services[@]}"; do
            if ! docker ps | grep -q "brokle-${service}"; then
                failed+=("$service")
            fi
        done

        if [ ${#failed[@]} -gt 0 ]; then
            error "The following services are not running: ${failed[*]}"
        fi

        info "All services are running"
    fi

    # Load domain for output
    source .env

    # Summary
    echo ""
    echo -e "${GREEN}==============================================${NC}"
    echo -e "${GREEN}  Brokle Deployment Successful!${NC}"
    echo -e "${GREEN}==============================================${NC}"
    echo ""
    echo "Services Status:"
    docker_compose ps
    echo ""
    echo "Access Points:"
    echo "  - Dashboard: https://$DOMAIN"
    echo "  - API: https://$DOMAIN/api/v1"
    echo "  - OTLP gRPC: $DOMAIN:4317"
    echo "  - Health Check: https://$DOMAIN/health"
    echo ""
    echo "Local Access (for debugging):"
    echo "  - Backend: http://localhost:8080"
    echo "  - Frontend: http://localhost:3000"
    echo ""
    echo "Next Steps:"
    echo "  1. Verify DNS propagation: nslookup $DOMAIN"
    echo "  2. Check SSL certificate: curl -I https://$DOMAIN"
    echo "  3. View logs: ./brokle.sh logs"
    echo "  4. Create admin user via dashboard"
    echo "  5. Generate API key for SDK integration"
    echo ""
    echo "Management Commands:"
    echo "  - View logs: ./brokle.sh logs"
    echo "  - Check status: ./brokle.sh status"
    echo "  - Update: ./brokle.sh update"
    echo "  - Backup: ./brokle.sh backup"
    echo ""

    info "Deployment complete!"
}

# =============================================================================
# Command: update
# =============================================================================

cmd_update() {
    local zero_downtime=false
    local skip_migrations=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-downtime)
                zero_downtime=true
                shift
                ;;
            --skip-migrations)
                skip_migrations=true
                shift
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done

    step "Starting Brokle Update"

    # Pre-update checks
    step "Running pre-update checks..."

    # Check compose file
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "$COMPOSE_FILE not found. Are you in the brokle directory?"
    fi

    check_docker
    info "Docker found"

    check_env_file
    info ".env file found"

    validate_env_vars
    info "Environment variables validated"

    secure_env_permissions

    check_caddyfile
    info "Caddyfile found"

    check_disk_space 5

    check_firewall

    check_services_running
    info "Services are running"

    info "Pre-update checks passed"

    # Backup
    step "Creating backup before update..."
    local backup_dir=$(create_backup)
    info "Backup created at: $backup_dir"

    # Pull code
    step "Pulling latest code from git..."

    # Block if uncommitted changes exist (user should handle explicitly)
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        if [ "$DRY_RUN" = true ]; then
            warn "Uncommitted changes detected (would block in real run)"
        else
            error "Uncommitted changes detected. Please commit, stash, or discard before updating:
    git stash      # Save changes for later
    git checkout . # Discard changes
    git commit     # Keep changes permanently"
        fi
    fi

    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    info "Current branch: $current_branch"

    git_cmd pull origin $current_branch || error "Failed to pull latest code"
    info "Code updated successfully"

    # Rebuild images
    step "Rebuilding Docker images..."

    if [ "$zero_downtime" = true ]; then
        info "Building new images (zero-downtime mode)..."
        docker_compose build --no-cache
    else
        info "Building new images..."
        docker_compose build
    fi

    info "Images rebuilt successfully"

    # Migrations
    if [ "$skip_migrations" = false ]; then
        step "Running database migrations..."
        run_migrations "all" "up"
        info "Migrations completed"
    else
        warn "Skipping migrations (--skip-migrations flag set)"
    fi

    # Update services
    step "Updating services..."

    if [ "$zero_downtime" = true ]; then
        info "Performing rolling update (zero-downtime)..."

        # Update backend
        docker_compose up -d --no-deps --force-recreate backend
        dry_sleep 5

        # Wait for backend health
        info "Waiting for backend to be healthy..."
        if ! wait_for_healthy "brokle-backend" 60; then
            error "Backend failed to become healthy"
        fi

        # Update worker
        docker_compose up -d --no-deps --force-recreate worker
        dry_sleep 3

        # Update frontend
        docker_compose up -d --no-deps --force-recreate frontend
        dry_sleep 5

        # Wait for frontend
        info "Waiting for frontend to be healthy..."
        if ! wait_for_healthy "brokle-frontend" 60; then
            error "Frontend failed to become healthy"
        fi

        # Update caddy
        docker_compose up -d --no-deps --force-recreate caddy
    else
        info "Restarting all services..."
        docker_compose up -d --force-recreate
    fi

    info "Services updated successfully"

    # Verify
    step "Verifying deployment..."
    dry_sleep 10

    info "Service status:"
    docker_compose ps

    if [ "$DRY_RUN" = true ]; then
        echo -e "${CYAN}[DRY-RUN]${NC} Would check backend health via docker exec"
        echo -e "${CYAN}[DRY-RUN]${NC} Would check frontend via docker exec"
    else
        # Use docker exec to check health inside container (works in both dev and prod)
        if docker exec brokle-backend wget -qO- http://localhost:8080/health 2>/dev/null >/dev/null; then
            info "Backend health check passed"
        else
            error "Backend health check failed"
        fi

        # Use docker exec to check frontend inside container (works in both dev and prod)
        if docker exec brokle-frontend wget -qO- http://localhost:3000 2>/dev/null >/dev/null; then
            info "Frontend is responding"
        else
            warn "Frontend may not be fully ready yet"
        fi
    fi

    # Cleanup
    step "Cleaning up..."
    docker_cmd image prune -f
    info "Cleanup completed"

    # Load domain for output
    source .env

    # Summary
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}Update completed successfully!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    echo "Backup location: $backup_dir"
    echo ""
    echo "Service URLs:"
    echo "  - Application: https://${DOMAIN:-your-domain.com}"
    echo "  - Backend Health: http://localhost:8080/health"
    echo ""
    echo "To view logs:"
    echo "  ./brokle.sh logs"
    echo ""
    echo "To rollback (if needed):"
    echo "  1. Restore .env: cp $backup_dir/.env.backup .env"
    echo "  2. Restore database: zcat $backup_dir/postgres_backup.sql.gz | docker exec -i brokle-postgres psql -U brokle brokle_prod"
    echo "  3. Rollback code: git reset --hard HEAD~1"
    echo "  4. Rebuild: docker compose -f $COMPOSE_FILE build && docker compose -f $COMPOSE_FILE up -d"
    echo ""
}

# =============================================================================
# Main Router
# =============================================================================

# Parse global flags
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            info "Dry-run mode enabled - no changes will be made"
            shift
            ;;
        *)
            break
            ;;
    esac
done

case "${1:-help}" in
    deploy)
        shift
        cmd_deploy "$@"
        ;;
    update)
        shift
        cmd_update "$@"
        ;;
    migrate)
        shift
        cmd_migrate "$@"
        ;;
    status)
        shift
        cmd_status "$@"
        ;;
    backup)
        shift
        cmd_backup "$@"
        ;;
    logs)
        shift
        cmd_logs "$@"
        ;;
    help|--help|-h)
        cmd_help
        ;;
    *)
        error "Unknown command: $1. Run './brokle.sh help' for usage."
        ;;
esac
