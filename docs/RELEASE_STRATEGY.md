# Brokle Release Strategy & Automation Plan

## Executive Summary

This document outlines a comprehensive release strategy for Brokle's platform (backend + web) and SDKs (Python + JavaScript). The strategy is based on industry best practices from OpenTelemetry, Sentry, PostHog, and other observability platforms, with automation-first approach for consistency and developer productivity.

## Current State Analysis

### Repository Architecture: Platform + SDK Submodules ⭐

Brokle uses a **multi-repository architecture with git submodules**:

```
github.com/brokle-ai/brokle/              # Main platform repository
├── internal/                              # Go backend
├── web/                                   # Next.js frontend
├── sdk/
│   ├── python/                            # ⚡ Git submodule → brokle-python.git
│   └── javascript/                        # ⚡ Git submodule → brokle-js.git
```

**Submodule Configuration** (`.gitmodules`):
```ini
[submodule "sdk/python"]
    path = sdk/python
    url = https://github.com/brokle-ai/brokle-python.git

[submodule "sdk/javascript"]
    path = sdk/javascript
    url = https://github.com/brokle-ai/brokle-js.git
```

**Key Insight**: SDKs are **completely independent repositories** that:
- ✅ Have their own CI/CD workflows and release automation
- ✅ Maintain independent versioning and release cycles
- ✅ Can be developed/released without platform repo access
- ✅ Allow platform to pin to specific SDK versions via submodule commits
- ✅ Enable SDK contributors to work in isolation
- ✅ Can be used standalone (not tied to platform monorepo)

---

### Repository 1: Platform (Backend + Web)
**Repository**: `github.com/brokle-ai/brokle`
**Location**: `/Users/Hashir/Projects/Brokle-Project/brokle/`

- **Structure**: Monorepo with Go backend and Next.js frontend
- **Current State**: ❌ No automated release workflow
- **Version**: Tracked in git tags (no current version tags)
- **Release Targets**: Docker images, Go binaries (OSS + Enterprise)
- **CI/CD**: Basic test workflow for web frontend only

---

### Repository 2: Python SDK
**Repository**: `github.com/brokle-ai/brokle-python` ⭐ (separate repo)
**Location as Submodule**: `/Users/Hashir/Projects/Brokle-Project/brokle/sdk/python/`

- **Current Version**: 0.2.9 (in `brokle/version.py`)
- **Build System**: setuptools with pyproject.toml
- **Package Manager**: pip/PyPI
- **Status**: Development Status :: 4 - Beta
- **Python Support**: 3.9, 3.10, 3.11, 3.12

**CI/CD Status**:
- ✅ **CI Pipeline** (`ci.yml`): Test matrix (Py 3.9-3.12), lint, type-check, security scan, build
- ✅ **Publish Workflow** (`publish.yml`): GitHub Release trigger → PyPI/TestPyPI publish
- ✅ **Integration Tests** (`integration-test.yml`): PostgreSQL + Redis integration
- ✅ **Release Script** (`scripts/release.py`): Automated version bumping with `make release-patch/minor/major`
- 🟡 **Needs Improvement**:
  - Script has path bug (`_version.py` vs `version.py`)
  - No Trusted Publishing (uses API tokens)
  - No pre-release auto-detection

---

### Repository 3: JavaScript SDK
**Repository**: `github.com/brokle-ai/brokle-js` ⭐ (separate repo)
**Location as Submodule**: `/Users/Hashir/Projects/Brokle-Project/brokle/sdk/javascript/`

- **Current Version**: 0.1.0 (all packages)
- **Structure**: pnpm monorepo with 4 packages:
  - `brokle` (core SDK)
  - `brokle-openai` (OpenAI wrapper)
  - `brokle-anthropic` (Anthropic wrapper)
  - `brokle-langchain` (LangChain integration)
- **Build System**: tsup (TypeScript → dual ESM/CJS)
- **Package Manager**: pnpm + npm publishing
- **Node Support**: >=18.0.0

**CI/CD Status**:
- ❌ **No GitHub Actions workflows exist**
- ❌ **No CI pipeline** (lint, test, typecheck)
- ❌ **No release automation**
- ❌ **No version management tool** (no Changesets, Lerna, or release-it)
- ❌ **Never published to npm**
- 🟡 **Has**: Test setup (Vitest), build scripts, monorepo structure

---

### Submodule Coordination
**Current State**:
- ❌ **No automation** for updating submodules when SDKs release
- ❌ **Manual process** required (`git submodule update --remote`)
- ❌ **No CI checks** to validate submodule references
- ❌ **No version compatibility matrix**
- 🟡 **Recent Activity**: 2 manual submodule commits in platform repo history

---

## Recommended Release Strategy

### 1. Versioning Strategy

#### Semantic Versioning (SemVer)
All Brokle components follow **Semantic Versioning 2.0.0**:

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

- **MAJOR**: Breaking changes (e.g., API redesign, removal of deprecated features)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, security patches
- **PRERELEASE**: alpha, beta, rc (e.g., `1.2.0-beta.1`)
- **BUILD**: Metadata (e.g., `1.2.0+20231115.sha.abc123`)

#### Version Prefixes
- **Platform**: `v3.2.1` (tags)
- **Python SDK**: `python-v1.2.3` (tags), `1.2.3` (PyPI)
- **JavaScript SDK**: `js-v0.5.0` (tags), `0.5.0` (npm)

#### Independent Versioning
Each component maintains independent versions:
- Platform can be at v3.x while Python SDK is at v2.x
- JavaScript packages in monorepo share the same version (synchronized releases)

---

### 2. Release Workflow Architecture

#### Multi-Repository Release Flow

**Important**: Each repository (Platform, Python SDK, JavaScript SDK) operates **independently** with its own release workflow:

```
┌───────────────────────────────────────────────────────────────────────┐
│  REPOSITORY 1: Platform (github.com/brokle-ai/brokle)               │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                                                       │
│  Developer → PR → CI → Merge → Tag v3.2.1 → CI/CD Workflow          │
│                                       │                               │
│                                       ├─ Build Go binaries           │
│                                       ├─ Build Next.js frontend      │
│                                       ├─ Docker multi-arch images    │
│                                       ├─ Push to ghcr.io             │
│                                       └─ GitHub Release              │
└───────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────┐
│  REPOSITORY 2: Python SDK (github.com/brokle-ai/brokle-python)      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                                                       │
│  Option A: Automated Release Script                                  │
│  Developer → make release-patch → Tests → Commit → Tag → Push       │
│                                             │                         │
│                                             └─ Triggers publish.yml  │
│                                                                       │
│  Option B: Manual GitHub Release                                     │
│  Developer → Create GitHub Release → publish.yml triggered           │
│                                         │                             │
│                                         ├─ Build package              │
│                                         ├─ Publish to PyPI            │
│                                         └─ Upload artifacts           │
└───────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────┐
│  REPOSITORY 3: JavaScript SDK (github.com/brokle-ai/brokle-js)      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                                                       │
│  Developer → Make changes → pnpm changeset → PR → Merge              │
│                                                     │                 │
│                                    Changesets creates "Version PR"   │
│                                                     │                 │
│  Maintainer → Review Version PR → Merge → Publish workflow          │
│                                               │                       │
│                                               ├─ Publish to npm      │
│                                               ├─ Create git tags     │
│                                               └─ GitHub Release      │
└───────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────┐
│  SUBMODULE COORDINATION (Platform repo triggers)                     │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                                                       │
│  SDK Release (Python or JS) → repository_dispatch webhook            │
│                                       │                               │
│                    Platform repo: Update submodule workflow          │
│                                       │                               │
│                                       ├─ git submodule update         │
│                                       ├─ Create PR in platform repo  │
│                                       ├─ Run integration tests        │
│                                       └─ Auto-merge if tests pass    │
└───────────────────────────────────────────────────────────────────────┘
```

#### Key Principles

1. **Independent Development**: SDKs can release without platform changes
2. **Platform Independence**: Platform can release without SDK updates
3. **Version Pinning**: Platform controls SDK versions via submodule commits
4. **Automated Coordination**: SDK releases can trigger platform submodule updates
5. **Testing Isolation**: Each repo tests in its own CI before release

---

### 3. Platform Release Workflow (Backend + Web)

#### Release Trigger
**Option A: Tag-based**
```yaml
on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'  # v3.2.1
      - 'v[0-9]+.[0-9]+.[0-9]+-*'  # v3.2.1-beta.1
```

**Option B: Manual Release**
```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Release version (e.g., v3.2.1)'
        required: true
```

#### Workflow Steps
1. **Validate Version**
   - Check tag format matches semver
   - Ensure version > previous version
   - Verify CHANGELOG.md updated

2. **Run Full CI Suite**
   - Go tests (backend)
   - Frontend tests (web)
   - Integration tests
   - E2E tests
   - Security scans

3. **Build Artifacts**
   - `make build-server-oss` → `bin/brokle-server`
   - `make build-worker-oss` → `bin/brokle-worker`
   - `make build-server-enterprise` → `bin/brokle-server-ee`
   - `make build-worker-enterprise` → `bin/brokle-worker-ee`
   - Frontend: `cd web && pnpm build`

4. **Build & Push Docker Images**
   - `ghcr.io/brokle-ai/brokle-server:3.2.1`
   - `ghcr.io/brokle-ai/brokle-server:3.2`
   - `ghcr.io/brokle-ai/brokle-server:latest`
   - `ghcr.io/brokle-ai/brokle-worker:3.2.1`
   - Multi-arch: linux/amd64, linux/arm64

5. **Create GitHub Release**
   - Auto-generate release notes from commits
   - Attach binaries (server, worker)
   - Include CHANGELOG.md excerpt
   - Mark as pre-release if `-alpha`, `-beta`, `-rc`

6. **Deploy to Staging** (optional)
   - Trigger ECS/k8s deployment
   - Run smoke tests
   - Health checks

7. **Notify**
   - Slack webhook
   - Discord webhook
   - Email to team

---

### 4. Python SDK Release Workflow

#### Current Setup ✅
- Publish workflow exists at `sdk/python/.github/workflows/publish.yml`
- Triggers on GitHub releases or manual dispatch
- Publishes to PyPI with `twine`

#### Recommended Improvements

**Enhanced CI/CD Pipeline**:
```yaml
name: Python SDK Release

on:
  push:
    tags:
      - 'python-v*'
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      target:
        type: choice
        options: [pypi, testpypi]
        default: pypi

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Validate version consistency
        run: |
          TAG_VERSION="${GITHUB_REF#refs/tags/python-v}"
          FILE_VERSION=$(python -c "from brokle.version import __version__; print(__version__)")
          if [ "$TAG_VERSION" != "$FILE_VERSION" ]; then
            echo "Version mismatch: tag=$TAG_VERSION, file=$FILE_VERSION"
            exit 1
          fi

  test:
    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11', '3.12']
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run full test suite
        run: make test-coverage

  build:
    needs: [validate, test]
    runs-on: ubuntu-latest
    steps:
      - name: Build package
        run: python -m build
      - name: Check package
        run: twine check dist/*

  publish-testpypi:
    needs: build
    if: contains(github.ref, '-alpha') || contains(github.ref, '-beta')
    runs-on: ubuntu-latest
    steps:
      - name: Publish to Test PyPI
        uses: pypa/gh-action-pypi-publish@release/v1
        with:
          repository-url: https://test.pypi.org/legacy/

  publish-pypi:
    needs: build
    if: startsWith(github.ref, 'refs/tags/python-v')
    runs-on: ubuntu-latest
    environment: pypi
    permissions:
      id-token: write  # Trusted Publishing
    steps:
      - name: Publish to PyPI
        uses: pypa/gh-action-pypi-publish@release/v1
        # Uses OIDC Trusted Publishing (no API token needed!)

  create-release:
    needs: publish-pypi
    runs-on: ubuntu-latest
    steps:
      - name: Create GitHub Release
        run: gh release create ${{ github.ref_name }} --generate-notes
```

**Key Improvements**:
1. **Trusted Publishing**: Use OIDC instead of API tokens (more secure)
2. **Version Validation**: Ensure tag matches `brokle/version.py`
3. **Pre-release Detection**: Auto-publish to Test PyPI for alpha/beta
4. **Matrix Testing**: Test across Python 3.9-3.12 before release

#### Release Process
1. **Update Version**: Edit `sdk/python/brokle/version.py`
2. **Update CHANGELOG**: Add release notes to `sdk/python/CHANGELOG.md`
3. **Commit & Tag**:
   ```bash
   git commit -am "chore(python): release v1.2.3"
   git tag python-v1.2.3
   git push origin main --tags
   ```
4. **Automated**:
   - CI runs tests
   - Builds package
   - Publishes to PyPI
   - Creates GitHub release

---

### 5. JavaScript SDK Release Workflow

#### Current Setup ❌
No workflows exist. Need full CI/CD setup.

#### Recommended Tooling: **Changesets**

**Why Changesets?**
- Industry standard (PostHog, many React libraries)
- Excellent monorepo support
- Auto-generates changelogs
- Handles versioning across multiple packages
- Supports pre-releases
- Better than Lerna (deprecated) or manual versioning

**Setup**:
```bash
cd sdk/javascript
pnpm add -D @changesets/cli
pnpm changeset init
```

#### Workflow Architecture

**1. Developer Flow (Adding Changes)**:
```bash
# Make changes to code
git add .

# Create a changeset (interactive)
pnpm changeset
# Choose: patch/minor/major
# Select packages affected
# Write changelog entry

# Commit changeset file
git commit -m "feat: add streaming support"
git push
```

**2. Release Flow (Automated)**:
```yaml
name: JavaScript SDK Release

on:
  push:
    branches: [main]
    paths:
      - 'sdk/javascript/**'
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 22]
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Lint
        run: pnpm lint

      - name: Type check
        run: pnpm typecheck

      - name: Test
        run: pnpm test

      - name: Build
        run: pnpm build

  release:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    permissions:
      contents: write
      pull-requests: write
      id-token: write  # For npm provenance
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pnpm/action-setup@v3
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          registry-url: 'https://registry.npmjs.org'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build packages
        run: pnpm build

      - name: Create Release PR or Publish
        uses: changesets/action@v1
        with:
          publish: pnpm release
          version: pnpm version-packages
          commit: 'chore: release packages'
          title: 'chore: release packages'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

**package.json scripts**:
```json
{
  "scripts": {
    "version-packages": "changeset version",
    "release": "changeset publish"
  }
}
```

#### Release Process with Changesets

**Scenario 1: Regular Release**
1. Developer merges PR with changeset to main
2. Changesets bot creates "Version Packages" PR
3. PR auto-updates package.json versions and CHANGELOGs
4. Maintainer reviews and merges Version Packages PR
5. CI auto-publishes to npm
6. CI creates git tags (`brokle@0.5.0`, `brokle-openai@0.3.0`)

**Scenario 2: Pre-release (Alpha/Beta)**
```bash
# Enter pre-release mode
pnpm changeset pre enter alpha

# Create changesets as normal
pnpm changeset

# Version packages (creates 0.5.0-alpha.1)
pnpm changeset version

# Publish with 'alpha' tag
pnpm changeset publish --tag alpha

# Exit pre-release mode
pnpm changeset pre exit
```

---

### 6. Submodule Coordination & Cross-Repository Workflows

**Challenge**: Platform repo uses SDKs as submodules, but they're in separate repositories with independent release cycles.

**Solution**: Automated workflows to keep platform repo in sync with SDK releases.

---

#### 6.1. Automatic Submodule Update Workflow

**Location**: Platform repo (`.github/workflows/update-submodules.yml`)

**Trigger**: When either SDK releases a new version

**Flow**:
```yaml
name: Update SDK Submodules

on:
  repository_dispatch:
    types: [sdk-released]  # Triggered by SDK repos on release
  workflow_dispatch:        # Manual trigger
    inputs:
      sdk:
        description: 'SDK to update (python, javascript, or both)'
        required: true
        type: choice
        options:
          - python
          - javascript
          - both

jobs:
  update-submodule:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write

    steps:
      - name: Checkout with submodules
        uses: actions/checkout@v4
        with:
          submodules: true
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Update Python SDK submodule
        if: github.event.inputs.sdk == 'python' || github.event.inputs.sdk == 'both' || github.event.client_payload.sdk == 'python'
        run: |
          cd sdk/python
          git fetch origin
          git checkout $(git describe --tags `git rev-list --tags --max-count=1`)
          cd ../..

      - name: Update JavaScript SDK submodule
        if: github.event.inputs.sdk == 'javascript' || github.event.inputs.sdk == 'both' || github.event.client_payload.sdk == 'javascript'
        run: |
          cd sdk/javascript
          git fetch origin
          git checkout $(git describe --tags `git rev-list --tags --max-count=1`)
          cd ../..

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          commit-message: 'chore(deps): update SDK submodules'
          title: 'chore(deps): Update SDK submodules to latest releases'
          body: |
            ## SDK Submodule Updates

            This PR updates SDK submodules to their latest released versions.

            **Python SDK**: `${{ steps.python-version.outputs.version }}`
            **JavaScript SDK**: `${{ steps.js-version.outputs.version }}`

            ### Changes
            - Updated submodule references to latest tags
            - Integration tests will run automatically

            ### Testing
            - [ ] Integration tests pass
            - [ ] Platform builds successfully
            - [ ] No breaking changes detected

            ---
            🤖 Auto-generated by submodule update workflow
          branch: bot/update-sdk-submodules
          delete-branch: true
          labels: |
            dependencies
            automated
```

**Trigger from SDK repos**: Add this to SDK release workflows:

```yaml
# In Python SDK .github/workflows/publish.yml (after successful publish)
- name: Trigger Platform Submodule Update
  if: success()
  run: |
    curl -X POST \
      -H "Accept: application/vnd.github.v3+json" \
      -H "Authorization: token ${{ secrets.PLATFORM_DISPATCH_TOKEN }}" \
      https://api.github.com/repos/brokle-ai/brokle/dispatches \
      -d '{"event_type":"sdk-released","client_payload":{"sdk":"python","version":"${{ steps.version.outputs.version }}"}}'
```

---

#### 6.2. Submodule Version Validation

**Location**: Platform repo (`.github/workflows/validate-submodules.yml`)

**Purpose**: Ensure submodules point to released versions (not random commits)

```yaml
name: Validate Submodules

on:
  pull_request:
  push:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Check Python SDK is on a tagged release
        run: |
          cd sdk/python
          COMMIT=$(git rev-parse HEAD)
          TAG=$(git describe --exact-match $COMMIT 2>/dev/null || echo "")
          if [ -z "$TAG" ]; then
            echo "⚠️  Warning: Python SDK submodule is not on a tagged release"
            echo "Current commit: $COMMIT"
            echo "Consider updating to a release tag"
            exit 1
          fi
          echo "✅ Python SDK on release: $TAG"

      - name: Check JavaScript SDK is on a tagged release
        run: |
          cd sdk/javascript
          COMMIT=$(git rev-parse HEAD)
          TAG=$(git describe --exact-match $COMMIT 2>/dev/null || echo "")
          if [ -z "$TAG" ]; then
            echo "⚠️  Warning: JavaScript SDK submodule is not on a tagged release"
            echo "Current commit: $COMMIT"
            echo "Consider updating to a release tag"
            exit 1
          fi
          echo "✅ JavaScript SDK on release: $TAG"
```

---

#### 6.3. SDK Version Compatibility Matrix

**Location**: Platform repo (`docs/SDK_COMPATIBILITY.md`)

```markdown
# SDK Version Compatibility Matrix

## Current SDK Versions (as of Platform v3.2.0)

| SDK | Version | Commit | Release Date | Notes |
|-----|---------|--------|--------------|-------|
| Python | 0.2.9 | abc1234 | 2025-01-10 | Stable |
| JavaScript | 0.1.0 | def5678 | 2025-01-08 | Beta |

## Version History

### Platform v3.2.0 (2025-01-15)
- Python SDK: v0.2.9
- JavaScript SDK: v0.1.0
- Breaking changes: None

### Platform v3.1.0 (2024-12-20)
- Python SDK: v0.2.8
- JavaScript SDK: v0.0.9
- Breaking changes: API endpoint changes

## Testing Compatibility

Run integration tests with specific SDK versions:

```bash
# Update submodules to specific versions
cd sdk/python && git checkout v0.2.9 && cd ../..
cd sdk/javascript && git checkout v0.1.0 && cd ../..

# Run integration tests
make test-integration
```
```

---

#### 6.4. Manual Submodule Update Process

For developers working locally:

```bash
# Update to latest SDK releases
git submodule update --remote --merge

# Or update specific SDK
cd sdk/python
git checkout v0.2.9  # or latest tag
cd ../..

# Commit the submodule reference change
git add sdk/python
git commit -m "chore(deps): update Python SDK to v0.2.9"
git push
```

**Best Practices**:
1. ✅ Always update to tagged releases (not random commits)
2. ✅ Run integration tests after submodule updates
3. ✅ Document SDK versions in CHANGELOG
4. ✅ Test platform builds after updates
5. ❌ Never commit submodule on unreleased commit

---

#### 6.5. Inter-Repository Dependencies

**Dependency Flow**:
```
Platform (v3.x)
  ├─ Depends on: Python SDK (API client)
  ├─ Depends on: JavaScript SDK (web integrations)
  └─ Independent release cycle

Python SDK (v0.x)
  ├─ Depends on: Platform API (HTTP endpoints)
  ├─ Tests against: Platform locally via Docker
  └─ Can release independently

JavaScript SDK (v0.x)
  ├─ Depends on: Platform API (HTTP endpoints)
  ├─ Tests against: Platform locally via Docker
  └─ Can release independently
```

**Breaking Change Protocol**:
1. Platform breaks SDK API → Major version bump → Warn SDK maintainers
2. SDK breaks platform integration → Coordinate release with platform team
3. Use feature flags for gradual rollouts

---

### 7. Changelog Management

#### Automated Changelog Generation

**Platform (Backend + Web)**:
Use `git-cliff` or `conventional-changelog`:
```yaml
- name: Generate Changelog
  run: |
    git cliff --tag ${{ github.ref_name }} > CHANGELOG.md
```

**Python SDK**:
Manual CHANGELOG.md with sections:
```markdown
## [1.2.3] - 2025-01-15

### Added
- New `observe()` decorator with streaming support

### Changed
- Updated OTLP exporter to v1.20.0

### Fixed
- Fixed memory leak in batch processor

### Deprecated
- `old_method()` will be removed in v2.0.0
```

**JavaScript SDK**:
Auto-generated by Changesets:
```markdown
# brokle

## 0.5.0

### Minor Changes
- abc123f: Add streaming support for OpenAI integration

### Patch Changes
- Updated dependencies
  - brokle-openai@0.3.0
```

---

### 8. Testing Strategy Before Release

#### Platform
- Unit tests: `make test`
- Integration tests: `make test-integration`
- E2E tests: `pnpm test:e2e` (web)
- Docker build test
- Migration tests
- Security scans

#### Python SDK
- Unit tests: `pytest tests/`
- Integration tests with local Brokle server
- Matrix testing: Python 3.9, 3.10, 3.11, 3.12
- Coverage threshold: 80%+
- Type checking: `mypy`
- Security: `bandit`, `safety`

#### JavaScript SDK
- Unit tests: `vitest`
- Integration tests: Test against Brokle server
- Matrix testing: Node 18, 20, 22
- Build test: Ensure all exports work
- Type checking: `tsc --noEmit`
- Package size limits (using `size-limit`)

---

### 9. Release Checklist Templates

#### Platform Release Checklist
```markdown
## Pre-Release
- [ ] All tests passing
- [ ] Migrations tested (up + down)
- [ ] CHANGELOG.md updated
- [ ] Version bumped in appropriate files
- [ ] Security scan clean
- [ ] Documentation updated
- [ ] Breaking changes documented

## Release
- [ ] Tag created: `v3.2.1`
- [ ] CI/CD pipeline passed
- [ ] Docker images pushed
- [ ] GitHub release created
- [ ] Release notes published

## Post-Release
- [ ] Staging deployment verified
- [ ] Production deployment scheduled
- [ ] Monitoring dashboards checked
- [ ] Team notified (Slack/Discord)
- [ ] Documentation site updated
```

#### SDK Release Checklist (Python)
```markdown
## Pre-Release
- [ ] Version updated in `brokle/version.py`
- [ ] CHANGELOG.md updated
- [ ] All tests passing (Python 3.9-3.12)
- [ ] Integration tests with Brokle server passed
- [ ] Type checking passed (mypy)
- [ ] Documentation updated
- [ ] README.md examples verified

## Release
- [ ] Tag created: `python-v1.2.3`
- [ ] PyPI package published
- [ ] GitHub release created
- [ ] Release notes generated

## Post-Release
- [ ] Test installation: `pip install brokle==1.2.3`
- [ ] Verify on Python 3.9, 3.11, 3.12
- [ ] Documentation site updated
- [ ] Example code tested
```

#### SDK Release Checklist (JavaScript)
```markdown
## Pre-Release
- [ ] Changesets created for all changes
- [ ] All tests passing (Node 18, 20, 22)
- [ ] Build successful for all packages
- [ ] Type checking passed
- [ ] Integration tests passed
- [ ] Documentation updated

## Release (via Changesets)
- [ ] "Version Packages" PR reviewed
- [ ] Versions and CHANGELOGs correct
- [ ] PR merged to main
- [ ] npm packages published automatically
- [ ] Git tags created

## Post-Release
- [ ] Test installation: `npm install brokle@latest`
- [ ] Verify on Node 18, 20, 22
- [ ] Documentation site updated
- [ ] Example code tested
```

---

### 10. Tooling & Infrastructure

#### Required Secrets
**GitHub Repository Secrets**:
- `PYPI_API_TOKEN` - Python SDK publishing (or use Trusted Publishing)
- `NPM_TOKEN` - JavaScript SDK publishing
- `DOCKERHUB_USERNAME` & `DOCKERHUB_TOKEN` - Docker image publishing
- `SLACK_WEBHOOK_URL` - Release notifications
- `DISCORD_WEBHOOK_URL` - Community notifications

#### Recommended GitHub Environments
1. **pypi** - For Python SDK releases (enable required reviewers)
2. **npm** - For JavaScript SDK releases
3. **production** - For platform releases

#### Branch Protection Rules
- `main` branch:
  - Require PR reviews (2 approvals)
  - Require status checks (CI must pass)
  - Require linear history
  - Block force pushes
  - Restrict deletions

---

### 11. Documentation Requirements

#### Each Release Must Include:
1. **CHANGELOG.md** - What changed
2. **Migration Guide** - For breaking changes
3. **GitHub Release Notes** - High-level overview
4. **API Documentation** - Updated for new features
5. **Blog Post** - For major/minor releases

#### Release Notes Template
```markdown
## Brokle v3.2.0 - 2025-01-15

### 🎉 Highlights
- New OpenTelemetry-native trace ingestion
- Cost optimization with intelligent caching
- Multi-project support

### ⚠️ Breaking Changes
- Deprecated `/v1/legacy/traces` endpoint
  - **Migration**: Use `/v1/otlp/traces` instead
  - **Timeline**: Legacy endpoint removed in v4.0.0

### ✨ New Features
- **Observability**: OTLP trace ingestion (#123)
- **Billing**: Cost calculation with model pricing (#124)
- **Gateway**: Intelligent provider routing (#125)

### 🐛 Bug Fixes
- Fixed span parent linking (#126)
- Resolved memory leak in batch processor (#127)

### 📚 Documentation
- Added OTLP migration guide
- Updated API reference
- New quickstart tutorial

### 🔧 Internal
- Upgraded ClickHouse to v24.3
- Improved test coverage to 85%

### 📦 Upgrade Instructions
```bash
docker pull ghcr.io/brokle-ai/brokle-server:3.2.0
docker pull ghcr.io/brokle-ai/brokle-worker:3.2.0
```

Full Changelog: v3.1.0...v3.2.0
```

---

### 12. Monitoring & Rollback

#### Release Monitoring
After each release, monitor:
- **Error rates** - Sentry/application logs
- **Performance** - Response times, throughput
- **Database** - Query performance, connection pool
- **Dependencies** - Third-party API health
- **User feedback** - GitHub issues, Discord, email

#### Rollback Strategy

**Platform**:
1. Revert to previous Docker image tag
2. Run database migration rollback if needed: `make migrate-down`
3. Verify health checks

**SDKs**:
1. Yank broken version from PyPI/npm (extreme cases only)
2. Publish hotfix version immediately
3. Communicate issue to users

---

### 13. Release Cadence

#### Platform (Backend + Web)
- **Major releases**: Every 6-12 months (breaking changes)
- **Minor releases**: Every 2-4 weeks (new features)
- **Patch releases**: As needed (bug fixes, security)

#### SDKs (Python + JavaScript)
- **Major releases**: Align with platform breaking changes
- **Minor releases**: Every 2-4 weeks (new features, integrations)
- **Patch releases**: As needed (bug fixes, dependency updates)

#### Pre-releases
- **Alpha**: Weekly (unstable, for testing)
- **Beta**: Bi-weekly (feature complete, bug fixes)
- **RC**: Before major releases (production-ready candidate)

---

### 14. Communication Strategy

#### Release Announcements
1. **GitHub Release** - Technical changelog
2. **Blog Post** - High-level overview (for major/minor)
3. **Discord/Slack** - Quick announcement
4. **Email Newsletter** - Monthly digest
5. **Twitter/X** - Major releases only

#### Deprecation Policy
- Announce deprecation 3 months before removal
- Show deprecation warnings in code
- Provide migration guides
- Support 2 major versions simultaneously

---

### 15. Implementation Phases

**Note**: Each phase targets a specific repository. SDKs are separate repos, so work can be done in parallel.

---

#### Phase 1: Python SDK Improvements (Week 1-2)
**Repository**: `github.com/brokle-ai/brokle-python`

- [ ] Fix release script path bug (`_version.py` → `version.py`)
- [ ] Enable PyPI Trusted Publishing (OIDC)
  - Configure on PyPI.org
  - Update `.github/workflows/publish.yml`
  - Remove `PYPI_API_TOKEN` secret
- [ ] Add version validation to publish workflow
  - Ensure git tag matches `brokle/version.py`
- [ ] Add pre-release auto-detection
  - `-alpha`, `-beta`, `-rc` → auto-publish to TestPyPI
- [ ] Update CHANGELOG.md template
- [ ] Test full release cycle

---

#### Phase 2: JavaScript SDK Complete Setup (Week 3-4)
**Repository**: `github.com/brokle-ai/brokle-js`

- [ ] Install and configure Changesets
  ```bash
  cd sdk/javascript
  pnpm add -D @changesets/cli
  pnpm changeset init
  ```
- [ ] Create `.github/workflows/ci.yml`
  - Lint, typecheck, test (Node 18, 20, 22)
  - Build all packages
- [ ] Create `.github/workflows/release.yml`
  - Changesets integration
  - npm publish with provenance
- [ ] Create `.github/workflows/integration.yml`
  - Test against local Brokle platform
- [ ] Add CLAUDE.md development guide
- [ ] Configure npm publishing
  - Add `NPM_TOKEN` secret
  - Test publish to npm (dry run)
- [ ] Test full release cycle

---

#### Phase 3: Platform Release Automation (Week 5-6)
**Repository**: `github.com/brokle-ai/brokle` (main platform)

- [ ] Create `.github/workflows/release.yml`
  - Trigger on version tags (`v*`)
  - Build Go binaries (server + worker, OSS + Enterprise)
  - Build Next.js frontend
  - Multi-arch Docker images (amd64, arm64)
  - Push to ghcr.io
  - Create GitHub Release
- [ ] Add changelog automation (git-cliff or conventional-changelog)
- [ ] Configure Docker publishing
  - GitHub Container Registry setup
  - Image tagging strategy
- [ ] Add release notification webhooks (Slack/Discord)
- [ ] Test end-to-end platform release

---

#### Phase 4: Submodule Coordination (Week 7)
**Repository**: `github.com/brokle-ai/brokle` (main platform)

- [ ] Create `.github/workflows/update-submodules.yml`
  - `repository_dispatch` trigger from SDK releases
  - Manual workflow_dispatch
  - Auto-create PR with submodule updates
- [ ] Create `.github/workflows/validate-submodules.yml`
  - Ensure submodules on tagged releases
  - Fail PR if pointing to non-release commit
- [ ] Add `PLATFORM_DISPATCH_TOKEN` to SDK repos
  - Configure in Python SDK publish workflow
  - Configure in JavaScript SDK release workflow
- [ ] Create `docs/SDK_COMPATIBILITY.md`
  - Version compatibility matrix
  - Manual update instructions
- [ ] Test cross-repo coordination
  - Release Python SDK → triggers platform PR
  - Release JavaScript SDK → triggers platform PR

---

#### Phase 5: Documentation & Training (Week 8)
**All Repositories**

- [ ] Complete release documentation
  - Platform release guide
  - Python SDK release guide
  - JavaScript SDK release guide
  - Submodule coordination guide
- [ ] Create release checklist templates
  - Platform checklist
  - SDK checklists
- [ ] Update all README.md files
  - Add release badges
  - Link to release documentation
- [ ] Create team training materials
  - Release runbooks
  - Video walkthrough (optional)
  - FAQ document
- [ ] Set up monitoring
  - Release metrics dashboard
  - Error tracking

---

#### Phase 6: Continuous Improvement (Ongoing)

- [ ] Monitor release metrics
  - Time to release
  - Failed releases
  - Rollback frequency
- [ ] Gather team feedback
  - Developer experience survey
  - Identify pain points
- [ ] Optimize workflows
  - Reduce CI/CD time
  - Automate more steps
- [ ] Add Dependabot/Renovate
  - Automated dependency updates
  - SDK dependency tracking

---

## Conclusion

This release strategy provides:
- **Automation**: Reduce manual work, prevent errors
- **Consistency**: Same process for all releases
- **Transparency**: Clear changelogs and release notes
- **Speed**: Ship faster with confidence
- **Quality**: Comprehensive testing before release

By following this strategy, Brokle will have a professional, scalable release process that matches industry leaders like OpenTelemetry, Sentry, and PostHog.

---

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Changesets Documentation](https://github.com/changesets/changesets)
- [OpenTelemetry Release Process](https://github.com/open-telemetry)
- [Sentry Release Documentation](https://develop.sentry.dev/sdk/processes/releases/)
- [GitHub Actions Publishing Guide](https://docs.github.com/en/actions/publishing-packages)
