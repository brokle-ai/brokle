#!/bin/bash

# Brokle Platform Release Script
# Automates version bumping, changelog updates, and git operations
# Usage: ./scripts/release.sh [patch|minor|major] [--skip-tests] [--dry-run]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VERSION_FILE="VERSION"
FRONTEND_VERSION_TS="web/src/constants/VERSION.ts"
FRONTEND_PACKAGE_JSON="web/package.json"

# Flags
SKIP_TESTS=false
DRY_RUN=false

# Parse arguments
BUMP_TYPE=$1
shift || true

while [[ $# -gt 0 ]]; do
  case $1 in
    --skip-tests)
      SKIP_TESTS=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate bump type
if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo -e "${RED}❌ Invalid bump type: $BUMP_TYPE${NC}"
  echo "Usage: $0 [patch|minor|major] [--skip-tests] [--dry-run]"
  exit 1
fi

echo -e "${BLUE}🚀 Brokle Platform Release${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Function: Check prerequisites
check_prerequisites() {
  echo -e "${BLUE}📋 Checking prerequisites...${NC}"

  # Check if git working directory is clean
  if [[ -n $(git status --porcelain) ]]; then
    echo -e "${RED}❌ Working directory is not clean${NC}"
    echo "Please commit or stash your changes before releasing"
    git status --short
    exit 1
  fi
  echo -e "${GREEN}✓${NC} Working directory is clean"

  # Check if on main branch
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo -e "${RED}❌ Not on main branch (current: $CURRENT_BRANCH)${NC}"
    echo "Please switch to main branch: git checkout main"
    exit 1
  fi
  echo -e "${GREEN}✓${NC} On main branch"

  # Check if up-to-date with remote
  git fetch origin main
  LOCAL=$(git rev-parse main)
  REMOTE=$(git rev-parse origin/main)

  if [[ "$LOCAL" != "$REMOTE" ]]; then
    echo -e "${RED}❌ Local main branch is not up-to-date with origin/main${NC}"
    echo "Please pull latest changes: git pull origin main"
    exit 1
  fi
  echo -e "${GREEN}✓${NC} Up-to-date with remote"

  echo ""
}

# Function: Get current version
get_current_version() {
  if [[ ! -f "$VERSION_FILE" ]]; then
    echo -e "${RED}❌ VERSION file not found${NC}"
    exit 1
  fi

  CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '\n')
  echo -e "${BLUE}📦 Current version:${NC} ${CURRENT_VERSION}"
}

# Function: Calculate new version
calculate_new_version() {
  # Remove 'v' prefix for calculation
  VERSION_NUM="${CURRENT_VERSION#v}"

  # Parse semantic version
  IFS='.' read -r -a parts <<< "$VERSION_NUM"
  MAJOR="${parts[0]}"
  MINOR="${parts[1]}"
  PATCH="${parts[2]}"

  # Calculate new version based on bump type
  case $BUMP_TYPE in
    patch)
      PATCH=$((PATCH + 1))
      ;;
    minor)
      MINOR=$((MINOR + 1))
      PATCH=0
      ;;
    major)
      MAJOR=$((MAJOR + 1))
      MINOR=0
      PATCH=0
      ;;
  esac

  NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
  echo -e "${BLUE}📦 New version:${NC} ${GREEN}${NEW_VERSION}${NC}"
  echo ""
}

# Function: Update version files
update_version_files() {
  echo -e "${BLUE}📝 Updating version files...${NC}"

  # Update VERSION file
  echo "$NEW_VERSION" > "$VERSION_FILE"
  echo -e "${GREEN}✓${NC} Updated $VERSION_FILE"

  # Update frontend VERSION.ts
  cat > "$FRONTEND_VERSION_TS" << EOF
/**
 * Brokle Platform Version
 *
 * This version is automatically updated during the release process.
 * Displayed in the UI footer and about page.
 */
export const VERSION = "$NEW_VERSION";
EOF
  echo -e "${GREEN}✓${NC} Updated $FRONTEND_VERSION_TS"

  # Update web/package.json (remove 'v' prefix for npm)
  NPM_VERSION="${NEW_VERSION#v}"
  if command -v jq &> /dev/null; then
    # Use jq if available
    jq ".version = \"$NPM_VERSION\"" "$FRONTEND_PACKAGE_JSON" > "$FRONTEND_PACKAGE_JSON.tmp"
    mv "$FRONTEND_PACKAGE_JSON.tmp" "$FRONTEND_PACKAGE_JSON"
  else
    # Fallback to sed
    sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$NPM_VERSION\"/" "$FRONTEND_PACKAGE_JSON"
    rm -f "$FRONTEND_PACKAGE_JSON.bak"
  fi
  echo -e "${GREEN}✓${NC} Updated $FRONTEND_PACKAGE_JSON"

  echo ""
}

# Function: Run tests
run_tests() {
  if [[ "$SKIP_TESTS" == true ]]; then
    echo -e "${YELLOW}⚠️  Skipping tests (--skip-tests flag)${NC}"
    echo ""
    return
  fi

  echo -e "${BLUE}🧪 Running tests...${NC}"

  # Run Go tests
  echo "Running Go tests..."
  if make test > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go tests passed"
  else
    echo -e "${RED}❌ Go tests failed${NC}"
    echo "Run 'make test' to see errors"
    exit 1
  fi

  # Run frontend tests
  echo "Running frontend tests..."
  if (cd web && pnpm test > /dev/null 2>&1); then
    echo -e "${GREEN}✓${NC} Frontend tests passed"
  else
    echo -e "${RED}❌ Frontend tests failed${NC}"
    echo "Run 'cd web && pnpm test' to see errors"
    exit 1
  fi

  echo ""
}

# Function: Confirm release
confirm_release() {
  echo -e "${YELLOW}📋 Release Summary${NC}"
  echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "Version: ${CURRENT_VERSION} → ${GREEN}${NEW_VERSION}${NC}"
  echo -e "Bump type: ${BUMP_TYPE}"
  echo ""
  echo "Files to be updated:"
  echo "  • VERSION"
  echo "  • web/src/constants/VERSION.ts"
  echo "  • web/package.json"
  echo ""
  echo "Git operations:"
  echo "  • Commit: chore: bump version to $NEW_VERSION"
  echo "  • Tag: $NEW_VERSION"
  echo "  • Push: origin main + tags"
  echo ""
  echo "After push, GitHub Actions will:"
  echo "  • Build 4 Go binaries (server/worker, OSS/Enterprise)"
  echo "  • Build 3 Docker images (multi-arch)"
  echo "  • Publish to ghcr.io"
  echo "  • Create GitHub Release"
  echo ""

  if [[ "$DRY_RUN" == true ]]; then
    echo -e "${YELLOW}🏁 DRY RUN - No changes made${NC}"
    exit 0
  fi

  read -p "Proceed with release? (y/N): " -n 1 -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Release cancelled${NC}"
    exit 0
  fi
  echo ""
}

# Function: Create release
create_release() {
  echo -e "${BLUE}🎯 Creating release...${NC}"

  # Commit changes
  git add "$VERSION_FILE" "$FRONTEND_VERSION_TS" "$FRONTEND_PACKAGE_JSON"
  git commit -m "chore: bump version to $NEW_VERSION"
  echo -e "${GREEN}✓${NC} Committed version bump"

  # Create tag
  git tag "$NEW_VERSION"
  echo -e "${GREEN}✓${NC} Created tag $NEW_VERSION"

  # Push to remote
  git push origin main --tags
  echo -e "${GREEN}✓${NC} Pushed to GitHub"

  echo ""
}

# Function: Print success message
print_success() {
  echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${GREEN}✅ Release $NEW_VERSION created successfully!${NC}"
  echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo ""
  echo -e "${BLUE}📦 Next steps:${NC}"
  echo ""
  echo "1. Watch GitHub Actions build and publish:"
  echo -e "   ${BLUE}https://github.com/brokle-ai/brokle/actions${NC}"
  echo ""
  echo "2. Monitor release workflow:"
  echo -e "   ${BLUE}https://github.com/brokle-ai/brokle/actions/workflows/release.yml${NC}"
  echo ""
  echo "3. After workflow completes (~10-15 min), verify:"
  echo "   • GitHub Release: https://github.com/brokle-ai/brokle/releases/tag/$NEW_VERSION"
  echo "   • Docker images: docker pull ghcr.io/brokle-ai/brokle-server:$NEW_VERSION"
  echo "   • Binaries: Download from GitHub Release"
  echo ""
  echo -e "${YELLOW}⚠️  Don't forget to update CHANGELOG.md if you haven't already!${NC}"
  echo ""
}

# Main execution
main() {
  check_prerequisites
  get_current_version
  calculate_new_version
  update_version_files
  run_tests
  confirm_release
  create_release
  print_success
}

main
