# GitHub Repository Setup Guide

This guide walks you through setting up the GitHub repository with automated container builds.

## Step 1: Create GitHub Repository

1. Go to https://github.com/new
2. Fill in repository details:
   - **Repository name**: `puzzle` (or your preferred name)
   - **Description**: "Sliding puzzle marketing tool with discount code generation"
   - **Visibility**: Choose **Public** (for free GitHub Container Registry) or **Private**
   - **DO NOT** check "Add a README file"
   - **DO NOT** check "Add .gitignore"
   - **DO NOT** add a license yet
3. Click **Create repository**

GitHub will show you quick setup commands. Keep this page open.

## Step 2: Initialize Git Repository Locally

If you haven't already initialized git in your project:

```bash
cd /home/admindt/Projecten/repute-software/puzzle

# Initialize git
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Sliding puzzle marketing tool with i18n support"
```

## Step 3: Connect to GitHub

Replace `USERNAME` with your actual GitHub username:

```bash
# Add remote repository
git remote add origin https://github.com/USERNAME/puzzle.git

# Rename branch to main (if needed)
git branch -M main

# Push to GitHub
git push -u origin main
```

You'll be prompted for GitHub credentials. If you have 2FA enabled, use a Personal Access Token instead of your password.

### Creating a Personal Access Token (if needed)

1. Go to https://github.com/settings/tokens
2. Click "Generate new token" → "Generate new token (classic)"
3. Give it a name: "Puzzle App Deployment"
4. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `write:packages` (Upload packages to GitHub Package Registry)
   - ✅ `delete:packages` (Delete packages from GitHub Package Registry)
5. Click "Generate token"
6. **Copy the token** (you won't see it again!)
7. Use this token as your password when pushing

## Step 4: Enable GitHub Actions

1. Go to your repository on GitHub
2. Click **Settings** (top menu)
3. In the left sidebar, click **Actions** → **General**
4. Scroll to **Workflow permissions**
5. Select **"Read and write permissions"**
6. Check ✅ **"Allow GitHub Actions to create and approve pull requests"**
7. Click **Save**

## Step 5: Verify GitHub Actions Workflow

1. Go to your repository on GitHub
2. Click the **Actions** tab
3. You should see a workflow run for your initial push
4. Click on the workflow to see build progress
5. Wait for it to complete (usually 2-5 minutes)

If the build succeeds, your Docker image has been published! 🎉

## Step 6: Make Package Public (Recommended)

By default, packages are private. To make it public:

1. Go to your GitHub profile: https://github.com/USERNAME
2. Click the **Packages** tab
3. Find and click on **puzzle** package
4. Click **Package settings** (right sidebar)
5. Scroll to **Danger Zone**
6. Click **Change visibility**
7. Select **Public**
8. Type the repository name to confirm
9. Click **I understand, change package visibility**

Now anyone can pull your image without authentication!

## Step 7: Update compose.prod.yaml

Update the image URL in `compose.prod.yaml` with your actual GitHub username:

```bash
# Edit the file
vim compose.prod.yaml

# Change this line:
# image: ghcr.io/USERNAME/puzzle:latest
# To (replace with your username):
image: ghcr.io/yourusername/puzzle:latest

# Commit and push
git add compose.prod.yaml
git commit -m "Update production compose with actual GitHub username"
git push
```

## Step 8: Create First Release (Optional but Recommended)

Create a version tag to mark your first production-ready release:

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0: Initial production release with i18n support"

# Push the tag
git push origin v1.0.0
```

This triggers another build that creates multiple tags:
- `ghcr.io/USERNAME/puzzle:v1.0.0` (exact version)
- `ghcr.io/USERNAME/puzzle:v1.0` (minor version)
- `ghcr.io/USERNAME/puzzle:v1` (major version)
- `ghcr.io/USERNAME/puzzle:latest` (always latest)

## Step 9: Test Production Deployment

On your production server (or test it locally):

```bash
# Pull the repository
git clone https://github.com/USERNAME/puzzle.git
cd puzzle

# Configure for production
vim config.yaml  # Set testing_mode: false

# Add images
cp /path/to/your-images/*.png images/

# Deploy
docker compose -f compose.prod.yaml up -d

# Check status
docker compose -f compose.prod.yaml ps

# View logs
docker compose -f compose.prod.yaml logs -f
```

## What Happens Automatically

### On Push to Main Branch
1. GitHub Actions triggers
2. Builds Docker image
3. Tags as `latest` and `main`
4. Pushes to `ghcr.io/USERNAME/puzzle:latest`

### On Git Tag (v1.2.3)
1. GitHub Actions triggers
2. Builds Docker image
3. Tags as `v1.2.3`, `v1.2`, `v1`, and `latest`
4. Pushes all tags to GitHub Container Registry

### On Pull Request
1. GitHub Actions triggers
2. Builds Docker image (tests build)
3. Does NOT push to registry (just validates)

## Troubleshooting

### Build Fails on GitHub Actions

Check the Actions tab for error details:
```bash
# Common issues:
# 1. Go version mismatch - check Dockerfile and go.mod
# 2. Missing files - check .gitignore isn't excluding needed files
# 3. Test failures - run tests locally first
```

### Can't Push to GitHub

```bash
# Check remote URL
git remote -v

# Should show:
# origin  https://github.com/USERNAME/puzzle.git (fetch)
# origin  https://github.com/USERNAME/puzzle.git (push)

# If wrong, update:
git remote set-url origin https://github.com/USERNAME/puzzle.git
```

### Package is Private

Follow Step 6 above to make it public, or authenticate when pulling:

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull private image
docker pull ghcr.io/USERNAME/puzzle:latest
```

### Image Not Found

Wait a few minutes after the GitHub Action completes. The image needs time to be available in the registry.

Check packages at: https://github.com/USERNAME?tab=packages

## Workflow Summary

```
Local Development → Git Push → GitHub Actions → Build Image → GitHub Container Registry → Production Server
       ↓                ↓              ↓              ↓                    ↓                        ↓
   Edit Code      Push to main    Run workflow   Build Docker    Push to ghcr.io         Pull & deploy
```

## Future Updates

To update the production deployment with new code:

```bash
# 1. Make changes locally
vim handlers/game.go

# 2. Test locally
./stop.sh
podman build -t sliding-puzzle:latest .
./start.sh

# 3. Commit and push
git add .
git commit -m "Add new feature"
git push

# 4. Wait for GitHub Actions to build (2-5 minutes)

# 5. On production server, pull and restart
docker compose -f compose.prod.yaml pull
docker compose -f compose.prod.yaml up -d
```

For stable releases, use tags:

```bash
# Create new version
git tag -a v1.1.0 -m "Release v1.1.0: Add new feature"
git push origin v1.1.0

# On production, use specific version
vim compose.prod.yaml  # Change to :v1.1.0
docker compose -f compose.prod.yaml pull
docker compose -f compose.prod.yaml up -d
```

## Next Steps

1. ✅ Repository created and pushed
2. ✅ GitHub Actions configured
3. ✅ First image built and published
4. ✅ Package made public
5. ✅ Production compose updated
6. ✅ First release tagged
7. → Deploy to production server
8. → Share with your team
9. → Start using for marketing campaigns!

---

**Congratulations!** 🎉 Your puzzle app is now set up with professional CI/CD pipeline and ready for production deployment.

For production deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).

