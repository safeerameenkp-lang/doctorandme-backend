# Shared Security Module Setup Guide

This guide explains how to set up the `shared-security` module as a separate Go module that can be used across all microservices.

## 📦 Module Structure

```
drandme-shared-security/
├── go.mod
├── go.sum
├── middleware.go
├── errors.go
└── README.md
```

## 🔧 Step 1: Create the Module Repository

1. **Create a new Git repository**:
   ```bash
   mkdir drandme-shared-security
   cd drandme-shared-security
   git init
   ```

2. **Copy the shared security files**:
   ```bash
   cp -r ../drandme-backend/shared/security/* .
   ```

3. **Update go.mod**:
   ```go
   module github.com/yourorg/drandme-shared-security
   
   go 1.21
   
   require (
       github.com/gin-gonic/gin v1.9.1
       github.com/golang-jwt/jwt/v5 v5.0.0
       github.com/lib/pq v1.10.9
   )
   ```

4. **Initialize Go module**:
   ```bash
   go mod tidy
   ```

## 🏷️ Step 2: Version and Tag

1. **Commit and push**:
   ```bash
   git add .
   git commit -m "Initial shared security module"
   git remote add origin https://github.com/yourorg/drandme-shared-security.git
   git push -u origin main
   ```

2. **Create version tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

## 📥 Step 3: Use in Services

### Option A: Public Repository

If the repository is public, services can import directly:

```go
// In service go.mod
require (
    github.com/yourorg/drandme-shared-security v1.0.0
)

// In service code
import "github.com/yourorg/drandme-shared-security"
```

### Option B: Private Repository

For private repositories, you need to configure Git credentials:

1. **Set up Git credentials**:
   ```bash
   git config --global url."https://username:token@github.com".insteadOf "https://github.com"
   ```

2. **Or use SSH**:
   ```bash
   git config --global url."git@github.com:".insteadOf "https://github.com/"
   ```

3. **In service go.mod, add replace directive** (for development):
   ```go
   require (
       github.com/yourorg/drandme-shared-security v1.0.0
   )
   
   replace github.com/yourorg/drandme-shared-security => github.com/yourorg/drandme-shared-security v1.0.0
   ```

### Option C: Go Workspace (Local Development)

For local development, you can use Go workspaces:

1. **Create workspace**:
   ```bash
   mkdir drandme-workspace
   cd drandme-workspace
   go work init
   ```

2. **Add modules**:
   ```bash
   go work use ../drandme-shared-security
   go work use ../drandme-auth-service
   go work use ../drandme-organization-service
   go work use ../drandme-appointment-service
   ```

3. **In service go.mod, use replace**:
   ```go
   replace github.com/yourorg/drandme-shared-security => ../drandme-shared-security
   ```

## 🔄 Step 4: Update Service Imports

In each service, update imports:

```go
// OLD (monorepo)
import "shared-security"

// NEW (separate repo)
import "github.com/yourorg/drandme-shared-security"
```

Update all files that use the shared module:
- `main.go`
- `routes/*.go`
- `controllers/*.go` (if they use security functions)

## 📝 Step 5: Update Service go.mod

Example for auth-service:

```go
module github.com/yourorg/drandme-auth-service

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/lib/pq v1.10.9
    golang.org/x/crypto v0.14.0
    github.com/yourorg/drandme-shared-security v1.0.0
)

// Remove this line (no longer needed):
// replace shared-security => ./shared-security
```

Then run:
```bash
go mod tidy
```

## 🚀 Step 6: Version Management

### Semantic Versioning

Use semantic versioning for the shared module:
- `v1.0.0` - Initial release
- `v1.0.1` - Bug fixes
- `v1.1.0` - New features (backward compatible)
- `v2.0.0` - Breaking changes

### Updating Services

When you update the shared module:

1. **Make changes** in `drandme-shared-security`
2. **Tag new version**:
   ```bash
   git tag v1.1.0
   git push origin v1.1.0
   ```
3. **Update services**:
   ```bash
   cd drandme-auth-service
   go get github.com/yourorg/drandme-shared-security@v1.1.0
   go mod tidy
   ```

## 🧪 Step 7: Testing

Test that services can import the module:

```bash
# In service directory
go mod download
go build
```

## 📋 Checklist

- [ ] Create `drandme-shared-security` repository
- [ ] Copy shared security files
- [ ] Update go.mod with proper module path
- [ ] Tag initial version (v1.0.0)
- [ ] Update all services to use new module path
- [ ] Remove local replace directives
- [ ] Test imports in all services
- [ ] Document versioning strategy
- [ ] Setup CI/CD for shared module

## 🔒 Security Considerations

1. **Keep secrets out of the shared module**: The module should only contain code, not secrets
2. **Use environment variables**: Services should pass secrets via environment variables
3. **Version pinning**: Always pin specific versions in production
4. **Private repositories**: Use private repositories for sensitive code

## 📚 Additional Resources

- [Go Modules Documentation](https://go.dev/ref/mod)
- [Go Workspaces](https://go.dev/doc/tutorial/workspaces)
- [Semantic Versioning](https://semver.org/)

