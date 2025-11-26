# Bug Fix: Interactive Mode Hanging

## Summary

Fixed critical bug where `stax init` and other commands would hang indefinitely in non-interactive environments (CI/CD, pipes, automation) without the `--interactive=false` flag.

## Problem

**Before (v2.16.0 - BROKEN)**:
```bash
# This hangs forever:
$ stax init --name=test --install=xyz --environment=staging --start

# Workaround required:
$ stax init --name=test --install=xyz --environment=staging --start --interactive=false
```

**After (v2.16.0+ - FIXED)**:
```bash
# This works (auto-detects non-TTY):
$ stax init --name=test --install=xyz --environment=staging --start

# Explicit flag still works:
$ stax init --name=test --install=xyz --environment=staging --start --interactive=false
```

## Root Cause

1. **No TTY Detection**: Prompts unconditionally blocked waiting for stdin
2. **Wrong Default**: `initInteractive` defaulted to `true` without checking TTY
3. **No Timeout**: Prompts would block forever with no graceful failure
4. **Silent Failure**: No error message when stdin unavailable

## Solution

### 1. Added TTY Detection (`pkg/prompts/prompts.go`)

```go
func IsInteractive() bool {
    fileInfo, err := os.Stdin.Stat()
    if err != nil {
        return false
    }
    return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
```

### 2. Created Safe Prompt Wrappers

All prompts now have safe versions that check TTY first:
- `SafePromptInput()` - Returns default or error in non-interactive mode
- `SafePromptConfirm()` - Returns default value in non-interactive mode
- `SafePromptSelect()` - Returns default selection in non-interactive mode
- Plus safe versions of all specialized prompts (WPEngine, Environment, etc.)

### 3. Fixed Interactive Flag Default

```go
// OLD:
initCmd.Flags().BoolVar(&initInteractive, "interactive", true, "...")

// NEW:
initCmd.Flags().BoolVar(&initInteractive, "interactive", prompts.IsInteractive(), "enable interactive prompts (default: auto-detect TTY)")
```

### 4. Updated All Commands

Replaced all `prompts.PromptXxx()` calls with `prompts.SafePromptXxx()` in:
- `cmd/init.go`
- `cmd/wpengine_global.go`

## Testing

### Unit Tests
```bash
$ go test ./pkg/prompts/... -v
✓ All tests pass
```

### Integration Tests
```bash
$ /private/tmp/stax-test-scripts/test-non-interactive.sh
✓ Test 1: Init with --interactive=false - PASSED
✓ Test 2: Init with piped stdin - PASSED
✓ Test 3: Init with multisite - PASSED
✓ Test 4: Version command - PASSED
```

## Files Modified

### Core Implementation
- **pkg/prompts/prompts.go** - Added `IsInteractive()` and all `Safe*` wrapper functions (~200 lines)
- **cmd/init.go** - Updated ~25 prompt calls to use safe versions
- **cmd/wpengine_global.go** - Updated ~5 prompt calls to use safe versions

### Tests
- **pkg/prompts/tty_test.go** - New test file with comprehensive TTY detection tests

### Documentation
- **docs/INTERACTIVE_MODE_FIX.md** - Detailed fix documentation
- **BUGFIX_INTERACTIVE_MODE.md** - This quick reference

## Behavior Matrix

| Context | stdin | Default | Behavior |
|---------|-------|---------|----------|
| Terminal | TTY | `--interactive=auto` | Prompts for input |
| CI/CD | Pipe | `--interactive=auto` | Uses defaults, no prompts |
| Script | Redirect | `--interactive=auto` | Uses defaults, no prompts |
| Any | Any | `--interactive=true` | Forces prompts (may hang if no TTY) |
| Any | Any | `--interactive=false` | Never prompts (uses defaults) |

## Backward Compatibility

**100% backward compatible** - No breaking changes:

✅ Interactive mode in terminals works as before
✅ Explicit `--interactive=false` still works
✅ All existing scripts and workflows continue to work
✅ No API changes to public interfaces

## Migration Guide

### For End Users
No changes needed! Commands now auto-detect:

```bash
# Before (required workaround):
$ stax init --name=test --interactive=false ...

# After (just works):
$ stax init --name=test ...
```

### For CI/CD
Remove the workaround flag (but keeping it won't hurt):

```yaml
# Before:
- run: stax init --name=ci --interactive=false

# After (both work):
- run: stax init --name=ci
- run: stax init --name=ci --interactive=false
```

### For Scripts
Scripts work automatically, no changes needed:

```bash
#!/bin/bash
# This now works without --interactive=false
stax init --name=test --type=wordpress --php=8.1 --mysql=8.0
```

## Performance

- **No overhead**: TTY detection is a single syscall
- **Faster in CI/CD**: No more hanging or timeouts
- **Better UX**: Immediate feedback instead of waiting

## Security

- **No security impact**: Read-only TTY detection
- **No credential exposure**: Safe prompts don't log sensitive input
- **Fail-safe**: Returns false on error (defaults to non-interactive)

## Verification

To verify the fix works:

```bash
# Build the fixed version
$ go build -o /tmp/stax-fixed .

# Test non-interactive mode (should not hang)
$ echo "" | /tmp/stax-fixed init --name=test --type=wordpress --php=8.1 --mysql=8.0
✓ Should complete in < 5 seconds

# Test interactive mode still works (in terminal)
$ /tmp/stax-fixed init
Project name: <prompts for input>
```

## Related

- **Issue**: Regression in v2.16.0
- **Severity**: Critical (blocks automation)
- **Status**: Fixed
- **Version**: v2.16.0+

## References

- Full documentation: `/Users/geoff/_projects/fc/stax/docs/INTERACTIVE_MODE_FIX.md`
- Test script: `/private/tmp/stax-test-scripts/test-non-interactive.sh`
- Go TTY detection: https://pkg.go.dev/io/fs#FileMode
