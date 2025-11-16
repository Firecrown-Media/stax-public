# Mirror Sync Testing Checklist

This document provides testing procedures for the public mirror sync workflow.

## Pre-Testing Setup

### 1. Verify Deploy Key Configuration

**Private Repository** (Firecrown-Media/stax):
- [ ] Navigate to Settings > Secrets and variables > Actions
- [ ] Verify `PUBLIC_MIRROR_DEPLOY_KEY` secret exists
- [ ] Secret should contain SSH private key starting with `-----BEGIN OPENSSH PRIVATE KEY-----`

**Public Repository** (Firecrown-Media/stax-public):
- [ ] Navigate to Settings > Deploy keys
- [ ] Verify deploy key exists with name like "Stax Public Mirror Sync"
- [ ] Verify "Allow write access" is checked
- [ ] Public key should start with `ssh-ed25519` or `ssh-rsa`

### 2. Verify Repository Setup

**Public Repository**:
- [ ] Repository exists at https://github.com/Firecrown-Media/stax-public
- [ ] Repository is public
- [ ] Repository has main branch
- [ ] Repository description mentions "Distribution repository"

### 3. Verify Workflow Files

- [ ] `.github/workflows/sync-public-mirror.yml` exists
- [ ] Workflow has proper YAML syntax
- [ ] Workflow references correct repositories
- [ ] Workflow uses `PUBLIC_MIRROR_DEPLOY_KEY` secret

## Manual Sync Test

### Test 1: Manual Dispatch with Latest Release

1. **Navigate to Workflow**:
   - [ ] Go to private repository: https://github.com/Firecrown-Media/stax
   - [ ] Click Actions tab
   - [ ] Select "Sync Public Mirror" workflow
   - [ ] Click "Run workflow" button

2. **Execute Workflow**:
   - [ ] Leave tag input empty (to use latest release)
   - [ ] Click "Run workflow"
   - [ ] Wait for workflow to start (refresh page)

3. **Monitor Execution**:
   - [ ] Click on the running workflow
   - [ ] Watch each step complete
   - [ ] Verify no errors in logs
   - [ ] Check "Summary" section shows success message

4. **Verify Public Repository**:
   - [ ] Go to public repository: https://github.com/Firecrown-Media/stax-public
   - [ ] Check main branch shows recent commit
   - [ ] Verify README is from `PUBLIC_MIRROR_README.md`
   - [ ] Check tags page shows synced tag
   - [ ] Verify no `.claude/` directory exists
   - [ ] Verify no `*.claude.md` files exist
   - [ ] Verify no `dist/` directory exists

### Test 2: Manual Dispatch with Specific Tag

1. **Select a Tag**:
   - [ ] List tags in private repo: `git tag -l`
   - [ ] Choose a recent tag (e.g., `v2.4.0`)

2. **Execute Workflow**:
   - [ ] Go to Actions > Sync Public Mirror
   - [ ] Click "Run workflow"
   - [ ] Enter specific tag name (e.g., `v2.4.0`)
   - [ ] Click "Run workflow"

3. **Verify Execution**:
   - [ ] Workflow completes successfully
   - [ ] Summary shows correct tag name
   - [ ] No errors in any step

4. **Verify Public Repository**:
   - [ ] Tag exists in public repository
   - [ ] Tag points to correct commit
   - [ ] Files match release state

### Test 3: Automatic Sync on Release

1. **Create Test Release** (or use existing):
   - [ ] Go to private repository releases
   - [ ] Note the most recent release tag
   - [ ] Check if sync workflow was triggered

2. **Verify Automatic Trigger**:
   - [ ] Go to Actions tab
   - [ ] Find "Sync Public Mirror" run matching release time
   - [ ] Verify it was triggered by "release" event
   - [ ] Check workflow completed successfully

3. **Verify Sync Results**:
   - [ ] Public repository has matching tag
   - [ ] Public repository main branch updated
   - [ ] GoReleaser created release in public repository

## File Cleaning Verification

### Test 4: Verify Sensitive Files Removed

1. **Check Private Repository** (what should be removed):
   ```bash
   # In private repository
   ls -la .claude/              # Should exist in private
   ls -la *.claude.md           # Should exist in private
   ls -la dist/                 # May exist after build
   find . -name ".DS_Store"     # May exist locally
   ```

2. **Check Public Repository** (should be clean):
   ```bash
   # Clone public repository
   git clone https://github.com/Firecrown-Media/stax-public.git
   cd stax-public

   # Verify cleaned files
   ls -la .claude/              # Should NOT exist
   ls -la *.claude.md           # Should NOT exist
   ls -la dist/                 # Should NOT exist
   find . -name ".DS_Store"     # Should NOT exist
   ```

3. **Verify Expected Files Present**:
   ```bash
   # Should exist in public repository
   ls -la README.md             # Public version
   ls -la .goreleaser.yml       # Build config
   ls -la Makefile              # Build scripts
   ls -la docs/                 # Documentation
   ls -la cmd/                  # Source code
   ls -la pkg/                  # Source code
   ls -la internal/             # Source code
   ```

### Test 5: Verify README Replacement

1. **Compare READMEs**:
   - [ ] Open private repo `README.md`
   - [ ] Check for "Repository Structure" section
   - [ ] Check for "Private Development" mentions

2. **Check Public README**:
   - [ ] Open public repo `README.md`
   - [ ] Verify it matches `docs/PUBLIC_MIRROR_README.md`
   - [ ] Check for "Public Distribution Repository" section
   - [ ] Verify no development-specific content

## GoReleaser Integration

### Test 6: Verify GoReleaser Configuration

1. **Check Configuration**:
   ```bash
   # In private repository
   grep "name: stax-public" .goreleaser.yml
   ```
   - [ ] Should show `name: stax-public`
   - [ ] Should NOT show `name: stax`

2. **Test Release Creation** (requires actual release):
   - [ ] Trigger release workflow
   - [ ] Verify GoReleaser runs
   - [ ] Check release created in stax-public
   - [ ] Verify artifacts uploaded to public repo

### Test 7: Homebrew Formula Update

1. **Check Formula Repository**:
   - [ ] Go to https://github.com/Firecrown-Media/homebrew-stax
   - [ ] Check Formula/stax.rb
   - [ ] Verify formula references stax-public
   - [ ] Check version matches latest release

2. **Test Installation**:
   ```bash
   # Update Homebrew
   brew update

   # Install or upgrade Stax
   brew install firecrown-media/stax/stax
   # or
   brew upgrade firecrown-media/stax/stax

   # Verify version
   stax --version
   ```

## Security Testing

### Test 8: Verify No Secrets Exposed

1. **Search Public Repository**:
   ```bash
   # Clone public repository
   cd stax-public

   # Search for potential secrets (should find nothing)
   grep -r "ssh-rsa" .
   grep -r "BEGIN.*PRIVATE KEY" .
   grep -r "ghp_" .
   grep -r "token.*=" .
   grep -r "password.*=" .
   grep -r "api.*key" .
   ```

2. **Verify Secret Patterns**:
   - [ ] No SSH keys found
   - [ ] No API tokens found
   - [ ] No passwords found
   - [ ] No sensitive credentials

### Test 9: Verify Deploy Key Scope

1. **Test Deploy Key Limits**:
   - [ ] Deploy key only works for stax-public
   - [ ] Deploy key cannot access other repositories
   - [ ] Deploy key has write access to stax-public
   - [ ] Deploy key cannot be used for authentication elsewhere

2. **Verify Workflow Permissions**:
   ```yaml
   permissions:
     contents: read  # Should be read-only for source
   ```
   - [ ] Workflow has minimal permissions
   - [ ] Uses deploy key only for push to public

## Error Handling

### Test 10: Missing Deploy Key

1. **Temporarily Remove Secret**:
   - [ ] Go to Settings > Secrets
   - [ ] Remove `PUBLIC_MIRROR_DEPLOY_KEY`

2. **Run Workflow**:
   - [ ] Trigger manual dispatch
   - [ ] Should fail at SSH configuration step
   - [ ] Error message should be clear

3. **Restore Secret**:
   - [ ] Re-add `PUBLIC_MIRROR_DEPLOY_KEY`
   - [ ] Re-run workflow
   - [ ] Should succeed

### Test 11: Invalid Tag

1. **Try Non-existent Tag**:
   - [ ] Manual dispatch with tag `v99.99.99`
   - [ ] Should fail at checkout step
   - [ ] Error should indicate tag not found

2. **Try Invalid Tag Format**:
   - [ ] Manual dispatch with tag `invalid-tag`
   - [ ] Should fail appropriately
   - [ ] Error message should be clear

### Test 12: Network Failure Recovery

1. **Simulate Push Failure**:
   - [ ] Remove write access from deploy key
   - [ ] Run workflow
   - [ ] Should fail at push step
   - [ ] Cleanup should still run

2. **Restore and Verify**:
   - [ ] Re-enable write access
   - [ ] Re-run workflow
   - [ ] Should succeed

## Performance Testing

### Test 13: Sync Speed

1. **Measure Workflow Duration**:
   - [ ] Note workflow start time
   - [ ] Note workflow end time
   - [ ] Calculate total duration
   - [ ] Should complete in < 2 minutes

2. **Check Step Performance**:
   - [ ] Checkout: < 10 seconds
   - [ ] Clean files: < 5 seconds
   - [ ] SSH config: < 5 seconds
   - [ ] Push: < 30 seconds
   - [ ] Verify: < 10 seconds

## Monitoring and Logging

### Test 14: Workflow Summary

1. **Check Summary Output**:
   - [ ] Workflow generates summary
   - [ ] Summary shows tag name
   - [ ] Summary shows repository link
   - [ ] Summary lists cleaned files

2. **Verify Log Quality**:
   - [ ] Each step has clear output
   - [ ] Success messages are visible
   - [ ] Error messages are descriptive
   - [ ] Debug info is available

### Test 15: Verification Step

1. **Check Tag Verification**:
   - [ ] Workflow fetches from public
   - [ ] Verifies tag exists
   - [ ] Confirms main branch updated
   - [ ] Fails if verification fails

2. **Test Verification Failure**:
   - [ ] Temporarily block push (remove deploy key)
   - [ ] Run workflow
   - [ ] Verification should fail
   - [ ] Workflow should report failure

## Rollback Testing

### Test 16: Rollback Procedure

1. **Identify Bad Sync**:
   - [ ] Create test scenario with broken sync
   - [ ] Note the tag that failed

2. **Rollback Steps**:
   ```bash
   # In public repository
   git checkout main
   git reset --hard <previous-good-tag>
   git push origin main --force
   ```
   - [ ] Force push succeeds
   - [ ] Public repo returns to good state

3. **Re-sync**:
   - [ ] Fix issue in private repo
   - [ ] Re-run sync workflow
   - [ ] Verify correct state

## Documentation Testing

### Test 17: Documentation Accuracy

- [ ] `docs/MIRROR_SYNC.md` matches actual implementation
- [ ] `docs/PUBLIC_MIRROR_README.md` is user-friendly
- [ ] Troubleshooting section covers common issues
- [ ] All referenced files exist
- [ ] All links work correctly

## Final Checklist

### Pre-Production

- [ ] All manual tests pass
- [ ] All automatic triggers work
- [ ] File cleaning verified
- [ ] Security checks pass
- [ ] GoReleaser integration works
- [ ] Homebrew formula updates
- [ ] Documentation complete
- [ ] Team trained on process

### Production Readiness

- [ ] Deploy key secured
- [ ] Secrets configured
- [ ] Workflows enabled
- [ ] Monitoring in place
- [ ] Incident response plan documented
- [ ] Rollback procedure tested
- [ ] Team notifications configured

## Continuous Monitoring

### Weekly Checks

- [ ] Review workflow run history
- [ ] Check for failed syncs
- [ ] Verify latest releases synced
- [ ] Monitor public repository state

### Monthly Checks

- [ ] Audit file cleaning rules
- [ ] Review security practices
- [ ] Check deploy key age
- [ ] Update documentation
- [ ] Test rollback procedure

## Issue Reporting

When reporting sync issues, include:
- [ ] Workflow run URL
- [ ] Tag being synced
- [ ] Error messages
- [ ] Steps to reproduce
- [ ] Expected vs actual behavior
- [ ] Screenshots of logs

## Success Criteria

A successful implementation means:
- [ ] Releases automatically sync to public repo
- [ ] Sensitive files never appear in public repo
- [ ] GoReleaser creates releases in public repo
- [ ] Homebrew installs from public repo
- [ ] Manual sync works for any tag
- [ ] Verification catches failures
- [ ] Documentation is complete and accurate
- [ ] Team understands the workflow
