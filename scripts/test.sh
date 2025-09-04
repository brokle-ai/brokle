#!/bin/bash
set -e

# Brokle Testing Script
# Comprehensive testing for the Brokle platform

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

# Test configuration
TEST_ENV=${TEST_ENV:-"test"}
COVERAGE_DIR="coverage"
TEST_RESULTS_DIR="test-results"

# Create directories
mkdir -p $COVERAGE_DIR
mkdir -p $TEST_RESULTS_DIR

# Default test type
TEST_TYPE=${1:-"all"}

run_go_tests() {
    print_step "Running Go tests..."
    
    # Unit tests
    print_step "Running Go unit tests..."
    go test -v -race -coverprofile=$COVERAGE_DIR/go-coverage.out ./... | tee $TEST_RESULTS_DIR/go-unit-tests.log
    
    # Generate coverage report
    go tool cover -html=$COVERAGE_DIR/go-coverage.out -o $COVERAGE_DIR/go-coverage.html
    
    # Coverage summary
    COVERAGE=$(go tool cover -func=$COVERAGE_DIR/go-coverage.out | grep total | awk '{print $3}')
    print_success "Go tests completed - Coverage: $COVERAGE"
}

run_frontend_tests() {
    print_step "Running frontend tests..."
    
    cd web
    
    # Unit tests with coverage
    print_step "Running React/TypeScript unit tests..."
    npm test -- --coverage --watchAll=false --testResultsProcessor=jest-junit
    
    # Copy coverage results
    cp -r coverage ../../$COVERAGE_DIR/frontend-coverage
    cp junit.xml ../../$TEST_RESULTS_DIR/frontend-tests.xml
    
    cd ../..
    
    print_success "Frontend tests completed"
}

run_integration_tests() {
    print_step "Running integration tests..."
    
    # Start test environment
    print_step "Starting test services..."
    docker-compose -f docker-compose.test.yml up -d --build
    
    # Wait for services to be ready
    sleep 10
    
    # Run integration tests
    print_step "Running API integration tests..."
    go test -v -tags=integration ./tests/integration/... | tee $TEST_RESULTS_DIR/integration-tests.log
    
    # Cleanup test environment
    docker-compose -f docker-compose.test.yml down -v
    
    print_success "Integration tests completed"
}

run_e2e_tests() {
    print_step "Running end-to-end tests..."
    
    # Start full environment
    docker-compose up -d
    sleep 15
    
    # Run E2E tests
    cd web
    
    # Check if Playwright is available
    if npm list @playwright/test &> /dev/null; then
        print_step "Running Playwright E2E tests..."
        npm run test:e2e
        
        # Copy results
        cp -r test-results ../../$TEST_RESULTS_DIR/e2e-results
        cp playwright-report ../../$TEST_RESULTS_DIR/playwright-report
    else
        print_warning "Playwright not found, skipping E2E tests"
        print_warning "Install with: npm install @playwright/test"
    fi
    
    cd ../..
    
    # Cleanup
    docker-compose down
    
    print_success "End-to-end tests completed"
}

run_load_tests() {
    print_step "Running load tests..."
    
    # Start services
    docker-compose up -d
    sleep 15
    
    # Check if k6 is available
    if command -v k6 &> /dev/null; then
        print_step "Running k6 load tests..."
        
        # Run basic load test
        k6 run tests/load/basic-load.js
        
        # Run stress test
        k6 run tests/load/stress-test.js
        
        print_success "Load tests completed"
    else
        print_warning "k6 not found, skipping load tests"
        print_warning "Install k6: https://k6.io/docs/getting-started/installation/"
    fi
    
    # Cleanup
    docker-compose down
}

run_security_tests() {
    print_step "Running security tests..."
    
    # Go security scan
    if command -v gosec &> /dev/null; then
        print_step "Running gosec security scan..."
        gosec -fmt json -out $TEST_RESULTS_DIR/gosec-report.json ./...
        gosec ./...
        print_success "Go security scan completed"
    else
        print_warning "gosec not found, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    fi
    
    # Frontend security scan
    cd web
    
    if npm list npm-audit &> /dev/null; then
        print_step "Running npm audit..."
        npm audit --audit-level=moderate | tee ../../$TEST_RESULTS_DIR/npm-audit.log
        print_success "npm audit completed"
    fi
    
    cd ../..
    
    # Docker security scan with Trivy
    if command -v trivy &> /dev/null; then
        print_step "Running Docker security scan..."
        
        # Build images first
        docker-compose build
        
        # Scan main API image
        trivy image --format json --output $TEST_RESULTS_DIR/trivy-api.json brokle-api:latest
        trivy image brokle-api:latest
        
        print_success "Docker security scan completed"
    else
        print_warning "Trivy not found for Docker security scanning"
    fi
}

run_lint_tests() {
    print_step "Running linting and code quality checks..."
    
    # Go linting
    if command -v golangci-lint &> /dev/null; then
        print_step "Running golangci-lint..."
        golangci-lint run --out-format json --out-format colored-line-number > $TEST_RESULTS_DIR/golangci-lint.json
        golangci-lint run
        print_success "Go linting completed"
    else
        print_warning "golangci-lint not found, running basic checks..."
        go vet ./...
        go fmt -l . | tee $TEST_RESULTS_DIR/go-fmt-issues.log
    fi
    
    # Frontend linting
    cd web
    
    print_step "Running ESLint and TypeScript checks..."
    npm run lint | tee ../../$TEST_RESULTS_DIR/eslint-results.log
    npm run type-check | tee ../../$TEST_RESULTS_DIR/typescript-check.log
    
    cd ../..
    
    print_success "Linting completed"
}

generate_test_report() {
    print_step "Generating test report..."
    
    REPORT_FILE="$TEST_RESULTS_DIR/test-report.md"
    
    cat > $REPORT_FILE << EOF
# Brokle Platform Test Report

Generated: $(date)

## Test Results Summary

### Go Tests
$(if [ -f "$COVERAGE_DIR/go-coverage.out" ]; then
    COVERAGE=$(go tool cover -func=$COVERAGE_DIR/go-coverage.out | grep total | awk '{print $3}')
    echo "- Coverage: $COVERAGE"
    echo "- Report: [Go Coverage Report]($COVERAGE_DIR/go-coverage.html)"
fi)

### Frontend Tests  
$(if [ -d "$COVERAGE_DIR/frontend-coverage" ]; then
    echo "- Coverage Report: [Frontend Coverage]($COVERAGE_DIR/frontend-coverage/lcov-report/index.html)"
fi)

### Integration Tests
$(if [ -f "$TEST_RESULTS_DIR/integration-tests.log" ]; then
    echo "- Results: [Integration Test Log]($TEST_RESULTS_DIR/integration-tests.log)"
fi)

### Security Scans
$(if [ -f "$TEST_RESULTS_DIR/gosec-report.json" ]; then
    echo "- Go Security: [Gosec Report]($TEST_RESULTS_DIR/gosec-report.json)"
fi)
$(if [ -f "$TEST_RESULTS_DIR/npm-audit.log" ]; then
    echo "- npm Audit: [npm Audit Log]($TEST_RESULTS_DIR/npm-audit.log)"
fi)

### Code Quality
$(if [ -f "$TEST_RESULTS_DIR/golangci-lint.json" ]; then
    echo "- Go Linting: [golangci-lint Report]($TEST_RESULTS_DIR/golangci-lint.json)"
fi)
$(if [ -f "$TEST_RESULTS_DIR/eslint-results.log" ]; then
    echo "- Frontend Linting: [ESLint Results]($TEST_RESULTS_DIR/eslint-results.log)"
fi)

## Files Generated
- Test Results: \`$TEST_RESULTS_DIR/\`
- Coverage Reports: \`$COVERAGE_DIR/\`

EOF
    
    print_success "Test report generated: $REPORT_FILE"
}

# Main test execution
case $TEST_TYPE in
    "unit")
        run_go_tests
        run_frontend_tests
        ;;
    "integration")
        run_integration_tests
        ;;
    "e2e")
        run_e2e_tests
        ;;
    "load")
        run_load_tests
        ;;
    "security")
        run_security_tests
        ;;
    "lint")
        run_lint_tests
        ;;
    "all")
        print_step "Running complete test suite..."
        run_lint_tests
        run_go_tests
        run_frontend_tests
        run_security_tests
        run_integration_tests
        # Skip load and e2e by default for CI
        if [ "${CI}" != "true" ]; then
            run_e2e_tests
        fi
        generate_test_report
        ;;
    "ci")
        print_step "Running CI test suite..."
        run_lint_tests
        run_go_tests
        run_frontend_tests
        run_security_tests
        run_integration_tests
        generate_test_report
        ;;
    "help"|*)
        echo "Brokle Testing Script"
        echo ""
        echo "Usage: $0 [test-type]"
        echo ""
        echo "Test Types:"
        echo "  unit          Run unit tests (Go + Frontend)"
        echo "  integration   Run integration tests"
        echo "  e2e           Run end-to-end tests"
        echo "  load          Run load tests"
        echo "  security      Run security scans"
        echo "  lint          Run linting and code quality"
        echo "  all           Run all tests (default)"
        echo "  ci            Run CI test suite (no e2e/load)"
        echo "  help          Show this help"
        echo ""
        echo "Examples:"
        echo "  $0 unit       # Run only unit tests"
        echo "  $0 security   # Run security scans"
        echo "  $0 all        # Run complete test suite"
        echo ""
        echo "Environment Variables:"
        echo "  TEST_ENV      Test environment (default: test)"
        echo "  CI            CI mode flag (skips long tests)"
        echo ""
        exit 0
        ;;
esac

echo ""
print_success "Testing completed! Results in: $TEST_RESULTS_DIR/"
print_success "Coverage reports in: $COVERAGE_DIR/"