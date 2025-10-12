#!/bin/bash
# Helper script to initialize git and push to GitHub

set -e  # Exit on error

echo "================================================"
echo "  Puzzle App - Push to GitHub Helper Script"
echo "================================================"
echo ""

# Check if git is already initialized
if [ -d .git ]; then
    echo "✓ Git repository already initialized"
else
    echo "→ Initializing git repository..."
    git init
    echo "✓ Git initialized"
fi

# Get GitHub username
echo ""
read -p "Enter your GitHub username: " GITHUB_USER

if [ -z "$GITHUB_USER" ]; then
    echo "❌ Error: GitHub username is required"
    exit 1
fi

# Get repository name (default: puzzle)
echo ""
read -p "Enter repository name [puzzle]: " REPO_NAME
REPO_NAME=${REPO_NAME:-puzzle}

# Update compose.prod.yaml with actual username
echo ""
echo "→ Updating compose.prod.yaml with your GitHub username..."
sed -i "s/USERNAME/$GITHUB_USER/g" compose.prod.yaml
echo "✓ Updated compose.prod.yaml"

# Show status
echo ""
echo "→ Checking git status..."
git status --short

# Stage all files
echo ""
read -p "Stage all files for commit? [Y/n]: " STAGE
STAGE=${STAGE:-Y}

if [[ "$STAGE" =~ ^[Yy]$ ]]; then
    echo "→ Staging files..."
    git add .
    echo "✓ Files staged"
fi

# Create commit
echo ""
read -p "Create initial commit? [Y/n]: " COMMIT
COMMIT=${COMMIT:-Y}

if [[ "$COMMIT" =~ ^[Yy]$ ]]; then
    echo "→ Creating initial commit..."
    git commit -m "Initial commit: Sliding puzzle marketing tool with i18n and CI/CD"
    echo "✓ Initial commit created"
fi

# Add remote
echo ""
REMOTE_URL="https://github.com/$GITHUB_USER/$REPO_NAME.git"
echo "→ Adding remote: $REMOTE_URL"

if git remote get-url origin &>/dev/null; then
    echo "⚠ Remote 'origin' already exists"
    git remote -v
    read -p "Update remote URL? [y/N]: " UPDATE_REMOTE
    if [[ "$UPDATE_REMOTE" =~ ^[Yy]$ ]]; then
        git remote set-url origin "$REMOTE_URL"
        echo "✓ Remote updated"
    fi
else
    git remote add origin "$REMOTE_URL"
    echo "✓ Remote added"
fi

# Ensure we're on main branch
echo ""
echo "→ Ensuring main branch..."
git branch -M main
echo "✓ On main branch"

# Push to GitHub
echo ""
echo "================================================"
echo "  Ready to push to GitHub!"
echo "================================================"
echo ""
echo "Repository: $REMOTE_URL"
echo "Branch: main"
echo ""
read -p "Push to GitHub now? [Y/n]: " PUSH
PUSH=${PUSH:-Y}

if [[ "$PUSH" =~ ^[Yy]$ ]]; then
    echo ""
    echo "→ Pushing to GitHub..."
    echo "  (You may be prompted for your GitHub credentials)"
    echo ""
    
    git push -u origin main
    
    echo ""
    echo "================================================"
    echo "  ✓ Successfully pushed to GitHub!"
    echo "================================================"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Go to: https://github.com/$GITHUB_USER/$REPO_NAME"
    echo "2. Enable GitHub Actions:"
    echo "   → Settings → Actions → General"
    echo "   → Select 'Read and write permissions'"
    echo "   → Save"
    echo "3. Wait for first build (2-5 minutes)"
    echo "4. Make package public (optional):"
    echo "   → Your profile → Packages → puzzle → Settings"
    echo "   → Change visibility to Public"
    echo "5. Create first release:"
    echo "   git tag -a v1.0.0 -m 'Release v1.0.0'"
    echo "   git push origin v1.0.0"
    echo ""
    echo "See GITHUB_SETUP.md for detailed instructions."
    echo ""
else
    echo ""
    echo "Push skipped. To push manually, run:"
    echo "  git push -u origin main"
    echo ""
fi

