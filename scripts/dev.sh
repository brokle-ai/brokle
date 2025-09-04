#!/bin/bash
set -e

# Brokle Development Script
# Quick development environment management

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Default command
COMMAND=${1:-"start"}

case $COMMAND in
    "start")
        print_step "Starting Brokle development environment..."
        
        # Start infrastructure first
        docker-compose up -d postgres redis clickhouse
        sleep 3
        
        # Run migrations if needed
        make migrate-up
        
        # Start all services
        docker-compose up -d
        
        # Start frontend in development mode
        print_step "Starting frontend development server..."
        cd web
        npm run dev &
        FRONTEND_PID=$!
        cd ../..
        
        print_success "Development environment started!"
        echo ""
        echo "🔗 Services running:"
        echo "  • Frontend: http://localhost:3000 (Next.js dev server)"
        echo "  • Backend: http://localhost:8080 (Go API)"
        echo "  • Database: localhost:5432 (PostgreSQL)"
        echo "  • Analytics: localhost:8123 (ClickHouse)"
        echo "  • Cache: localhost:6379 (Redis)"
        echo ""
        echo "Press Ctrl+C to stop all services"
        
        # Wait for interrupt
        trap "print_step 'Stopping services...'; kill $FRONTEND_PID 2>/dev/null || true; docker-compose down; print_success 'Stopped'" INT
        wait
        ;;
        
    "stop")
        print_step "Stopping Brokle development environment..."
        
        # Kill any running npm processes
        pkill -f "npm run dev" || true
        pkill -f "next dev" || true
        
        # Stop Docker services
        docker-compose down
        
        print_success "Development environment stopped"
        ;;
        
    "restart")
        print_step "Restarting Brokle development environment..."
        $0 stop
        sleep 2
        $0 start
        ;;
        
    "reset")
        print_warning "This will destroy all data and restart fresh. Continue? (y/N)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            print_step "Resetting development environment..."
            
            # Stop everything
            $0 stop
            
            # Remove volumes
            docker-compose down -v
            
            # Rebuild and start
            docker-compose build
            $0 start
            
            print_success "Development environment reset complete"
        else
            print_warning "Reset cancelled"
        fi
        ;;
        
    "logs")
        SERVICE=${2:-""}
        if [ -n "$SERVICE" ]; then
            docker-compose logs -f $SERVICE
        else
            docker-compose logs -f
        fi
        ;;
        
    "shell")
        SERVICE=${2:-"api"}
        docker-compose exec $SERVICE /bin/sh
        ;;
        
    "db")
        SUBCOMMAND=${2:-"psql"}
        case $SUBCOMMAND in
            "psql")
                docker-compose exec postgres psql -U brokle -d brokle_dev
                ;;
            "migrate")
                make migrate-up
                ;;
            "seed")
                make seed-dev
                ;;
            "reset")
                print_warning "This will destroy all database data. Continue? (y/N)"
                read -r response
                if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
                    make db-reset
                    make migrate-up
                    make seed-dev
                    print_success "Database reset complete"
                else
                    print_warning "Database reset cancelled"
                fi
                ;;
            *)
                echo "Usage: $0 db [psql|migrate|seed|reset]"
                ;;
        esac
        ;;
        
    "test")
        print_step "Running tests..."
        
        # Run Go tests
        go test ./...
        
        # Run frontend tests
        cd web
        npm test
        cd ../..
        
        print_success "All tests completed"
        ;;
        
    "lint")
        print_step "Running linters..."
        
        # Go linting
        if command -v golangci-lint &> /dev/null; then
            golangci-lint run
        else
            go vet ./...
            go fmt ./...
        fi
        
        # Frontend linting
        cd web
        npm run lint
        cd ../..
        
        print_success "Linting completed"
        ;;
        
    "build")
        print_step "Building all services..."
        
        # Build backend
        go build -o bin/brokle ./cmd/server
        
        # Build frontend
        cd web
        npm run build
        cd ../..
        
        # Build Docker images
        docker-compose build
        
        print_success "Build completed"
        ;;
        
    "clean")
        print_step "Cleaning up..."
        
        # Stop services
        $0 stop
        
        # Remove containers and networks
        docker-compose down --remove-orphans
        
        # Clean Go cache
        go clean -cache
        
        # Clean npm cache
        cd web
        npm cache clean --force
        cd ../..
        
        # Remove binaries
        rm -rf bin/
        
        print_success "Cleanup completed"
        ;;
        
    "status")
        print_step "Checking service status..."
        docker-compose ps
        echo ""
        make health
        ;;
        
    "help"|*)
        echo "Brokle Development Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  start         Start development environment"
        echo "  stop          Stop development environment"
        echo "  restart       Restart development environment"
        echo "  reset         Reset environment (destroys data)"
        echo "  logs [svc]    View logs (optionally for specific service)"
        echo "  shell [svc]   Access service shell (default: api)"
        echo "  db [cmd]      Database operations (psql|migrate|seed|reset)"
        echo "  test          Run all tests"
        echo "  lint          Run linters"
        echo "  build         Build all services"
        echo "  clean         Clean up everything"
        echo "  status        Check service status"
        echo "  help          Show this help"
        echo ""
        echo "Examples:"
        echo "  $0 start              # Start dev environment"
        echo "  $0 logs api           # View API logs"
        echo "  $0 shell postgres     # Access PostgreSQL shell"
        echo "  $0 db psql            # Connect to database"
        echo "  $0 db reset           # Reset database"
        echo ""
        ;;
esac