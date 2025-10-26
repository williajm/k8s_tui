#!/bin/bash
set -e

# Branch Protection Setup Script for K8S-TUI (Solo Developer)
# Usage: ./setup-branch-protection.sh [branch-name]

BRANCH="${1:-main}"
REPO="williajm/k8s-tui"

echo "Setting up branch protection for: $BRANCH"
echo "Repository: $REPO"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed."
    echo "Install from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI."
    echo "Run: gh auth login"
    exit 1
fi

echo "✅ GitHub CLI is installed and authenticated"
echo ""

# Create the protection rule using GitHub API
echo "Creating branch protection rule..."

# Note: We use a simplified rule that doesn't require specific check names
# since those are generated dynamically by the matrix jobs
gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/$REPO/branches/$BRANCH/protection" \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "checks": []
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "required_approving_review_count": 0,
    "dismiss_stale_reviews": false,
    "require_code_owner_reviews": false,
    "require_last_push_approval": false
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_linear_history": true,
  "required_conversation_resolution": false,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Branch protection rule created successfully!"
    echo ""
    echo "⚙️  Settings Applied:"
    echo "  • Require pull requests before merging (0 approvals)"
    echo "  • Require status checks to pass"
    echo "  • Require linear history (no merge commits)"
    echo "  • Prevent force pushes"
    echo "  • Prevent branch deletion"
    echo "  • Admin bypass: ENABLED (for solo developer flexibility)"
    echo ""
    echo "📝 Next Steps:"
    echo "  1. Go to: https://github.com/$REPO/settings/branches"
    echo "  2. Edit the '$BRANCH' rule"
    echo "  3. Add required status checks:"
    echo "     - Search for 'test' and add all matrix jobs"
    echo "     - Search for 'lint' and add"
    echo "     - Search for 'build' and add all matrix jobs"
    echo "  4. (Optional) Add yourself to 'Allow bypass' for emergencies"
    echo ""
    echo "🔗 View rule: https://github.com/$REPO/settings/branch_protection_rules"
else
    echo ""
    echo "❌ Failed to create branch protection rule"
    echo "You may need to set it up manually via GitHub web UI"
    exit 1
fi
