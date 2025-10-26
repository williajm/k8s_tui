# Development Workflow Quick Reference

## Daily Development Workflow

### Starting New Work

```bash
# 1. Ensure main is up to date
git checkout main
git pull origin main

# 2. Create feature branch (branch protection requires this)
git checkout -b feature/my-new-feature

# 3. Make changes, commit frequently
git add .
git commit -m "feat: Add new feature description"

# 4. Push to remote
git push origin feature/my-new-feature

# 5. Create PR on GitHub
# Visit: https://github.com/williajm/k8s-tui/compare
# Or use: gh pr create --title "Add new feature" --body "Description"
```

### While PR is Open

```bash
# Make additional changes
git add .
git commit -m "refactor: Improve implementation"
git push origin feature/my-new-feature

# CI will automatically re-run on each push
# Check status at: https://github.com/williajm/k8s-tui/actions
```

### After CI Passes

```bash
# Option 1: Merge via GitHub UI
# - Click "Squash and merge" or "Rebase and merge"
# - Delete branch after merge

# Option 2: Merge via CLI
gh pr merge --squash --delete-branch

# 3. Update local main
git checkout main
git pull origin main
git branch -d feature/my-new-feature  # Clean up local branch
```

## Common Scenarios

### Scenario 1: Quick Bug Fix

```bash
# Create hotfix branch
git checkout main
git pull origin main
git checkout -b hotfix/fix-critical-bug

# Fix the bug
git add .
git commit -m "fix: Resolve critical issue with X"

# Push and create PR
git push origin hotfix/fix-critical-bug
gh pr create --title "Fix critical bug" --body "Fixes #123"

# Wait for CI, then merge immediately
gh pr merge --squash --delete-branch
```

### Scenario 2: Working on Dev Branch

```bash
# Switch to dev branch for experimental work
git checkout dev
git pull origin dev

# Make changes freely (no branch protection on dev)
git add .
git commit -m "Experiment with new approach"
git push origin dev

# When ready to merge to main
git checkout main
git pull origin main
git checkout -b feature/from-dev

# Cherry-pick or merge changes from dev
git cherry-pick <commit-hash>
# Or: git merge dev

# Push and create PR
git push origin feature/from-dev
gh pr create --title "Add feature from dev" --body "Ready for main"
```

### Scenario 3: Emergency Hotfix (Bypass Protection)

**Only if you've configured yourself as bypass actor:**

```bash
# Check out main
git checkout main
git pull origin main

# Make critical fix
git add .
git commit -m "hotfix: Critical production issue"

# Push directly (bypassing PR requirement)
git push origin main

# Note: CI will still run and may fail, but code is already merged
# Fix any CI failures immediately with follow-up commits
```

### Scenario 4: CI Fails on Your PR

```bash
# Check the error in GitHub Actions
# Visit: https://github.com/williajm/k8s-tui/actions

# Fix the issue locally
git add .
git commit -m "fix: Resolve linting errors"
git push origin feature/my-feature

# CI will automatically re-run
# Watch: https://github.com/williajm/k8s-tui/actions
```

### Scenario 5: Update PR with Main Changes

```bash
# If main has moved ahead while your PR is open
git checkout feature/my-feature
git fetch origin

# Option 1: Rebase (cleaner history, preferred)
git rebase origin/main
git push --force-with-lease origin feature/my-feature

# Option 2: Merge (simpler, but creates merge commit)
git merge origin/main
git push origin feature/my-feature
```

## Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding missing tests
- `chore`: Changes to build process or auxiliary tools
- `ci`: CI/CD configuration changes

**Examples:**
```bash
git commit -m "feat(ui): Add pod log streaming view"
git commit -m "fix(k8s): Handle connection timeout gracefully"
git commit -m "docs: Update README with installation instructions"
git commit -m "test: Add unit tests for pod list component"
git commit -m "ci: Add coverage reporting to workflow"
```

## Running Tests Locally Before Pushing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector (slower, catches concurrency issues)
go test -race ./...

# Run linter
golangci-lint run ./...

# Format code
go fmt ./...
gofmt -w -s .
```

## Pre-Push Checklist

Before pushing code, ensure:

- [ ] All tests pass: `go test ./...`
- [ ] Code is formatted: `go fmt ./...`
- [ ] No linting errors: `golangci-lint run ./...`
- [ ] Commit messages follow convention
- [ ] No sensitive data (API keys, passwords) in commits
- [ ] Changes are on a feature branch, not main

## Git Aliases for Faster Workflow

Add to `~/.gitconfig`:

```ini
[alias]
    # Quick status
    st = status -sb

    # Pretty log
    lg = log --oneline --graph --decorate --all

    # Quick commit
    c = commit -m

    # Amend last commit
    amend = commit --amend --no-edit

    # Push current branch
    pushf = push -u origin HEAD

    # Pull with rebase
    pullr = pull --rebase origin main

    # Clean up merged branches
    cleanup = !git branch --merged | grep -v '\\*\\|main\\|dev' | xargs -n 1 git branch -d

    # Create feature branch
    feature = !sh -c 'git checkout main && git pull && git checkout -b feature/$1' -

    # Create hotfix branch
    hotfix = !sh -c 'git checkout main && git pull && git checkout -b hotfix/$1' -
```

Usage:
```bash
git st                          # Quick status
git feature my-new-thing        # Create feature/my-new-thing branch
git c "feat: Add new thing"     # Quick commit
git pushf                       # Push current branch
git cleanup                     # Remove merged branches
```

## GitHub CLI Shortcuts

```bash
# Create PR
gh pr create

# Create PR with auto-fill
gh pr create --fill

# Check PR status
gh pr status

# View PR checks
gh pr checks

# Merge PR
gh pr merge --squash --delete-branch

# View CI runs
gh run list

# Watch CI run
gh run watch

# View PR in browser
gh pr view --web
```

## Troubleshooting

### "Protected branch update failed"

**Cause:** Trying to push directly to main.

**Solution:**
```bash
# Create a branch and PR instead
git checkout -b fix/my-fix
git push origin fix/my-fix
gh pr create
```

### "Required status checks must pass"

**Cause:** CI tests/linting failed.

**Solution:**
```bash
# Check which checks failed
gh pr checks

# View detailed logs
gh run view <run-id>

# Fix locally and push again
```

### "This branch has conflicts with the base branch"

**Cause:** Main has changed since you branched.

**Solution:**
```bash
git checkout feature/my-branch
git fetch origin
git rebase origin/main
# Resolve conflicts if any
git push --force-with-lease origin feature/my-branch
```

### Force push accidentally disabled

**Cause:** Trying to force push to protected branch.

**Solution:**
- Never force push to main (protection prevents this)
- Only force push to feature branches after rebase
- Use `--force-with-lease` for safety

## Resources

- [Git Documentation](https://git-scm.com/doc)
- [GitHub CLI Manual](https://cli.github.com/manual/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
