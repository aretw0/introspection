# Module Rename: instrospection → introspection

This document describes the steps required after merging the module rename PR that fixes the typo `instrospection` → `introspection` in the Go module path.

## Context

The Go module was originally published as `github.com/aretw0/instrospection` (with an extra `r`). Since the GitHub repository has always been named `introspection`, the module path was inconsistent with the repository URL.

PR #6 fixes the typo in `go.mod`, all import paths, and documentation. However, because Go modules are immutable once published, additional manual steps are required after merging.

## Post-Merge Steps

### 1. Delete Old Tags and Releases

The existing tags (`v0.1.0`, `v0.1.1`) point to commits with the old (incorrect) module path. They must be removed and recreated so that `pkg.go.dev` and `go get` resolve the correct module.

```bash
# Delete remote tags
git push origin :refs/tags/v0.1.0
git push origin :refs/tags/v0.1.1

# Delete local tags
git tag -d v0.1.0
git tag -d v0.1.1
```

Then delete the corresponding GitHub Releases via the GitHub UI:

1. Go to <https://github.com/aretw0/introspection/releases>
2. Delete the `v0.1.0` and `v0.1.1` releases

### 2. Retract Old Versions (Optional but Recommended)

If any users already consumed the old module path, add a `retract` directive to `go.mod` to signal that the old versions should not be used:

```go
module github.com/aretw0/introspection

go 1.24.13

retract (
    v0.1.0 // Published with incorrect module path (instrospection)
    v0.1.1 // Published with incorrect module path (instrospection)
)
```

> **Note**: The `retract` directive only takes effect when a version **higher** than the retracted versions is published. It will be processed when `v0.1.2` (or later) is tagged.

### 3. Bump Version and Re-Tag

After the PR is merged into `main`, create a new release with the corrected module path:

```bash
# Ensure VERSION is updated
echo "0.1.2" > VERSION
git add VERSION
git commit -m "Bump version to 0.1.2"
git push origin main

# Create and push the new tag
git tag -a v0.1.2 -m "Release v0.1.2 - Fix module path typo"
git push origin v0.1.2
```

This will trigger GoReleaser via GitHub Actions to create the release automatically.

### 4. Verify pkg.go.dev

After the new tag is pushed:

1. Visit <https://pkg.go.dev/github.com/aretw0/introspection>
2. If the page doesn't appear automatically, request indexing:

   ```text
   https://pkg.go.dev/github.com/aretw0/introspection@v0.1.2
   ```

3. You can also trigger indexing via the Go module proxy:

   ```bash
   curl https://proxy.golang.org/github.com/aretw0/introspection/@v/v0.1.2.info
   ```

4. Verify the module documentation renders correctly

### 5. Verify Go Report Card

1. Visit <https://goreportcard.com/report/github.com/aretw0/introspection>
2. Trigger a new report if needed

### 6. Update Downstream Consumers

If any other projects (e.g., `lifecycle`) depend on this module, update their imports:

```bash
# In the downstream project
go get github.com/aretw0/introspection@v0.1.2
```

## Important Notes

### Why Not Rewrite Git History?

While the original issue mentions git history rewriting, this is **not recommended** for published Go modules because:

1. **Go module proxy caching**: The Go module proxy (`proxy.golang.org`) caches module versions permanently. Rewriting history doesn't change what's cached.
2. **Checksum database**: The Go checksum database (`sum.golang.org`) has recorded hashes for published versions. Changed hashes would cause verification failures.
3. **Force push risks**: Force-pushing to `main` can break other contributors' local clones and CI pipelines.

The **correct approach** for Go modules is to:

- Delete the old tags/releases
- Publish a new version with the fix
- Optionally retract the old versions

### Go Module Path vs Repository URL

In Go, the module path in `go.mod` must match the repository URL for the module proxy to resolve it correctly. The old path `github.com/aretw0/instrospection` could never be fetched via `go get` since the repository is `github.com/aretw0/introspection`.

## Checklist

- [ ] Delete old tags (`v0.1.0`, `v0.1.1`) from remote
- [ ] Delete old GitHub Releases
- [ ] Optionally add `retract` directives to `go.mod`
- [ ] Bump VERSION to `0.1.2`
- [ ] Create and push `v0.1.2` tag
- [ ] Verify GoReleaser creates the release
- [ ] Verify pkg.go.dev indexes the correct module
- [ ] Verify Go Report Card
- [ ] Update downstream consumers (if any)
