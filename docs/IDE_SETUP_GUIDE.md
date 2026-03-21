# IDE Setup & Test Discovery Fix Guide

## Problem Diagnosis

Your IDE is experiencing test discovery issues due to a **dual-workspace configuration**:

### Root Causes Identified

1. **Two identical agg_go directories**
   - `/mnt/projekte/Code/agg_go` (primary, currently active)
   - `/home/christian/Code/agg_go` (secondary)
   - Both are complete git repositories pointing to the same remote
   - Both have identical test files causing duplicates in the test explorer

2. **Gopls caching both workspaces**
   - Multiple gopls processes running (at least 4 active)
   - Gopls cache at `~/.cache/gopls/1d501daf/` contains merged information from both directories
   - This causes the IDE to see duplicate tests from both workspace paths

3. **Test duplication mechanism**
   - Go test discovery (`go test -list .`) shows 7 instances of TestEdgeCases, 4 of TestVertexGeneration, etc.
   - These duplicates come from the same test packages being indexed from multiple roots
   - IDE shows these as separate test items in the test explorer

4. **IDE hanging during test runs**
   - Gopls attempting to run tests against both workspace paths simultaneously
   - Memory pressure from duplicate package indexing (gopls using 1.3GB and 663MB in separate processes)
   - Race conditions when multiple instances try to cache the same module

## Solution Strategy

### Phase 1: Consolidate Workspaces (Required)

Choose which directory to use as your primary workspace:

**Option A: Use `/mnt/projekte/Code/agg_go` (Recommended)**

- More accessible path
- Likely your actual working directory
- Keep this, remove the `/home/christian/Code/agg_go` copy

**Option B: Use `/home/christian/Code/agg_go`**

- Update your IDE to point here instead
- Delete `/mnt/projekte/Code/agg_go`

### Phase 2: Clean Gopls State

After consolidating workspaces:

```bash
# Kill all gopls processes
pkill -9 gopls

# Clear gopls cache (requires removing entire cache)
rm -rf ~/.cache/gopls/

# Clear gopls configuration
rm -rf ~/.config/gopls/

# Optional: Clear Go build cache
go clean -cache -testcache -modcache

# Restart your IDE (will restart gopls automatically)
```

### Phase 3: IDE Configuration

#### VS Code Settings

Create or update `.vscode/settings.json` in your primary workspace:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.vetOnSave": "package",
  "go.useLanguageServer": true,
  "go.languageServerFlags": ["-logfile", "/tmp/gopls.log", "-rpc.trace"],
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  },
  "go.testOnSave": "off",
  "go.testTimeout": "30s",
  "go.coverageOptions": "setbenchmarkstats",
  "go.coverageDecorator": "highlight",
  "go.testExplorer.showOutput": true,
  "go.testExplorer.concatenateOutput": true
}
```

#### Gopls Configuration

Create `~/.config/gopls/settings.json`:

```json
{
  "deepCompletion": true,
  "completeUnimported": true,
  "staticcheck": true,
  "matcher": "Fuzzy",
  "symbolMatcher": "FastFuzzy",
  "gofumpt": true,
  "hints": {
    "assignVariableTypes": false,
    "compositeLiteralFields": false,
    "compositeLiteralTypes": true,
    "constantValues": false,
    "functionTypeParameters": false,
    "parameterNames": false,
    "rangeVariableTypes": false
  },
  "standaloneTags": ["integration", "bench", "e2e"],
  "codelens": {
    "generate": true,
    "test": true,
    "run_govulncheck": false,
    "tidy": false,
    "upgrade_dependency": false,
    "vendor": false
  }
}
```

## Implementation Steps

### Step 1: Backup Your Work

```bash
cd /mnt/projekte/Code/agg_go
git status  # Ensure no uncommitted changes
git log --oneline -5  # Verify you're on the right branch
```

### Step 2: Choose Primary Workspace

```bash
# Option A: Keep /mnt/projekte/Code/agg_go (recommended)
rm -rf /home/christian/Code/agg_go

# Option B: If you prefer /home/christian/Code/agg_go
# rm -rf /mnt/projekte/Code/agg_go
# cd /home/christian/Code/agg_go
```

### Step 3: Clean All IDE Caches

```bash
# Kill all gopls processes
pkill -9 gopls

# Remove caches
rm -rf ~/.cache/gopls/
rm -rf ~/.config/gopls/

# Go cache clean
go clean -cache -testcache -modcache
```

### Step 4: Create IDE Configuration

```bash
# Create VS Code settings
mkdir -p /mnt/projekte/Code/agg_go/.vscode
cat > /mnt/projekte/Code/agg_go/.vscode/settings.json << 'EOF'
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.vetOnSave": "package",
  "go.useLanguageServer": true,
  "go.languageServerFlags": [
    "-logfile", "/tmp/gopls.log"
  ],
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  },
  "go.testOnSave": "off",
  "go.testTimeout": "30s",
  "go.testExplorer.showOutput": true
}
EOF
```

### Step 5: Create Gopls Configuration

```bash
mkdir -p ~/.config/gopls
cat > ~/.config/gopls/settings.json << 'EOF'
{
  "deepCompletion": true,
  "completeUnimported": true,
  "staticcheck": true,
  "matcher": "Fuzzy",
  "symbolMatcher": "FastFuzzy",
  "gofumpt": true
}
EOF
```

### Step 6: Verify Configuration

```bash
# Start fresh gopls
cd /mnt/projekte/Code/agg_go

# Verify tests are clean
go test -list ./... 2>&1 | sort | uniq -c | sort -rn | head -10
# Should show each test exactly ONCE

# Run quick validation
just quick  # fmt + vet
```

### Step 7: Restart IDE

Close and reopen VS Code completely (not just the folder). This ensures:

- Fresh gopls instance starts
- Proper workspace initialization
- Cache rebuilt from scratch

### Step 8: Verify Test Discovery

In VS Code:

1. Open Test Explorer (left sidebar)
2. Click refresh button
3. Verify each test appears exactly once
4. Click one test to run - should complete without hanging

## Troubleshooting

### Tests still show duplicates

```bash
# Check for remaining duplicate directories
find ~ -name "agg_go" -type d 2>/dev/null

# Verify only one active workspace
ls -la /mnt/projekte/Code/agg_go/.vscode/settings.json
```

### Gopls still running old processes

```bash
# Force kill all
killall -9 gopls

# Wait 5 seconds
sleep 5

# Verify none remain
pgrep gopls
```

### IDE still hangs on tests

```bash
# Check gopls log for errors
tail -100 /tmp/gopls.log

# Reduce timeout and increase memory
export GOGC=50  # More aggressive garbage collection
export GOMAXPROCS=4  # Limit parallel processing

# Restart IDE with env vars
```

### Can't find test files

```bash
# Verify module is correct
cat go.mod

# Rebuild module cache
go mod verify
go mod tidy

# List tests to verify they exist
go test ./... -list . 2>&1 | head -20
```

## Performance Optimization

After cleanup, optimize your IDE setup:

### Reduce Analysis Scope

```json
// In .vscode/settings.json
"go.languageServerFlags": [
  "-logfile", "/tmp/gopls.log",
  "-rpc.trace"
],
"go.buildOnSave": "off",
"go.lintOnSave": "package",  // Not "workspace"
"go.vetOnSave": "package"     // Not "workspace"
```

### Optimize Test Discovery

```json
{
  "go.testExplorer.showOutput": true,
  "go.testExplorer.concatenateOutput": true,
  "go.testTimeout": "30s",
  "go.testOnSave": "off"
}
```

### Monitor Performance

```bash
# Watch gopls memory usage
while true; do
  echo "=== $(date) ==="
  ps aux | grep gopls | grep -v grep | awk '{print $2, $6, $11}'
  sleep 5
done
```

## Verification Checklist

After completing setup:

- [ ] Only one agg_go directory exists
- [ ] `pgrep gopls` shows 0-1 processes
- [ ] `go test -list ./... | sort | uniq -c | sort -rn` shows all tests appear exactly once
- [ ] VS Code test explorer loads without hanging
- [ ] Running a single test completes in <5 seconds
- [ ] `just quick` completes without errors
- [ ] `just test` runs all tests without hanging
- [ ] IDE autocomplete works smoothly
- [ ] Go to definition works correctly

## Prevention

To avoid this in the future:

1. **Use absolute paths** in IDE workspace configuration
2. **Document your primary workspace** in README
3. **Add workspace lockfile** to .gitignore to prevent accidental adds:
   ```bash
   echo ".vscode-server/" >> .gitignore
   echo ".vscode/" >> .gitignore
   ```
4. **Set workspace-level settings** instead of global gopls settings
5. **Monitor gopls processes** periodically for accumulation

## Additional Resources

- Gopls documentation: https://github.com/golang/tools/blob/master/gopls/README.md
- VS Code Go extension: https://github.com/golang/vscode-go
- Common Go IDE issues: https://github.com/golang/vscode-go/wiki/troubleshooting

---

If you continue to experience issues after following this guide, capture:

1. Output of `go list ./...`
2. Output of `go test -list ./... | sort | uniq -c`
3. Content of `/tmp/gopls.log` (last 50 lines)
4. Output of `pgrep -a gopls`
