#!/bin/bash
set -e  # Exit on error

# =============================================================================
# Brokle GCP Deployment Script
# =============================================================================
# This script automates the deployment of Brokle to GCP VM using Docker Compose
# Usage: ./deploy-gcp.sh
# =============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
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

# Check if running on Ubuntu
if [ ! -f /etc/os-release ]; then
    error "Cannot detect OS. This script is designed for Ubuntu."
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" ]]; then
    error "This script is designed for Ubuntu. Detected: $ID"
fi

info "Starting Brokle deployment on $PRETTY_NAME"

# =============================================================================
# 1. Prerequisites Check
# =============================================================================
info "Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    warn "Docker not found. Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
    info "Docker installed. You may need to log out and back in for group changes to take effect."
    info "After logging back in, run this script again."
    exit 0
else
    info "✓ Docker found: $(docker --version)"
fi

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    error "Docker Compose not found. Please install Docker Compose v2+"
fi
info "✓ Docker Compose found: $(docker compose version)"

# Check git
if ! command -v git &> /dev/null; then
    warn "Git not found. Installing git..."
    sudo apt update
    sudo apt install -y git
fi
info "✓ Git found: $(git --version)"

# Check jq
if ! command -v jq &> /dev/null; then
    warn "jq not found. Installing jq..."
    sudo apt update
    sudo apt install -y jq
fi
info "✓ jq found: $(jq --version)"

# =============================================================================
# 2. Environment Configuration
# =============================================================================
info "Checking environment configuration..."

if [ ! -f .env ]; then
    if [ -f .env.production ]; then
        warn ".env not found. Creating from .env.production template..."
        cp .env.production .env
        warn "⚠️  IMPORTANT: Edit .env file and update all CHANGE_ME values!"
        warn "    Required changes:"
        warn "    - DOMAIN=your-domain.com"
        warn "    - POSTGRES_PASSWORD"
        warn "    - CLICKHOUSE_PASSWORD"
        warn "    - JWT_SECRET (generate with: openssl rand -base64 64)"
        warn "    - SESSION_SECRET (generate with: openssl rand -base64 64)"
        warn "    - MINIO credentials"
        echo ""
        read -p "Have you updated all credentials in .env? (yes/no): " confirm
        if [[ "$confirm" != "yes" ]]; then
            error "Please update .env file and run this script again."
        fi
    else
        error ".env.production template not found. Cannot create .env file."
    fi
else
    info "✓ .env file found"
fi

# Check for CHANGE_ME values
if grep -q "CHANGE_ME" .env; then
    error "Found CHANGE_ME values in .env file. Please update all credentials first."
fi

# Secure .env file
chmod 600 .env
info "✓ .env file permissions secured (600)"

# Validate critical environment variables
source .env
if [ -z "$DOMAIN" ] || [ "$DOMAIN" == "your-domain.com" ]; then
    error "DOMAIN not configured in .env file"
fi
if [ -z "$JWT_SECRET" ]; then
    error "JWT_SECRET not configured in .env file"
fi
if [ -z "$SESSION_SECRET" ]; then
    error "SESSION_SECRET not configured in .env file"
fi
info "✓ Critical environment variables configured"

# =============================================================================
# 3. Caddyfile Configuration
# =============================================================================
info "Checking Caddyfile configuration..."

if [ ! -f Caddyfile ]; then
    error "Caddyfile not found. Please ensure Caddyfile exists in the project root."
fi

if grep -q "your-domain.com" Caddyfile; then
    warn "Found 'your-domain.com' in Caddyfile. Please update with your actual domain."
    read -p "Have you updated Caddyfile with your domain? (yes/no): " confirm
    if [[ "$confirm" != "yes" ]]; then
        error "Please update Caddyfile and run this script again."
    fi
fi
info "✓ Caddyfile configured"

# =============================================================================
# 4. Disk Space Check
# =============================================================================
info "Checking disk space..."

AVAILABLE_GB=$(df -BG / | tail -1 | awk '{print $4}' | sed 's/G//')
if [ "$AVAILABLE_GB" -lt 20 ]; then
    error "Insufficient disk space. Available: ${AVAILABLE_GB}GB, Required: 20GB minimum"
fi
info "✓ Sufficient disk space: ${AVAILABLE_GB}GB available"

# =============================================================================
# 5. Firewall Check
# =============================================================================
info "Checking firewall configuration..."

if command -v ufw &> /dev/null; then
    if sudo ufw status | grep -q "Status: active"; then
        info "UFW firewall is active"

        # Check if required ports are open
        if ! sudo ufw status | grep -q "80/tcp"; then
            warn "Port 80 (HTTP) not open in UFW. Opening..."
            sudo ufw allow 80/tcp
        fi
        if ! sudo ufw status | grep -q "443/tcp"; then
            warn "Port 443 (HTTPS) not open in UFW. Opening..."
            sudo ufw allow 443/tcp
        fi
        if ! sudo ufw status | grep -q "4317/tcp"; then
            warn "Port 4317 (OTLP gRPC) not open in UFW. Opening..."
            sudo ufw allow 4317/tcp
        fi
        info "✓ Required ports (80, 443, 4317) are open"
    fi
else
    info "UFW not installed (OK for GCP - using GCP firewall rules)"
fi

# =============================================================================
# 6. Create Required Directories
# =============================================================================
info "Creating required directories..."

mkdir -p logs
mkdir -p backups/postgres
mkdir -p backups/clickhouse

info "✓ Directories created"

# =============================================================================
# 7. Pull/Build Docker Images
# =============================================================================
info "Building Docker images (this may take 5-10 minutes)..."

if ! docker compose -f docker-compose.prod.yml build; then
    error "Failed to build Docker images"
fi

info "✓ Docker images built successfully"

# =============================================================================
# 8. Start Services
# =============================================================================
info "Starting services..."

# Stop any existing services
docker compose -f docker-compose.prod.yml down 2>/dev/null || true

# Start services
if ! docker compose -f docker-compose.prod.yml up -d; then
    error "Failed to start services"
fi

info "✓ Services started"

# =============================================================================
# 9. Wait for Services to be Healthy
# =============================================================================
info "Waiting for services to be healthy (this may take 30-60 seconds)..."

MAX_WAIT=120  # 2 minutes
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $MAX_WAIT ]; do
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))

    # Check if backend is healthy
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        info "✓ Backend is healthy"
        break
    fi

    echo -n "."
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    error "Services did not become healthy within ${MAX_WAIT} seconds. Check logs with: docker compose -f docker-compose.prod.yml logs"
fi

# =============================================================================
# 10. Verify Deployment
# =============================================================================
info "Verifying deployment..."

# Check service status
SERVICES=("postgres" "clickhouse" "redis" "backend" "worker" "frontend" "caddy" "minio")
FAILED_SERVICES=()

for service in "${SERVICES[@]}"; do
    if ! docker ps | grep -q "brokle-${service}"; then
        FAILED_SERVICES+=("$service")
    fi
done

if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
    error "The following services are not running: ${FAILED_SERVICES[*]}"
fi

info "✓ All services are running"

# Check backend health endpoint
if ! curl -s http://localhost:8080/health | grep -q "healthy"; then
    warn "Backend health check returned unexpected response"
fi

info "✓ Backend health check passed"

# =============================================================================
# 11. Display Status
# =============================================================================
echo ""
echo "=============================================="
echo "  🚀 Brokle Deployment Successful!"
echo "=============================================="
echo ""
echo "Services Status:"
docker compose -f docker-compose.prod.yml ps
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
echo "  3. View logs: docker compose -f docker-compose.prod.yml logs -f"
echo "  4. Create admin user via dashboard"
echo "  5. Generate API key for SDK integration"
echo ""
echo "Management Commands:"
echo "  - View logs: docker compose -f docker-compose.prod.yml logs -f"
echo "  - Restart: docker compose -f docker-compose.prod.yml restart"
echo "  - Stop: docker compose -f docker-compose.prod.yml down"
echo "  - Update: git pull && ./deploy-gcp.sh"
echo ""
echo "Documentation:"
echo "  - Full deployment guide: docs/DEPLOYMENT_GCP.md"
echo "  - Architecture: docs/ARCHITECTURE.md"
echo "  - API reference: docs/API.md"
echo ""
echo "=============================================="
echo ""

# =============================================================================
# 12. Create Helper Scripts
# =============================================================================
info "Creating helper scripts..."

# Health check script
cat > check-health.sh << 'EOFHEALTH'
#!/bin/bash
echo "=== Brokle Health Check ==="
echo ""
echo "Services:"
docker compose -f docker-compose.prod.yml ps
echo ""
echo "Backend Health:"
curl -s http://localhost:8080/health | jq '.' 2>/dev/null || curl -s http://localhost:8080/health
echo ""
echo "Disk Usage:"
df -h / | grep -v tmpfs
echo ""
echo "Memory Usage:"
free -h
echo ""
echo "Docker Stats:"
docker stats --no-stream
EOFHEALTH

chmod +x check-health.sh

# Backup script
cat > backup.sh << 'EOFBACKUP'
#!/bin/bash
# Source environment variables
if [ -f .env ]; then
    source .env
else
    echo "Error: .env file not found"
    exit 1
fi

# PostgreSQL Backup
BACKUP_DIR=~/backups/postgres
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
echo "Creating PostgreSQL backup..."
docker exec brokle-postgres pg_dump -U ${POSTGRES_USER} ${POSTGRES_DB} | gzip > $BACKUP_DIR/backup_$TIMESTAMP.sql.gz
echo "Backup created: $BACKUP_DIR/backup_$TIMESTAMP.sql.gz"

# ClickHouse Backup
echo ""
echo "Creating ClickHouse backup..."
CLICKHOUSE_BACKUP_DIR=~/backups/clickhouse
mkdir -p $CLICKHOUSE_BACKUP_DIR

# Export schema
docker exec brokle-clickhouse clickhouse-client --query "SHOW CREATE DATABASE default" > $CLICKHOUSE_BACKUP_DIR/schema_$TIMESTAMP.sql 2>/dev/null || true

# Export all tables
docker exec brokle-clickhouse clickhouse-client --query "SELECT name FROM system.tables WHERE database = 'default'" | while read table; do
    if [ -n "$table" ]; then
        echo "  Backing up table: $table"
        docker exec brokle-clickhouse clickhouse-client --query "SELECT * FROM default.$table FORMAT TabSeparated" | gzip > $CLICKHOUSE_BACKUP_DIR/${table}_$TIMESTAMP.tsv.gz 2>/dev/null || echo "  Warning: Failed to backup $table"
    fi
done

echo "ClickHouse backup created in: $CLICKHOUSE_BACKUP_DIR/"

# Keep only last 7 days
echo ""
echo "Cleaning old backups (keeping last 7 days)..."
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete
find $CLICKHOUSE_BACKUP_DIR -name "*_*.sql" -mtime +7 -delete
find $CLICKHOUSE_BACKUP_DIR -name "*_*.tsv.gz" -mtime +7 -delete
echo "Backup complete!"
EOFBACKUP

chmod +x backup.sh

info "✓ Helper scripts created: check-health.sh, backup.sh"

# Deployment complete
info "Deployment complete! 🎉"
