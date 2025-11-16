# ✅ Docker Hub Publishing Implementation Complete

**Date**: 2024-11-16
**Status**: Ready for Configuration
**Repository**: `github.com/brokle-ai/brokle`

---

## What Was Implemented

### Dual Registry Publishing

**Platform Docker images now publish to TWO registries**:

1. **GitHub Container Registry** (ghcr.io)
   - `ghcr.io/brokle-ai/brokle-server`
   - `ghcr.io/brokle-ai/brokle-worker`
   - `ghcr.io/brokle-ai/brokle-web`

2. **Docker Hub** (docker.io) - NEW! ✅
   - `brokle/brokle-server`
   - `brokle/brokle-worker`
   - `brokle/brokle-web`

---

## Changes Made

### 1. CI Workflow Updated
**File**: `.github/workflows/ci.yml`

**Added**:
- Docker Hub login step
- Docker Hub images to all metadata actions (server, worker, web)

**Result**: Main branch pushes publish to BOTH registries

---

### 2. Release Workflow Updated
**File**: `.github/workflows/release.yml`

**Added**:
- Docker Hub login step
- Docker Hub images to matrix configuration
- Updated GitHub Release body with Docker Hub pull commands

**Result**: Tagged releases publish to BOTH registries

---

### 3. GitHub Release Body Enhanced
**Shows users how to pull from**:
- Docker Hub (recommended - more discoverable)
- GitHub Container Registry (alternative)
- Multi-arch support noted (amd64 + arm64)

---

## Setup Required

### Step 1: Create Docker Hub Account

**If using organization (recommended)**:
```
1. Go to: https://hub.docker.com/signup
   Email: brokle.project@gmail.com
   Username: brokleai (or brokle if available)

2. Create organization:
   Go to: https://hub.docker.com/orgs
   Name: brokle
   Email: brokle.project@gmail.com

3. Result: Images will be brokle/brokle-server (professional!)
```

**If using personal account**:
```
Username will be your Docker Hub username
Images: USERNAME/brokle-server
```

---

### Step 2: Generate Access Token

```
1. Go to: https://hub.docker.com/settings/security
2. Click "New Access Token"
3. Description: "GitHub Actions CI/CD"
4. Access permissions: "Read, Write, Delete"
5. Click "Generate"
6. Copy token immediately (won't be shown again!)
```

---

### Step 3: Add GitHub Secrets

```
1. Go to: https://github.com/brokle-ai/brokle/settings/secrets/actions
2. Click "New repository secret"

3. Add DOCKERHUB_USERNAME:
   Name: DOCKERHUB_USERNAME
   Value: brokle (or brokleai, or your Docker Hub username/org name)

4. Add DOCKERHUB_TOKEN:
   Name: DOCKERHUB_TOKEN
   Value: (paste the access token from Step 2)

5. Click "Add secret" for each
```

---

### Step 4: Update Workflows (If Using Different Org Name)

**If your Docker Hub org is NOT "brokle"**, update image names:

**Files to update**:
- `.github/workflows/ci.yml` (lines 167, 190, 213)
- `.github/workflows/release.yml` (lines 144, 149, 155)

**Change**:
```yaml
# From:
brokle/brokle-server

# To:
YOUR_ORG_NAME/brokle-server
```

---

## Where Images Will Be Published

### Main Branch (Continuous Delivery)

**Every push to main** publishes to:

| Image | Docker Hub | GitHub Container Registry |
|-------|------------|--------------------------|
| Server | `brokle/brokle-server:main` | `ghcr.io/brokle-ai/brokle-server:main` |
| Worker | `brokle/brokle-worker:main` | `ghcr.io/brokle-ai/brokle-worker:main` |
| Web | `brokle/brokle-web:main` | `ghcr.io/brokle-ai/brokle-web:main` |

**Plus SHA tags**: `sha-abc1234`

---

### Tagged Releases

**Every version tag** (e.g., v0.1.1) publishes to:

| Image | Docker Hub Tags | GitHub Container Registry Tags |
|-------|-----------------|-------------------------------|
| Server | `v0.1.1`, `v0.1`, `v0`, `latest` | Same |
| Worker | `v0.1.1`, `v0.1`, `v0`, `latest` | Same |
| Web | `v0.1.1`, `v0.1`, `v0`, `latest` | Same |

**Platforms**: linux/amd64, linux/arm64 (multi-arch)

---

## Testing

### Test 1: Verify Secrets Configured

```bash
# Check if secrets exist (can't see values, but can check names)
gh secret list

# Should show:
# DOCKERHUB_USERNAME
# DOCKERHUB_TOKEN
```

---

### Test 2: Push to Main (Trigger CI)

```bash
# Make a small change
echo "\n<!-- Test Docker Hub publishing -->" >> README.md
git add README.md
git commit -m "test: verify Docker Hub publishing"
git push origin main

# Watch workflow:
# https://github.com/brokle-ai/brokle/actions/workflows/ci.yml

# After ~5-10 minutes, verify:
docker pull brokle/brokle-server:main
docker pull ghcr.io/brokle-ai/brokle-server:main

# Both should work! ✅
```

---

### Test 3: Create Release (Full Test)

```bash
# Create v0.1.2 release
make release-patch

# Watch workflow:
# https://github.com/brokle-ai/brokle/actions/workflows/release.yml

# After ~15 minutes, verify Docker Hub:
docker pull brokle/brokle-server:v0.1.2
docker pull brokle/brokle-server:latest
docker pull brokle/brokle-worker:v0.1.2
docker pull brokle/brokle-web:v0.1.2

# Verify multi-arch:
docker pull --platform linux/amd64 brokle/brokle-server:v0.1.2
docker pull --platform linux/arm64 brokle/brokle-server:v0.1.2

# Check Docker Hub UI:
# https://hub.docker.com/r/brokle/brokle-server
# Should see v0.1.2, latest tags
```

---

## Benefits of Docker Hub

### Why Publish to Both Registries?

**Docker Hub** (docker.io):
- ✅ More discoverable (default Docker registry)
- ✅ Better SEO and visibility
- ✅ Familiar to users (most know Docker Hub)
- ✅ Docker Hub Verified Publisher badge (future)
- ✅ Higher pull limits for public images

**GitHub Container Registry** (ghcr.io):
- ✅ Tightly integrated with GitHub
- ✅ Automatic with GITHUB_TOKEN (no setup)
- ✅ Unlimited bandwidth for public images
- ✅ Version control integration

**Publishing to both**: Best of both worlds! 🎯

---

## Docker Hub Organization Setup (Optional)

**Recommended for professional appearance**:

### Create Organization

```
1. Go to: https://hub.docker.com/orgs
2. Click "Create Organization"
3. Name: brokle
4. Billing: Free (for public images)
5. Email: brokle.project@gmail.com
```

### Add Team Members (Future)

```
Organization → Teams → Add member
Invite by Docker Hub username or email
Permissions: Read, Write, or Admin
```

### Repository Settings

```
For each image (brokle/brokle-server, etc.):
1. Description: "Brokle AI Control Plane - HTTP Server"
2. README: Link to GitHub repository
3. Categories: Developer Tools, Monitoring
```

---

## Image URLs

### Docker Hub

- **Server**: https://hub.docker.com/r/brokle/brokle-server
- **Worker**: https://hub.docker.com/r/brokle/brokle-worker
- **Web**: https://hub.docker.com/r/brokle/brokle-web

### GitHub Container Registry

- **Server**: https://github.com/brokle-ai/brokle/pkgs/container/brokle-server
- **Worker**: https://github.com/brokle-ai/brokle/pkgs/container/brokle-worker
- **Web**: https://github.com/brokle-ai/brokle/pkgs/container/brokle-web

---

## Files Modified

1. `.github/workflows/ci.yml` - Added Docker Hub login + image metadata
2. `.github/workflows/release.yml` - Added Docker Hub login + image metadata + release notes

---

## Verification Checklist

Before testing:
- [ ] Docker Hub account created (brokle.project@gmail.com)
- [ ] Docker Hub organization created (optional but recommended)
- [ ] Access token generated on Docker Hub
- [ ] `DOCKERHUB_USERNAME` secret added to GitHub
- [ ] `DOCKERHUB_TOKEN` secret added to GitHub
- [ ] Image names updated in workflows (if not using "brokle" org)

After first publish:
- [ ] Images visible on Docker Hub
- [ ] Images visible on ghcr.io
- [ ] Both registries have same tags
- [ ] Multi-arch images work (amd64 + arm64)
- [ ] `latest` tag points to correct version

---

## Git Commit Message

Here's your commit message (you'll commit):

```bash
git commit -m "feat(docker): add Docker Hub publishing to CI/CD workflows

Publish Docker images to both Docker Hub and GitHub Container Registry
for maximum discoverability and availability.

Changes:
- Add Docker Hub login to ci.yml and release.yml
- Update metadata actions to include both registries:
  - ghcr.io/brokle-ai/* (GitHub Container Registry)
  - brokle/* (Docker Hub)
- Update GitHub Release body with dual registry pull commands
- Add DOCKER_HUB_PUBLISHING_COMPLETE.md guide

Images published:
- brokle/brokle-server (+ ghcr.io/brokle-ai/brokle-server)
- brokle/brokle-worker (+ ghcr.io/brokle-ai/brokle-worker)
- brokle/brokle-web (+ ghcr.io/brokle-ai/brokle-web)

Tags: v0.1.1, v0.1, v0, latest, main, sha-{commit}
Platforms: linux/amd64, linux/arm64

Requires:
- DOCKERHUB_USERNAME secret (Docker Hub username/org)
- DOCKERHUB_TOKEN secret (Docker Hub access token)

Next: Configure Docker Hub account and add GitHub secrets

Co-authored-by: Claude <noreply@anthropic.com>"
```

---

## ✅ Implementation Complete!

**What you need to do**:

1. **Create Docker Hub account/org** (5 min)
2. **Generate access token** (2 min)
3. **Add GitHub secrets** (2 min)
4. **Commit and push** (1 min)
5. **Test** - Push to main or create v0.1.2 tag

**After setup**: Images will automatically publish to BOTH Docker Hub AND ghcr.io! 🐳

---

**Total setup time**: ~10 minutes, then fully automated! 🎉
