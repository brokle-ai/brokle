# Brokle Platform Release Guide

This guide explains how to release new versions of the Brokle platform (backend + web).

---

## Release Overview

The Brokle platform uses **tag-based automated releases** with GitHub Actions:

```
Developer creates tag → GitHub Actions triggers
    ↓
Runs full test suite
    ↓
Builds artifacts:
- 4 Go binaries (server/worker, OSS/Enterprise)
- 5 Docker images (multi-arch)
    ↓
Publishes to GitHub Container Registry
    ↓
Creates GitHub Release with binaries
```

---

## Version Management

### Version Storage

**Single Source of Truth**: `/VERSION` file

**Synchronized Locations**:
1. `VERSION` - Root version file (e.g., `v0.1.0`)
2. `internal/version/version.go` - Injected via ldflags during build
3. `web/src/constants/VERSION.ts` - Frontend version constant
4. `web/package.json` - Frontend package version

### Versioning Strategy

**Semantic Versioning**: `vMAJOR.MINOR.PATCH[-PRERELEASE]`

- **Major (v1.0.0)**: Breaking changes, incompatible API changes
- **Minor (v0.2.0)**: New features, backward compatible
- **Patch (v0.1.1)**: Bug fixes, backward compatible
- **Pre-release**: `-alpha.1`, `-beta.1`, `-rc.1`

---

## Release Process

### Prerequisites

Before creating a release:

- [ ] All tests passing locally: `make test`
- [ ] Frontend tests passing: `cd web && pnpm test`
- [ ] Clean git working directory
- [ ] On `main` branch and up-to-date
- [ ] CHANGELOG.md updated with release notes
- [ ] Migration guides written (if breaking changes)

---

### Quick Release (Automated - Recommended)

**Using Makefile commands** (same as SDKs):

```bash
# Preview release (dry run - always do this first!)
make release-dry

# Release patch version (0.1.0 → 0.1.1)
make release-patch

# Release minor version (0.1.0 → 0.2.0)
make release-minor

# Release major version (0.1.0 → 1.0.0)
make release-major

# Skip tests (faster, for hotfixes)
make release-patch-skip-tests
```

**What the script does automatically**:
1. ✅ Validates git state (clean, on main, up-to-date)
2. ✅ Calculates new version based on bump type
3. ✅ Updates all 3 version files (VERSION, VERSION.ts, package.json)
4. ✅ Runs tests (optional, can skip)
5. ✅ Asks for confirmation
6. ✅ Commits changes
7. ✅ Creates git tag
8. ✅ Pushes to GitHub
9. ✅ GitHub Actions automatically builds and publishes

**Total time**: ~2 minutes from command to live Docker images!

**Example**:
```bash
$ make release-patch

🚀 Brokle Platform Release
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Checking prerequisites...
✓ Working directory is clean
✓ On main branch
✓ Up-to-date with remote

📦 Current version: v0.1.0
📦 New version: v0.1.1

📝 Updating version files...
✓ Updated VERSION
✓ Updated web/src/constants/VERSION.ts
✓ Updated web/package.json

🧪 Running tests...
✓ Go tests passed
✓ Frontend tests passed

📋 Release Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Version: v0.1.0 → v0.1.1
Bump type: patch

Proceed with release? (y/N): y

🎯 Creating release...
✓ Committed version bump
✓ Created tag v0.1.1
✓ Pushed to GitHub

✅ Release v0.1.1 created successfully!

📦 Next steps:
Watch GitHub Actions: https://github.com/brokle-ai/brokle/actions
```

---

### Manual Release (Alternative)

#### Step 1: Update Version Files

```bash
# 1. Update VERSION file
echo "v0.2.0" > VERSION

# 2. Update frontend version constant
# Edit web/src/constants/VERSION.ts:
export const VERSION = "v0.2.0";

# 3. Update frontend package.json
cd web
# Edit package.json version field to "0.2.0" (no 'v' prefix)
cd ..

# 4. Verify changes
cat VERSION
cat web/src/constants/VERSION.ts
grep '"version"' web/package.json
```

---

#### Step 2: Update CHANGELOG.md

```bash
# Edit CHANGELOG.md
# Move Unreleased section to new version:

## [Unreleased]

## [0.2.0] - 2024-11-20

### Added
- New feature X
- New feature Y

### Changed
- Updated component Z

### Fixed
- Fixed bug in A
- Fixed issue with B

[Unreleased]: https://github.com/brokle-ai/brokle/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/brokle-ai/brokle/releases/tag/v0.2.0
```

---

#### Step 3: Commit Version Bump

```bash
# Commit all version changes
git add VERSION web/src/constants/VERSION.ts web/package.json CHANGELOG.md
git commit -m "chore: bump version to v0.2.0"
git push origin main
```

---

#### Step 4: Create and Push Git Tag

```bash
# Create tag
git tag v0.2.0

# Push tag (this triggers the release workflow)
git push origin v0.2.0
```

---

#### Step 5: Monitor Release Workflow

**Watch GitHub Actions**:
- Go to: https://github.com/brokle-ai/brokle/actions/workflows/release.yml
- Monitor the release workflow progress

**Workflow will**:
1. ✅ Validate version format
2. ✅ Run full test suite (Go + Frontend + Docker)
3. ✅ Build 4 Go binaries with version injection
4. ✅ Build and publish 5 Docker images (multi-arch)
5. ✅ Create GitHub Release with binaries attached

**Duration**: ~10-15 minutes

---

#### Step 6: Verify Release

**Check GitHub Release**:
- Go to: https://github.com/brokle-ai/brokle/releases
- Verify release created with:
  - Version number (v0.2.0)
  - Binary attachments (4 files + SHA256SUMS.txt)
  - Auto-generated notes
  - CHANGELOG excerpt

**Check Docker Images**:
```bash
# Pull and verify images
docker pull ghcr.io/brokle-ai/brokle-server:v0.2.0
docker pull ghcr.io/brokle-ai/brokle-server:latest
docker pull ghcr.io/brokle-ai/brokle-worker:v0.2.0
docker pull ghcr.io/brokle-ai/brokle-web:v0.2.0

# Verify version in binary
docker run --rm ghcr.io/brokle-ai/brokle-server:v0.2.0 --version
# Should output: v0.2.0
```

**Verify Docker Tags**:
- Go to: https://github.com/brokle-ai/brokle/pkgs/container/brokle-server
- Should see tags: `v0.2.0`, `v0.2`, `v0`, `latest`, `main`, `sha-abc1234`

---

## Pre-release Process

For alpha, beta, or RC releases:

```bash
# 1. Update VERSION to pre-release format
echo "v0.2.0-rc.1" > VERSION

# 2. Update other version files
# web/src/constants/VERSION.ts: "v0.2.0-rc.1"
# web/package.json: "0.2.0-rc.1"

# 3. Update CHANGELOG.md with pre-release notes
## [0.2.0-rc.1] - 2024-11-18

### Features (Release Candidate)
- [RC] New feature X ready for testing

### Known Issues
- Minor issue Y being investigated

# 4. Commit and tag
git add VERSION web/ CHANGELOG.md
git commit -m "chore: release v0.2.0-rc.1"
git tag v0.2.0-rc.1
git push origin v0.2.0-rc.1

# 5. GitHub Actions will:
# - Mark as pre-release ✅
# - NOT tag Docker images as 'latest' ✅
# - Tag as v0.2.0-rc.1 only
```

**Pre-release benefits**:
- Test in production-like environment
- Get early feedback
- Iterate before stable release
- Don't affect `latest` tag

---

## Build Artifacts

### Go Binaries (4 variants)

**Server Binaries**:
1. `brokle-server` - OSS HTTP server
2. `brokle-server-enterprise` - Enterprise HTTP server (SSO, RBAC, compliance)

**Worker Binaries**:
3. `brokle-worker` - OSS background worker
4. `brokle-worker-enterprise` - Enterprise background worker

**All binaries**:
- Platform: Linux amd64
- CGO: Disabled (static linking)
- Debug symbols: Stripped (`-w -s`)
- Version: Injected via ldflags

### Docker Images (5 images)

**OSS Images**:
1. `ghcr.io/brokle-ai/brokle-server` - HTTP API server
2. `ghcr.io/brokle-ai/brokle-worker` - Background worker
3. `ghcr.io/brokle-ai/brokle-web` - Next.js frontend

**Enterprise Images**:
4. `ghcr.io/brokle-ai/brokle-server-enterprise` - Server with EE features
5. `ghcr.io/brokle-ai/brokle-worker-enterprise` - Worker with EE analytics

**All images**:
- Platforms: linux/amd64, linux/arm64 (releases only)
- Registry: GitHub Container Registry (ghcr.io)
- Base: Alpine Linux
- Non-root user: `brokle`

---

## Docker Image Tags

### Tag Strategy

**For release v0.2.3**:
- `v0.2.3` - Exact version
- `v0.2` - Major.minor (latest 0.2.x)
- `v0` - Major version (latest 0.x.x)
- `latest` - Latest stable release (no RC/alpha/beta)
- `sha-abc1234` - Git commit SHA

**For main branch**:
- `main` - Latest main branch build
- `sha-abc1234` - Git commit SHA

**For pre-releases** (v0.2.0-rc.1):
- `v0.2.0-rc.1` - Exact version
- NO `latest` tag (pre-releases don't affect latest)

---

## Deployment

### Using Docker Images

**Production deployment**:

```bash
# Pull specific version
docker pull ghcr.io/brokle-ai/brokle-server:v0.2.0
docker pull ghcr.io/brokle-ai/brokle-worker:v0.2.0
docker pull ghcr.io/brokle-ai/brokle-web:v0.2.0

# Or use latest
docker pull ghcr.io/brokle-ai/brokle-server:latest
```

**docker-compose.yml**:

```yaml
services:
  server:
    image: ghcr.io/brokle-ai/brokle-server:v0.2.0
    # ...

  worker:
    image: ghcr.io/brokle-ai/brokle-worker:v0.2.0
    # ...

  web:
    image: ghcr.io/brokle-ai/brokle-web:v0.2.0
    # ...
```

### Using Binaries

```bash
# Download from GitHub Release
wget https://github.com/brokle-ai/brokle/releases/download/v0.2.0/brokle-server
chmod +x brokle-server

# Verify version
./brokle-server --version
# Output: v0.2.0

# Run server
./brokle-server
```

---

## CI/CD Workflows

### CI Pipeline (.github/workflows/ci.yml)

**Triggers**:
- Push to `main`
- Pull requests to `main`
- Manual dispatch

**Jobs**:
1. **test-go**: Go unit tests + linting
2. **test-frontend**: Frontend tests (ESLint, TypeScript, Vitest)
3. **build-docker**: Docker build tests
4. **all-tests-passed**: Gate for merges
5. **publish-docker-main**: Publishes to ghcr.io (main branch only)

**Docker publishing on main**:
- Tags: `main`, `sha-{commit}`
- Platform: linux/amd64 only (fast builds)
- Purpose: Continuous delivery for staging

---

### Release Workflow (.github/workflows/release.yml)

**Triggers**:
- Git tags matching `v*.*.*`
- Manual workflow dispatch

**Jobs**:
1. **validate**: Validate semver, detect pre-release
2. **test**: Run full test suite (reuses ci.yml)
3. **build-binaries**: Build 4 Go binaries (matrix)
4. **build-docker**: Build and publish 5 Docker images (multi-arch)
5. **create-release**: Create GitHub Release with artifacts

**Multi-arch builds**:
- Releases: linux/amd64 + linux/arm64
- Main branch: linux/amd64 only

---

## Rollback

If a release has issues:

### Rollback Docker Deployment

```bash
# Revert to previous version
docker pull ghcr.io/brokle-ai/brokle-server:v0.1.0
docker pull ghcr.io/brokle-ai/brokle-worker:v0.1.0
docker pull ghcr.io/brokle-ai/brokle-web:v0.1.0

# Update deployment
kubectl set image deployment/brokle-server brokle-server=ghcr.io/brokle-ai/brokle-server:v0.1.0
```

### Yank GitHub Release (Extreme Cases)

```bash
# Mark release as pre-release to remove from "latest"
# Or delete release entirely (npm style "yank")
# Go to: https://github.com/brokle-ai/brokle/releases
```

### Database Migrations

If release included migrations:

```bash
# Rollback migrations
make migrate-down

# Or use specific steps
go run cmd/migrate/main.go -db postgres down -steps 1
```

---

## Release Checklist

### Pre-Release

- [ ] All tests passing locally
- [ ] Database migrations tested (up and down)
- [ ] CHANGELOG.md updated
- [ ] VERSION file updated
- [ ] Frontend version updated
- [ ] Breaking changes documented
- [ ] Migration guide written (if needed)
- [ ] Code review completed

### Release

- [ ] Version files committed
- [ ] Git tag created and pushed
- [ ] GitHub Actions workflow succeeded
- [ ] Docker images published
- [ ] Binaries attached to release
- [ ] GitHub Release created

### Post-Release

- [ ] Docker images verified (pull and test)
- [ ] Binary downloads work
- [ ] Health endpoint shows correct version
- [ ] Staging deployment updated
- [ ] Production deployment scheduled
- [ ] Team notified (Slack/Discord)
- [ ] Documentation updated
- [ ] Monitoring dashboards checked

---

## Troubleshooting

### Build Failures

**Go binary build fails**:
```bash
# Test locally
make build-server-oss
make build-worker-oss

# Check for compilation errors
go build ./cmd/server
go build ./cmd/worker
```

**Docker build fails**:
```bash
# Test locally
docker build -t test-server -f Dockerfile .
docker build -t test-worker -f Dockerfile.worker .
docker build -t test-web -f web/Dockerfile ./web
```

### Version Injection Issues

**Verify version in binary**:
```bash
# After build
./bin/brokle-server --version

# Or check health endpoint
curl http://localhost:8080/health | jq '.version'
```

### Docker Publishing Issues

**Check GitHub Container Registry permissions**:
- Go to: https://github.com/brokle-ai/brokle/settings/actions
- Ensure: "Read and write permissions" for GITHUB_TOKEN

**Verify image exists**:
```bash
# List tags
gh api /orgs/brokle-ai/packages/container/brokle-server/versions

# Or check UI
# https://github.com/brokle-ai/brokle/pkgs/container/brokle-server
```

---

## Advanced Topics

### Manual Release (Without Tag)

```bash
# Trigger manually from GitHub UI
# Go to: https://github.com/brokle-ai/brokle/actions/workflows/release.yml
# Click "Run workflow"
# Enter version: v0.2.0
# Click "Run workflow"
```

### Building Locally

**Build all artifacts locally** (for testing):

```bash
# Set version
export VERSION=v0.2.0-test

# Build binaries
make build-server-oss
make build-server-enterprise
make build-worker-oss
make build-worker-enterprise

# Build Docker images
docker build --build-arg VERSION=$VERSION -t brokle-server:$VERSION -f Dockerfile .
docker build --build-arg VERSION=$VERSION -t brokle-worker:$VERSION -f Dockerfile.worker .
docker build --build-arg VERSION=$VERSION -t brokle-web:$VERSION -f web/Dockerfile ./web
```

### Release Automation Script (Future)

Create `scripts/release.sh` (like SDKs):

```bash
#!/bin/bash
# Future enhancement: Automate version file updates
# make release-patch
# make release-minor
# make release-major
```

---

## Comparison with SDKs

| Feature | Platform | Python SDK | JavaScript SDK |
|---------|----------|------------|----------------|
| **Release Command** | Manual tag | `make release-patch` | `make release-patch` |
| **Tag Format** | `v0.1.0` | `v0.2.10` | `v0.1.3` |
| **Automation** | GitHub Actions | GitHub Actions | GitHub Actions |
| **Artifacts** | Binaries + Docker | PyPI package | npm packages (4) |
| **Publishing** | ghcr.io | PyPI (Trusted) | npm (NPM_TOKEN) |

**Platform is similar but more complex** (binaries + Docker images)

---

## Future Enhancements

### Phase 1 (Current)
- [x] Version management system
- [x] Automated CI/CD
- [x] Docker publishing
- [x] Binary artifacts
- [x] GitHub Releases

### Phase 2 (Future)
- [ ] Automated version bumping script (`scripts/release.sh`)
- [ ] Automated changelog generation (git-cliff)
- [ ] Docker Hub publishing (in addition to ghcr.io)
- [ ] Slack/Discord release notifications
- [ ] Deploy automation to staging/production

### Phase 3 (Later)
- [ ] Release notes templates
- [ ] Automated security scanning
- [ ] Performance benchmarking before release
- [ ] Canary deployments
- [ ] Rollback automation

---

## References

- **GitHub Actions**: https://docs.github.com/en/actions
- **Docker Buildx**: https://docs.docker.com/build/
- **GitHub Container Registry**: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
- **Semantic Versioning**: https://semver.org/

---

## Quick Reference

```bash
# Check current version
cat VERSION

# Update version
echo "v0.3.0" > VERSION

# Release workflow
git add VERSION web/
git commit -m "chore: bump version to v0.3.0"
git push origin main
git tag v0.3.0
git push origin v0.3.0

# Monitor
https://github.com/brokle-ai/brokle/actions

# Verify
docker pull ghcr.io/brokle-ai/brokle-server:v0.3.0
curl -s https://github.com/brokle-ai/brokle/releases | grep v0.3.0
```

---

**For questions or issues, see the main [RELEASE_STRATEGY.md](./RELEASE_STRATEGY.md) document.**
