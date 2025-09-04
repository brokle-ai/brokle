#!/bin/bash
set -e

# Brokle Platform Setup Script
# This script sets up the complete development environment

echo "🚀 Setting up Brokle Platform..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if running in correct directory
if [ ! -f "go.mod" ]; then
    print_error "Please run this script from the root of the Brokle project"
    exit 1
fi

# Check dependencies
print_step "Checking dependencies..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker Desktop."
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

# Check Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21+."
    exit 1
fi

# Check Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18+."
    exit 1
fi

print_success "All dependencies are installed"

# Create environment files
print_step "Creating environment files..."

if [ ! -f ".env" ]; then
    cp .env.example .env
    print_success "Created .env file from template"
else
    print_warning ".env file already exists"
fi

if [ ! -f "web/.env.local" ]; then
    cp web/.env.example web/.env.local
    print_success "Created frontend .env.local file"
else
    print_warning "Frontend .env.local file already exists"
fi

# Install Go dependencies
print_step "Installing Go dependencies..."
go mod download
go mod tidy
print_success "Go dependencies installed"

# Install Node.js dependencies
print_step "Installing Node.js dependencies..."
cd web
npm install
cd ../..
print_success "Node.js dependencies installed"

# Build Docker images
print_step "Building Docker images..."
docker-compose build
print_success "Docker images built"

# Start infrastructure services
print_step "Starting infrastructure services..."
docker-compose up -d postgres redis clickhouse
sleep 5
print_success "Infrastructure services started"

# Run database migrations
print_step "Running database migrations..."
make migrate-up
print_success "Database migrations completed"

# Seed development data
print_step "Seeding development data..."
make seed-dev
print_success "Development data seeded"

# Start all services
print_step "Starting all services..."
docker-compose up -d
print_success "All services started"

# Wait for services to be healthy
print_step "Waiting for services to be ready..."
sleep 10

# Health check
print_step "Performing health checks..."
make health

echo ""
echo -e "${GREEN}🎉 Brokle Platform setup completed successfully!${NC}"
echo ""
echo "🔗 Access the platform:"
echo "  • Dashboard: http://localhost:3000"
echo "  • API Gateway: http://localhost:8080"
echo "  • Swagger UI: http://localhost:8080/swagger/index.html"
echo "  • Grafana: http://localhost:3001 (admin/admin)"
echo "  • Prometheus: http://localhost:9090"
echo ""
echo "📚 Next steps:"
echo "  • Read docs/DEVELOPMENT.md for development guidelines"
echo "  • Run 'make logs' to view service logs"
echo "  • Run 'make test' to run tests"
echo "  • Run 'make stop' to stop all services"
echo ""
echo "🛠  Development commands:"
echo "  • make start    - Start all services"
echo "  • make stop     - Stop all services"
echo "  • make restart  - Restart all services"
echo "  • make logs     - View logs"
echo "  • make shell    - Access service shell"
echo "  • make test     - Run tests"
echo ""