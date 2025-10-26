# Branch Protection Setup Script for K8S-TUI (Solo Developer)
# Usage: .\setup-branch-protection.ps1 [-Branch "main"]

param(
    [string]$Branch = "main"
)

$ErrorActionPreference = "Stop"

$Repo = "williajm/k8s-tui"

Write-Host "Setting up branch protection for: $Branch" -ForegroundColor Cyan
Write-Host "Repository: $Repo" -ForegroundColor Cyan
Write-Host ""

# Check if gh CLI is installed
try {
    $null = Get-Command gh -ErrorAction Stop
    Write-Host "✅ GitHub CLI is installed" -ForegroundColor Green
} catch {
    Write-Host "❌ GitHub CLI (gh) is not installed." -ForegroundColor Red
    Write-Host "Install from: https://cli.github.com/" -ForegroundColor Yellow
    exit 1
}

# Check if authenticated
try {
    gh auth status 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Not authenticated"
    }
    Write-Host "✅ Authenticated with GitHub CLI" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Not authenticated with GitHub CLI." -ForegroundColor Red
    Write-Host "Run: gh auth login" -ForegroundColor Yellow
    exit 1
}

# Create the protection rule
Write-Host "Creating branch protection rule..." -ForegroundColor Cyan

$protectionConfig = @"
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
"@

try {
    $protectionConfig | gh api `
        --method PUT `
        -H "Accept: application/vnd.github+json" `
        "/repos/$Repo/branches/$Branch/protection" `
        --input -

    Write-Host ""
    Write-Host "✅ Branch protection rule created successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚙️  Settings Applied:" -ForegroundColor Cyan
    Write-Host "  • Require pull requests before merging (0 approvals)"
    Write-Host "  • Require status checks to pass"
    Write-Host "  • Require linear history (no merge commits)"
    Write-Host "  • Prevent force pushes"
    Write-Host "  • Prevent branch deletion"
    Write-Host "  • Admin bypass: ENABLED (for solo developer flexibility)"
    Write-Host ""
    Write-Host "📝 Next Steps:" -ForegroundColor Yellow
    Write-Host "  1. Go to: https://github.com/$Repo/settings/branches"
    Write-Host "  2. Edit the '$Branch' rule"
    Write-Host "  3. Add required status checks:"
    Write-Host "     - Search for 'test' and add all matrix jobs"
    Write-Host "     - Search for 'lint' and add"
    Write-Host "     - Search for 'build' and add all matrix jobs"
    Write-Host "  4. (Optional) Add yourself to 'Allow bypass' for emergencies"
    Write-Host ""
    Write-Host "🔗 View rule: https://github.com/$Repo/settings/branch_protection_rules" -ForegroundColor Cyan
} catch {
    Write-Host ""
    Write-Host "❌ Failed to create branch protection rule" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    Write-Host "You may need to set it up manually via GitHub web UI" -ForegroundColor Yellow
    exit 1
}
