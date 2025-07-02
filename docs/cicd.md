# CI/CD and Automation

This document describes the Continuous Integration and Continuous Deployment (CI/CD) pipeline for the SpaceTraders MCP Server, including automated workflows, quality gates, and release processes.

## Overview

The SpaceTraders MCP Server uses GitHub Actions for CI/CD automation, providing:

- **Automated testing** on every pull request and commit
- **Multi-platform builds** for macOS, Windows, and Linux
- **Quality gates** to ensure code standards
- **Automated releases** with semantic versioning
- **Dependency management** and security scanning
- **Integration testing** with real SpaceTraders API

## Workflows

The CI/CD system consists of several automated workflows:

### 1. **Main CI Pipeline** (`.github/workflows/ci.yml`)

**Triggers:**
- Pull requests to main branch
- Commits to main branch
- Manual workflow dispatch

**Jobs:**
1. **Code Quality**
   - Go formatting check (`gofmt`)
   - Linting with `golangci-lint`
   - Security scanning with `gosec`
   - Dependency vulnerability check

2. **Testing**
   - Unit tests with coverage reporting
   - Race condition detection
   - Cross-platform compatibility tests

3. **Build Verification**
   - Build for multiple platforms (Linux, macOS, Windows)
   - Verify binary functionality
   - Check for build warnings

### 2. **Integration Tests** (`.github/workflows/integration.yml`)

**Triggers:**
- Scheduled runs (daily)
- Manual workflow dispatch
- Release preparation

**Requirements:**
- SpaceTraders API test credentials
- Live API connectivity

**Test Coverage:**
- Full MCP protocol compliance
- Real API interactions
- End-to-end workflows
- Performance benchmarks

### 3. **Release Automation** (`.github/workflows/release.yml`)

**Triggers:**
- Git tags matching `v*.*.*` pattern
- Manual release creation

**Process:**
1. Build binaries for all supported platforms
2. Run comprehensive test suite
3. Generate release notes from commits
4. Create GitHub release with artifacts
5. Update documentation

## Secrets Configuration

### Required Secrets

The following secrets must be configured in the GitHub repository:

```
SPACETRADERS_TOKEN          # SpaceTraders API token for testing
SPACETRADERS_AGENT_SYMBOL   # Test agent symbol
CODECOV_TOKEN               # Code coverage reporting (optional)
```

### Secret Management

**Setting secrets:**
1. Go to repository Settings → Secrets and variables → Actions
2. Add each required secret with appropriate values
3. Use environment-specific secrets for different branches if needed

**Security practices:**
- Use dedicated test accounts for CI/CD
- Rotate tokens regularly
- Limit token permissions to minimum required
- Monitor token usage in CI logs

## Automated Dependency Updates

### Dependabot Configuration

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    reviewers:
      - "maintainer-username"
    assignees:
      - "maintainer-username"
```

**Features:**
- Weekly dependency updates
- Automatic security updates
- Grouped updates for related packages
- Auto-merge for patch updates (optional)

### Dependency Scanning

- **Vulnerability scanning** with GitHub Security Advisories
- **License compliance** checking
- **Outdated dependency** reporting
- **Supply chain security** monitoring

## Quality Gates

### Code Quality Requirements

All changes must pass these quality gates:

1. **Formatting**: Code must pass `gofmt` formatting
2. **Linting**: No linting errors from `golangci-lint`
3. **Security**: No security issues from `gosec`
4. **Tests**: All tests must pass with ≥80% coverage
5. **Build**: Must build successfully on all target platforms

### Pull Request Checks

**Required status checks:**
- [ ] Code formatting
- [ ] Linting
- [ ] Security scan
- [ ] Unit tests
- [ ] Integration tests (for significant changes)
- [ ] Build verification

**Merge requirements:**
- All status checks must pass
- At least one approving review
- Branch must be up-to-date with main
- No merge conflicts

## Release Process

### Versioning Strategy

The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backward compatible)
- **PATCH** version: Bug fixes (backward compatible)

### Automated Release Steps

1. **Tag Creation**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **Build Process**
   - Cross-compile for all platforms
   - Run full test suite
   - Generate checksums for binaries

3. **Release Creation**
   - Create GitHub release
   - Upload platform-specific binaries
   - Generate release notes from commits
   - Update documentation links

4. **Post-Release**
   - Update installation instructions
   - Notify users of new version
   - Update example configurations

### Manual Release Process

For manual releases or hotfixes:

```bash
# Create release branch
git checkout -b release/v1.2.3

# Update version in code
# Make necessary changes
# Update CHANGELOG.md

# Create and push tag
git tag v1.2.3
git push origin v1.2.3

# GitHub Actions will handle the rest
```

## Development Workflow

### Branch Strategy

- **main**: Stable, production-ready code
- **develop**: Integration branch for features
- **feature/***: Individual feature branches
- **hotfix/***: Critical bug fixes
- **release/***: Release preparation branches

### Workflow Steps

1. **Feature Development**
   ```bash
   git checkout -b feature/new-feature
   # Develop feature
   # Add tests
   # Update documentation
   git push origin feature/new-feature
   # Create pull request
   ```

2. **Pull Request Process**
   - Create PR with descriptive title and description
   - Link related issues
   - Ensure all CI checks pass
   - Request review from maintainers
   - Address review feedback
   - Merge when approved

3. **Integration Testing**
   - Changes are tested on develop branch
   - Integration tests run automatically
   - Performance regression testing
   - Manual testing for complex changes

## Monitoring and Alerts

### CI/CD Monitoring

- **Build success/failure rates**
- **Test execution times**
- **Deployment frequency**
- **Mean time to recovery**

### Alert Configuration

**Failure notifications:**
- Failed builds on main branch
- Security vulnerabilities detected
- Performance regression alerts
- Dependency update failures

**Integration:**
- Slack notifications for critical failures
- Email alerts for maintainers
- GitHub issue creation for recurring failures

## Performance Optimization

### Build Optimization

- **Caching**: Go module cache, build cache
- **Parallelization**: Concurrent jobs where possible
- **Resource allocation**: Appropriate runner sizes
- **Artifact management**: Efficient storage and retrieval

### Test Optimization

- **Test parallelization**: Run tests concurrently
- **Smart test selection**: Run only affected tests for PRs
- **Test result caching**: Cache results for unchanged code
- **Flaky test detection**: Identify and fix unreliable tests

## Troubleshooting CI/CD Issues

### Common Problems

**Build Failures:**
- Check for dependency changes
- Verify environment consistency
- Look for platform-specific issues
- Check resource constraints

**Test Failures:**
- Review test logs for specific failures
- Check for race conditions
- Verify test data and mocks
- Look for environment-specific issues

**Deployment Issues:**
- Verify secrets and credentials
- Check API connectivity
- Review artifact generation
- Validate release process steps

### Debug Commands

```bash
# Local CI simulation
act -j ci

# Test specific workflow
act -j test

# Debug with verbose output
act -j ci --verbose
```

### Log Analysis

- **Structured logging**: Use consistent log format
- **Log aggregation**: Collect logs from all jobs
- **Error categorization**: Group similar errors
- **Trend analysis**: Track error patterns over time

## Best Practices

### CI/CD Pipeline Design

1. **Fast feedback**: Keep build times under 10 minutes
2. **Fail fast**: Run quick checks first
3. **Parallelization**: Run independent jobs concurrently
4. **Caching**: Cache dependencies and build artifacts
5. **Security**: Never log secrets or sensitive data

### Code Quality

1. **Automated checks**: Enforce standards automatically
2. **Consistent formatting**: Use automated formatters
3. **Security scanning**: Check for vulnerabilities
4. **Test coverage**: Maintain high coverage standards
5. **Documentation**: Keep docs up-to-date

### Release Management

1. **Semantic versioning**: Follow versioning standards
2. **Release notes**: Document changes clearly
3. **Backward compatibility**: Maintain API compatibility
4. **Rollback capability**: Ensure easy rollback process
5. **Gradual rollout**: Consider phased deployments

## Configuration Examples

### GitHub Actions Workflow

```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: make test-coverage
      env:
        SPACETRADERS_TOKEN: ${{ secrets.SPACETRADERS_TOKEN }}
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
```

### Makefile Automation

```makefile
.PHONY: ci
ci: fmt vet lint test build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test-coverage
test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
```
