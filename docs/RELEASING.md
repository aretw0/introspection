# Release Process

This document describes how to create releases for the introspection project.

## Overview

The project uses [GoReleaser](https://goreleaser.com/) to automate the release process. Releases are triggered automatically when a new tag following the `v*` pattern is pushed to the repository.

## Prerequisites

- Push access to the repository
- The main branch should be in a releasable state
- All tests should be passing

## Creating a Release

### 1. Update Version (if needed)

The VERSION file in the repository root contains the current version. Update it if needed:

```bash
echo "0.1.0" > VERSION
git add VERSION
git commit -m "Bump version to 0.1.0"
git push origin main
```

### 2. Create and Push a Tag

Create a new tag with the desired version number (following semantic versioning):

```bash
# Create an annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to GitHub
git push origin v0.1.0
```

### 3. Automated Release

Once the tag is pushed, GitHub Actions will automatically:

1. Checkout the code with full history
2. Set up Go
3. Run GoReleaser to create the release

The workflow will:
- Generate release notes from commits since the last tag
- Create a GitHub Release with the changelog
- Attach checksums for verification

### 4. Verify the Release

After the workflow completes:

1. Go to https://github.com/aretw0/introspection/releases
2. Verify the release was created successfully
3. Check the changelog for accuracy
4. Verify the release assets (checksums.txt)

## Release Workflow Details

The release workflow is defined in `.github/workflows/release.yml` and triggers on tags matching `v*`.

Key features:
- **Automated changelog**: Generated from commit messages
- **Semantic versioning**: Follows the `v*` tag pattern
- **GitHub integration**: Creates releases directly on GitHub

## GoReleaser Configuration

The GoReleaser configuration is in `.goreleaser.yaml`. Since introspection is a library (not a binary), the build is skipped, but GoReleaser still:
- Generates checksums
- Creates release notes
- Publishes to GitHub Releases

## Troubleshooting

### Release Failed

If the GitHub Actions workflow fails:

1. Check the workflow logs at https://github.com/aretw0/introspection/actions
2. Verify the tag format follows `v*` (e.g., `v0.1.0`, `v1.2.3`)
3. Ensure the repository has the necessary permissions

### Deleting a Failed Release

If you need to delete a tag and recreate it:

```bash
# Delete local tag
git tag -d v0.1.0

# Delete remote tag
git push origin :refs/tags/v0.1.0

# Recreate and push
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

## Version Scheme

This project follows [Semantic Versioning](https://semver.org/):

- **Major version** (v1.0.0): Breaking changes
- **Minor version** (v0.1.0): New features, backward compatible
- **Patch version** (v0.0.1): Bug fixes, backward compatible

## Reference

- [GoReleaser Documentation](https://goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)
