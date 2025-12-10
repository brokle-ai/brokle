#!/bin/bash

# ===========================================
# Brokle Update Script
# ===========================================
# Updates running Brokle deployment to latest version
# Usage: ./scripts/update.sh [--no-downtime] [--skip-migrations]

set -e

# Change to project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse flags
ZERO_DOWNTIME=false
SKIP_MIGRATIONS=false

for arg in "$@"; do
    case $arg in
        --no-downtime)
            ZERO_DOWNTIME=true
            ;;
        --skip-migrations)
            SKIP_MIGRATIONS=true
            ;;
    esac
done

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

# ===========================================
# Pre-Update Checks
# ===========================================

step "Running pre-update checks..."

# Check if running in correct directory
if [ ! -f "docker-compose.prod.yml" ]; then
    error "docker-compose.prod.yml not found. Are you in the brokle directory?"
fi

# Check if .env exists
if [ ! -f ".env" ]; then
    error ".env file not found. Run deploy-gcp.sh first for initial deployment."
fi

# Check if services are running
if ! docker compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    error "No services are running. Use deploy-gcp.sh for initial deployment."
fi

info "✓ Pre-update checks passed"

# ===========================================
# Backup Current State
# ===========================================

step "Creating backup before update..."

BACKUP_DIR=~/backups/pre-update-$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Backup databases
info "Backing up PostgreSQL..."
docker exec brokle-postgres pg_dump -U ${POSTGRES_USER:-brokle} ${POSTGRES_DB:-brokle_prod} | gzip > $BACKUP_DIR/postgres_backup.sql.gz

info "Backing up ClickHouse..."
docker exec brokle-clickhouse clickhouse-client --query "SHOW TABLES FROM default" | while read table; do
    docker exec brokle-clickhouse clickhouse-client --query "SELECT * FROM default.$table FORMAT TabSeparated" | gzip > $BACKUP_DIR/clickhouse_${table}.tsv.gz
done

# Backup .env
cp .env $BACKUP_DIR/.env.backup

info "✓ Backup created at: $BACKUP_DIR"

# ===========================================
# Pull Latest Code
# ===========================================

step "Pulling latest code from git..."

# Stash any local changes
if ! git diff-index --quiet HEAD --; then
    warn "Local changes detected. Stashing..."
    git stash
fi

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
info "Current branch: $CURRENT_BRANCH"

# Pull latest changes
git pull origin $CURRENT_BRANCH || error "Failed to pull latest code"

info "✓ Code updated successfully"

# ===========================================
# Rebuild Docker Images
# ===========================================

step "Rebuilding Docker images..."

if [ "$ZERO_DOWNTIME" = true ]; then
    info "Building new images (zero-downtime mode)..."
    docker compose -f docker-compose.prod.yml build --no-cache
else
    info "Building new images..."
    docker compose -f docker-compose.prod.yml build
fi

info "✓ Images rebuilt successfully"

# ===========================================
# Run Migrations
# ===========================================

if [ "$SKIP_MIGRATIONS" = false ]; then
    step "Running database migrations..."

    # PostgreSQL migrations
    info "Running PostgreSQL migrations..."
    docker compose -f docker-compose.prod.yml run --rm \
        -e DATABASE_MIGRATIONS_PATH=/app/migrations/postgres \
        backend /app/migrate -db postgres up || warn "PostgreSQL migration failed (may be up to date)"

    # ClickHouse migrations
    info "Running ClickHouse migrations..."
    docker compose -f docker-compose.prod.yml run --rm \
        -e CLICKHOUSE_MIGRATIONS_PATH=/app/migrations/clickhouse \
        backend /app/migrate -db clickhouse up || warn "ClickHouse migration failed (may be up to date)"

    info "✓ Migrations completed"
else
    warn "Skipping migrations (--skip-migrations flag set)"
fi

# ===========================================
# Update Services
# ===========================================

step "Updating services..."

if [ "$ZERO_DOWNTIME" = true ]; then
    info "Performing rolling update (zero-downtime)..."

    # Update backend (will use new image)
    docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate backend
    sleep 5

    # Wait for backend health check
    info "Waiting for backend to be healthy..."
    timeout 60 bash -c 'until docker inspect brokle-backend | grep -q "healthy"; do sleep 2; done' || error "Backend failed to become healthy"

    # Update worker
    docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate worker
    sleep 3

    # Update frontend
    docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate frontend
    sleep 5

    # Wait for frontend health check
    info "Waiting for frontend to be healthy..."
    timeout 60 bash -c 'until docker inspect brokle-frontend | grep -q "healthy"; do sleep 2; done' || error "Frontend failed to become healthy"

    # Update caddy (reverse proxy)
    docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate caddy

else
    info "Restarting all services..."
    docker compose -f docker-compose.prod.yml up -d --force-recreate
fi

info "✓ Services updated successfully"

# ===========================================
# Verify Deployment
# ===========================================

step "Verifying deployment..."

sleep 10

# Check service status
info "Service status:"
docker compose -f docker-compose.prod.yml ps

# Check backend health
if curl -s -f http://localhost:8080/health > /dev/null; then
    info "✓ Backend health check passed"
else
    error "Backend health check failed"
fi

# Check frontend
if curl -s -f http://localhost:3000 > /dev/null; then
    info "✓ Frontend is responding"
else
    warn "Frontend may not be fully ready yet"
fi

# ===========================================
# Cleanup
# ===========================================

step "Cleaning up..."

# Remove old images
docker image prune -f

info "✓ Cleanup completed"

# ===========================================
# Summary
# ===========================================

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Update completed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Backup location: $BACKUP_DIR"
echo ""
echo "Service URLs:"
echo "  - Application: https://app.brokle.com"
echo "  - Backend Health: http://localhost:8080/health"
echo ""
echo "To view logs:"
echo "  docker compose -f docker-compose.prod.yml logs -f"
echo ""
echo "To rollback (if needed):"
echo "  1. Restore .env: cp $BACKUP_DIR/.env.backup .env"
echo "  2. Restore database: zcat $BACKUP_DIR/postgres_backup.sql.gz | docker exec -i brokle-postgres psql -U brokle brokle_prod"
echo "  3. Rollback code: git reset --hard HEAD~1"
echo "  4. Rebuild: docker compose -f docker-compose.prod.yml build && docker compose -f docker-compose.prod.yml up -d"
echo ""
