# Branch Protection Rules for K8S-TUI

This document describes the recommended branch protection settings for a lone developer workflow.

## Philosophy for Solo Development

As a solo developer, branch protection serves to:
- ✅ Prevent accidental force pushes to main
- ✅ Ensure CI passes before merging
- ✅ Maintain code quality through automated checks
- ✅ Create a safety net without blocking solo work

## Recommended Settings for `main` Branch

### 1. Require Pull Request Before Merging

**Enable:** ✅ Yes
**Settings:**
- ✅ Require approvals: **0** (lone developer doesn't need self-approval)
- ✅ Dismiss stale pull request approvals when new commits are pushed: No
- ✅ Require review from Code Owners: No
- ✅ Require approval of the most recent reviewable push: No
- ✅ Allow specified actors to bypass required pull requests: **Your username** (for hotfixes)

**Why:** Encourages working on feature branches while allowing flexibility for urgent fixes.

### 2. Require Status Checks Before Merging

**Enable:** ✅ Yes
**Settings:**
- ✅ Require branches to be up to date before merging: **Yes**
- ✅ Status checks that are required:
  - `test` (all matrix jobs)
  - `lint`
  - `build` (all matrix jobs)

**Why:** Ensures all CI checks pass before code enters main branch.

### 3. Require Conversation Resolution Before Merging

**Enable:** ❌ No

**Why:** Not needed for solo development.

### 4. Require Signed Commits

**Enable:** ⚠️ Optional (Recommended for security)

**Why:** Verifies commit authenticity. Setup: https://docs.github.com/en/authentication/managing-commit-signature-verification

### 5. Require Linear History

**Enable:** ✅ Yes

**Why:** Keeps git history clean and easy to follow. Use "Rebase and merge" or "Squash and merge" instead of merge commits.

### 6. Require Deployments to Succeed Before Merging

**Enable:** ❌ No

**Why:** Not applicable for this project (no deployments yet).

### 7. Lock Branch

**Enable:** ❌ No

**Why:** Would prevent all pushes. Only use for archived projects.

### 8. Do Not Allow Bypassing the Above Settings

**Enable:** ❌ No

**Why:** As a solo dev, you need the ability to bypass in emergencies.

### 9. Restrict Who Can Push to Matching Branches

**Enable:** ⚠️ Optional
**Settings:**
- Restrict pushes that create matching branches: No
- Allow specific actors to bypass: Your username

**Why:** Extra safety layer, but can be skipped for solo projects.

### 10. Allow Force Pushes

**Enable:** ❌ No

**Why:** Prevents accidentally overwriting history on main.

### 11. Allow Deletions

**Enable:** ❌ No

**Why:** Prevents accidentally deleting the main branch.

## Workflow with Branch Protection

### Standard Feature Development

```bash
# Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/my-new-feature

# Make changes and commit
git add .
git commit -m "Add new feature"

# Push to remote
git push origin feature/my-new-feature

# Create PR on GitHub
# Wait for CI to pass (or fix issues)
# Merge PR using "Squash and merge" or "Rebase and merge"
```

### Hotfix (Bypass Protection)

If you've set yourself as an allowed actor to bypass:

```bash
# Option 1: Direct push to main (emergency only)
git checkout main
git pull origin main
# Make quick fix
git add .
git commit -m "hotfix: Critical bug fix"
git push origin main

# Option 2: Fast PR workflow
git checkout -b hotfix/critical-issue
# Make fix
git push origin hotfix/critical-issue
# Create PR, CI passes, merge immediately
```

### Working on Dev Branch

```bash
# Normal development happens on dev
git checkout dev
git pull origin dev

# Make changes
git add .
git commit -m "WIP: Testing new feature"
git push origin dev

# When ready, create PR from dev -> main
# This triggers CI and requires passing checks
```

## Setting Up Branch Protection on GitHub

### Via GitHub Web UI

1. Go to your repository on GitHub
2. Click **Settings** → **Branches**
3. Click **Add branch protection rule**
4. Branch name pattern: `main`
5. Configure settings as described above
6. Click **Create** or **Save changes**

### Via GitHub CLI

```bash
# Install GitHub CLI if not already installed
# https://cli.github.com/

# Set up basic protection for main
gh api repos/williajm/k8s-tui/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["test","lint","build"]}' \
  --field enforce_admins=false \
  --field required_pull_request_reviews='{"required_approving_review_count":0,"dismiss_stale_reviews":false}' \
  --field restrictions=null \
  --field allow_force_pushes=false \
  --field allow_deletions=false \
  --field required_linear_history=true
```

### Via GitHub API (curl)

Create a file `branch-protection.json`:

```json
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "test (ubuntu-latest, 1.21)",
      "test (ubuntu-latest, 1.22)",
      "test (ubuntu-latest, 1.23)",
      "test (macos-latest, 1.21)",
      "test (macos-latest, 1.22)",
      "test (macos-latest, 1.23)",
      "test (windows-latest, 1.21)",
      "test (windows-latest, 1.22)",
      "test (windows-latest, 1.23)",
      "lint",
      "build (ubuntu-latest)",
      "build (macos-latest)",
      "build (windows-latest)"
    ]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "required_approving_review_count": 0,
    "dismiss_stale_reviews": false,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": true
}
```

Then apply:

```bash
curl -X PUT \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer YOUR_GITHUB_TOKEN" \
  https://api.github.com/repos/williajm/k8s-tui/branches/main/protection \
  -d @branch-protection.json
```

## Verifying Protection Rules

```bash
# Check current protection status
gh api repos/williajm/k8s-tui/branches/main/protection

# Or via web UI
# Settings → Branches → View rule for 'main'
```

## Benefits for Solo Development

1. **Safety Net**: Prevents accidental destructive operations
2. **Quality Gate**: Forces you to run CI before merging
3. **Clean History**: Encourages rebase/squash for readable git log
4. **Professional Practice**: Same workflow as team projects
5. **Future-Proof**: Easy transition when collaborators join

## Common Scenarios

### Scenario 1: Quick Fix Needed

**Problem:** CI takes 10 minutes, but you need to fix a typo in README.

**Solution:**
- If you're set as bypass actor: Push directly to main
- Otherwise: Create PR, skip waiting, merge with admin override
- Trade-off: CI will still run and show failure/success after merge

### Scenario 2: Experimental Work

**Problem:** Want to try something risky without affecting main.

**Solution:**
- Use dev branch (no protection)
- Or create feature branch
- Delete when done experimenting

### Scenario 3: CI Failure on Valid Change

**Problem:** Linter fails on generated code or false positive.

**Solution:**
- Fix the linter config (`.golangci.yml`)
- Or add `//nolint:rulename` comment
- Or bypass protection (last resort)

## Recommended Development Flow

```
main (protected)
  ↑
  │ PR (requires CI passing)
  │
dev (unprotected, auto-deploys to staging if you add CD later)
  ↑
  │ Regular commits
  │
feature/xxx (short-lived branches)
```

## Additional Resources

- [GitHub Branch Protection Docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [GitHub CLI Branch Protection](https://cli.github.com/manual/gh_api)
- [Signing Commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)

## Summary: Your Quick Setup Checklist

- [ ] Navigate to Settings → Branches on GitHub
- [ ] Add rule for `main` branch
- [ ] ✅ Require pull request (0 approvals)
- [ ] ✅ Require status checks: test, lint, build
- [ ] ✅ Require linear history
- [ ] ❌ Disable force pushes
- [ ] ❌ Disable deletions
- [ ] ⚠️ Optional: Add yourself as bypass actor
- [ ] Save and test with a feature branch PR
